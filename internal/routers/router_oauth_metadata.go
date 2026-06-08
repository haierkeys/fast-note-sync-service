package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/config"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

const protectedResourceMetadataAPIPath = "/.well-known/oauth-protected-resource/api/mcp"

func registerOAuthMetadataRoutes(r *gin.Engine, appContainer *app.App) {
	registerOAuthMetadataRoutesWithConfig(r, appContainer.Config().OAuth)
}

func registerOAuthMetadataRoutesWithConfig(r *gin.Engine, cfg config.OAuthConfig) {
	if !cfg.Enabled {
		return
	}

	handler := mcpserver.NewProtectedResourceMetadataHandler(cfg.ProtectedResourceMetadata())
	ginHandler := func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}

	r.Any(mcpserver.WellKnownProtectedResourcePath, ginHandler)
	r.Any(protectedResourceMetadataAPIPath, ginHandler)

	if metadataPath := mcpserver.ProtectedResourceMetadataPath(cfg.Resource); metadataPath != "" &&
		metadataPath != mcpserver.WellKnownProtectedResourcePath &&
		metadataPath != protectedResourceMetadataAPIPath {
		r.Any(metadataPath, ginHandler)
	}
}
