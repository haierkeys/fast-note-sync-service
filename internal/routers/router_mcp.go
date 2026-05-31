package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	"github.com/haierkeys/fast-note-sync-service/internal/routers/mcp_router"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
)

func registerMCPRoutes(api *gin.RouterGroup, appContainer *app.App, wss *pkgapp.WebsocketServer) {
	cfg := appContainer.Config()
	// MCP routes (No Timeout)
	// MCP 路由 (无超时限制)
	mcpHandler := mcp_router.NewMCPHandler(appContainer, wss)
	mcpGroup := api.Group("/mcp")
	mcpGroup.Use(middleware.UserAuthTokenWithConfig(cfg.Security.AuthTokenKey, appContainer.TokenService))
	{
		// Legacy SSE transport (backward compatible) / 旧版 SSE 传输（向后兼容）
		mcpGroup.Match([]string{http.MethodGet, http.MethodHead}, "/sse", mcpHandler.HandleSSE)
		mcpGroup.POST("/message", mcpHandler.HandleMessage)

		// StreamableHTTP transport: POST (request), GET (listen), DELETE (terminate session)
		// StreamableHTTP 传输: POST（请求）、GET（监听通知）、DELETE（终止会话）
		mcpGroup.Any("", mcpHandler.HandleStreamableHTTP)
	}
}
