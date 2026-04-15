package mcp_router

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

	sseSrv := mcpserver.NewSSEServer(srv, mcpserver.WithMessageEndpoint("/api/mcp/message"), mcpserver.WithSSEContextFunc(func(ctx context.Context, r *http.Request) context.Context {
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

	// 创建可取消的 context，用于控制心跳 goroutine
	// Create cancellable context for heartbeat goroutine lifecycle
	heartbeatCtx, cancelHeartbeat := context.WithCancel(ctx)
	defer cancelHeartbeat()

	// 启动心跳 goroutine，每 30 秒发送 SSE 注释行保持连接活跃
	// Start heartbeat goroutine to send SSE comment every 30s to keep connection alive
	go func() {
		ticker := time.NewTicker(h.ssePingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-heartbeatCtx.Done():
				return
			case <-ticker.C:
				// SSE comment line, MCP clients ignore this
				// SSE 注释行，MCP 客户端忽略此内容
				_, _ = fmt.Fprint(c.Writer, ":\n\n")
				c.Writer.Flush()
			}
		}
	}()

	// Let SSEServer handle the SSE connection
	h.sseServer.SSEHandler().ServeHTTP(c.Writer, c.Request.WithContext(ctx))
}

func (h *MCPHandler) HandleMessage(c *gin.Context) {
	// Let SSEServer handle the message
	h.sseServer.MessageHandler().ServeHTTP(c.Writer, c.Request)
}
