package mcp_router

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/gookit/goutil/dump"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// endpointRewriter is an http.ResponseWriter wrapper that rewrites the
// SSE endpoint event from a relative path to an absolute URL.
// This fixes compatibility with MCP clients (e.g., Hermes/Anthropic Python SDK)
// that cannot resolve relative endpoint paths returned by mark3labs/mcp-go SSEServer.
type endpointRewriter struct {
	http.ResponseWriter
	absoluteBase string // e.g. "http://192.168.1.89:9000"
	endpointDone bool
}

func (w *endpointRewriter) Write(data []byte) (int, error) {
	if !w.endpointDone && bytes.Contains(data, []byte("event: endpoint")) {
		w.endpointDone = true
		// Replace relative path with absolute URL in the endpoint event
		data = []byte(strings.Replace(
			string(data),
			"/api/mcp/message?",
			w.absoluteBase+"/api/mcp/message?",
			1,
		))
	}
	return w.ResponseWriter.Write(data)
}

type MCPHandler struct {
	mcpServer        *mcpserver.MCPServer
	sseServer        *mcpserver.SSEServer
	streamableServer *mcpserver.StreamableHTTPServer // StreamableHTTP transport server / StreamableHTTP 传输协议服务
	ssePingInterval  time.Duration                   // SSE heartbeat interval / SSE 心跳间隔
}

func NewMCPHandler(appContainer *app.App, wss *pkgapp.WebsocketServer) *MCPHandler {
	cfg := appContainer.Config()
	pingInterval := time.Duration(cfg.Server.MCPSSEPingInterval) * time.Second
	if pingInterval <= 0 {
		pingInterval = 30 * time.Second // fallback default
	}

	srv := NewMCPServer(appContainer, wss)

	sseSrv := mcpserver.NewSSEServer(srv,
		mcpserver.WithMessageEndpoint("/api/mcp/message"),
		mcpserver.WithKeepAlive(true),
		mcpserver.WithKeepAliveInterval(pingInterval),
		mcpserver.WithSSEContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			if val := r.Context().Value("uid"); val != nil {
				ctx = context.WithValue(ctx, "uid", val)
			}
			if vaultName := r.Header.Get("X-Default-Vault-Name"); vaultName != "" {
				ctx = context.WithValue(ctx, "default_vault_name", vaultName)
			}

			// Extract client info
			if clientType := r.Header.Get("X-Client"); clientType != "" {
				ctx = context.WithValue(ctx, "client_type", clientType)
			}
			clientName := r.Header.Get("X-Client-Name")
			if clientName == "" {
				clientName = "MCP"
			} else {
				if decoded, err := url.QueryUnescape(clientName); err == nil {
					clientName = decoded
				}
				clientName = "MCP " + clientName
			}
			ctx = context.WithValue(ctx, "client_name", clientName)
			if clientVersion := r.Header.Get("X-Client-Version"); clientVersion != "" {
				ctx = context.WithValue(ctx, "client_version", clientVersion)
			}
			return ctx
		}))

	// StreamableHTTP server shares the same MCPServer instance as SSEServer.
	// StreamableHTTP 服务与 SSEServer 共享同一 MCPServer 实例。
	streamableSrv := mcpserver.NewStreamableHTTPServer(srv,
		mcpserver.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			// uid is pre-injected into the request context by HandleStreamableHTTP
			// before calling ServeHTTP, so we forward it here.
			// uid 已由 HandleStreamableHTTP 在调用 ServeHTTP 前注入请求上下文，此处直接透传。
			if val := r.Context().Value("uid"); val != nil {
				ctx = context.WithValue(ctx, "uid", val)
			}
			if vaultName := r.Header.Get("X-Default-Vault-Name"); vaultName != "" {
				ctx = context.WithValue(ctx, "default_vault_name", vaultName)
			}

			// Extract client info / 提取客户端信息
			if clientType := r.Header.Get("X-Client"); clientType != "" {
				ctx = context.WithValue(ctx, "client_type", clientType)
			}
			clientName := r.Header.Get("X-Client-Name")
			if clientName == "" {
				clientName = "MCP"
			} else {
				if decoded, err := url.QueryUnescape(clientName); err == nil {
					clientName = decoded
				}
				clientName = "MCP " + clientName
			}
			ctx = context.WithValue(ctx, "client_name", clientName)
			if clientVersion := r.Header.Get("X-Client-Version"); clientVersion != "" {
				ctx = context.WithValue(ctx, "client_version", clientVersion)
			}
			return ctx
		}),
	)

	return &MCPHandler{
		mcpServer:        srv,
		sseServer:        sseSrv,
		streamableServer: streamableSrv,
		ssePingInterval:  pingInterval,
	}
}

func (h *MCPHandler) HandleSSE(c *gin.Context) {
	uid := pkgapp.GetUID(c)
	ctx := context.WithValue(c.Request.Context(), "uid", uid)
	if vaultName := c.GetHeader("X-Default-Vault-Name"); vaultName != "" {
		ctx = context.WithValue(ctx, "default_vault_name", vaultName)
	}

	// Extract client info
	if clientType := c.GetHeader("X-Client"); clientType != "" {
		ctx = context.WithValue(ctx, "client_type", clientType)
	}
	if clientName := c.GetHeader("X-Client-Name"); clientName != "" {
		if decoded, err := url.QueryUnescape(clientName); err == nil {
			clientName = decoded
		}
		ctx = context.WithValue(ctx, "client_name", clientName)
	}
	if clientVersion := c.GetHeader("X-Client-Version"); clientVersion != "" {
		ctx = context.WithValue(ctx, "client_version", clientVersion)
	}

	// Set SSE headers
	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Proxy-Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable proxy buffering / 禁用代理缓冲

	// Flush headers immediately
	// 立即发送响应头
	c.Writer.Flush()

	// If it's a HEAD request, we've sent the headers, so we can return
	// 如果是 HEAD 请求，我们已经发送了响应头，可以直接返回
	if c.Request.Method == http.MethodHead {
		return
	}

	// Build absolute URL from the incoming request to fix MCP clients
	// (e.g., Hermes/Anthropic Python SDK) that cannot resolve relative
	// endpoint paths returned by mark3labs/mcp-go SSEServer.
	// See: https://github.com/haierkeys/fast-note-sync-service/issues/258
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	absoluteBase := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	// Let SSEServer handle the SSE connection, with endpoint URL rewriting
	rewriter := &endpointRewriter{
		ResponseWriter: c.Writer,
		absoluteBase:   absoluteBase,
	}
	h.sseServer.SSEHandler().ServeHTTP(rewriter, c.Request.WithContext(ctx))
}

func (h *MCPHandler) HandleMessage(c *gin.Context) {
	// Let SSEServer handle the message
	h.sseServer.MessageHandler().ServeHTTP(c.Writer, c.Request)
}

// HandleStreamableHTTP handles the MCP StreamableHTTP transport protocol.
// It accepts POST (request/notification), GET (SSE listening), and DELETE (session termination).
// HandleStreamableHTTP 处理 MCP StreamableHTTP 传输协议。
// 支持 POST（请求/通知）、GET（SSE 监听）和 DELETE（终止会话）。
func (h *MCPHandler) HandleStreamableHTTP(c *gin.Context) {
	uid := pkgapp.GetUID(c)
	// Pre-inject uid into the request context so that WithHTTPContextFunc can forward it.
	// 将 uid 预注入请求 context，以便 WithHTTPContextFunc 能够透传。
	ctx := context.WithValue(c.Request.Context(), "uid", uid)
	h.streamableServer.ServeHTTP(c.Writer, c.Request.WithContext(ctx))
}
