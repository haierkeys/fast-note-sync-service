package mcp_router

import (
	"context"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/gookit/goutil/dump"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

type MCPHandler struct {
	mcpServer       *mcpserver.MCPServer
	sseServer       *mcpserver.SSEServer
	ssePingInterval time.Duration // SSE heartbeat interval / SSE 心跳间隔
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
			return ctx
		}))

	return &MCPHandler{
		mcpServer:       srv,
		sseServer:       sseSrv,
		ssePingInterval: pingInterval,
	}
}

func (h *MCPHandler) HandleSSE(c *gin.Context) {
	uid := pkgapp.GetUID(c)
	ctx := context.WithValue(c.Request.Context(), "uid", uid)
	if vaultName := c.GetHeader("X-Default-Vault-Name"); vaultName != "" {
		ctx = context.WithValue(ctx, "default_vault_name", vaultName)
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

	// Let SSEServer handle the SSE connection
	h.sseServer.SSEHandler().ServeHTTP(c.Writer, c.Request.WithContext(ctx))
}

func (h *MCPHandler) HandleMessage(c *gin.Context) {
	// Let SSEServer handle the message
	h.sseServer.MessageHandler().ServeHTTP(c.Writer, c.Request)
}
