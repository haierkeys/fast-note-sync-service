package routers

import (
	"embed"
	"io/fs"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
)

func registerStaticRoutes(r *gin.Engine, frontendFiles embed.FS, appContainer *app.App) {
	cfg := appContainer.Config()
	frontendAssets, _ := fs.Sub(frontendFiles, "frontend/assets")
	frontendStatic, _ := fs.Sub(frontendFiles, "frontend/static")
	frontendIndexContent, _ := frontendFiles.ReadFile("frontend/index.html")
	frontendShareContent, _ := frontendFiles.ReadFile("frontend/share.html")

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/webgui")
	})
	r.GET("/webgui/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", frontendIndexContent)
	})

	r.GET("/share/:side/:token", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", frontendShareContent)
	})

	userStaticPath := "storage/user_static"
	if _, err := os.Stat(userStaticPath); os.IsNotExist(err) {
		_ = os.MkdirAll(userStaticPath, os.ModePerm)
	}

	cacheMiddleware := func(c *gin.Context) {
		// Set strong cache, cache for one year
		// 设置强缓存，缓存一年
		c.Header("Cache-Control", "public, s-maxage=31536000, max-age=31536000, must-revalidate")
		c.Next()
	}

	r.Group("/assets", cacheMiddleware, middleware.StaticCompressMiddleware(frontendFiles)).StaticFS("/", http.FS(frontendAssets))
	r.Group("/static", cacheMiddleware, middleware.StaticCompressMiddleware(frontendFiles)).StaticFS("/", http.FS(frontendStatic))
	r.Group("/user_static", cacheMiddleware).Static("/", userStaticPath)

	if cfg.Storage.LocalFS.HttpfsIsEnable && cfg.Storage.LocalFS.IsEnabled {
		r.StaticFS(cfg.Storage.LocalFS.SavePath, http.Dir(cfg.Storage.LocalFS.SavePath))
		r.OPTIONS(cfg.Storage.LocalFS.SavePath+"/*filepath", func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		})
	}
}
