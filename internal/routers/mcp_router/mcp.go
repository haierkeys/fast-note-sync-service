package mcp_router

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

type MCPHandler struct {
	mcpServer *mcpserver.MCPServer
	sseServer *mcpserver.SSEServer
}

func NewMCPHandler(appContainer *app.App, wss *pkgapp.WebsocketServer) *MCPHandler {
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
		mcpServer: srv,
		sseServer: sseSrv,
	}
}

func (h *MCPHandler) HandleSSE(c *gin.Context) {
	uid := pkgapp.GetUID(c)
	ctx := context.WithValue(c.Request.Context(), "uid", uid)
	if vaultName := c.GetHeader("X-Default-Vault-Name"); vaultName != "" {
		ctx = context.WithValue(ctx, "default_vault_name", vaultName)
	}
	// Let SSEServer handle the SSE connection
	h.sseServer.SSEHandler().ServeHTTP(c.Writer, c.Request.WithContext(ctx))
}

func (h *MCPHandler) HandleMessage(c *gin.Context) {
	// Let SSEServer handle the message
	h.sseServer.MessageHandler().ServeHTTP(c.Writer, c.Request)
}
