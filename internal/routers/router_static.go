package routers

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
)

func registerStaticFiles(r *gin.Engine, frontendFiles embed.FS, appContainer *app.App) {
	cfg := appContainer.Config()
	frontendAssets, _ := fs.Sub(frontendFiles, "frontend/assets")
	frontendStatic, _ := fs.Sub(frontendFiles, "frontend/static")

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

func registerWebGuiRoutes(r *gin.Engine, frontendFiles embed.FS, appContainer *app.App) {
	cfg := appContainer.Config()
	frontendIndexContent, _ := frontendFiles.ReadFile("frontend/index.html")
	apiUrl := cfg.Server.ExtApiUrl

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/webgui")
	})

	r.GET("/webgui/", func(c *gin.Context) {
		renderHTMLWithAPI(c, frontendIndexContent, apiUrl)
	})
}

func registerShareRoutes(r *gin.Engine, frontendFiles embed.FS, appContainer *app.App) {
	cfg := appContainer.Config()
	frontendShareContent, _ := frontendFiles.ReadFile("frontend/share.html")
	apiUrl := cfg.Server.ExtApiUrl

	r.GET("/share/:side/:token", func(c *gin.Context) {
		renderHTMLWithAPI(c, frontendShareContent, apiUrl)
	})
}

// renderHTMLWithAPI injects API_URL into HTML
// renderHTMLWithAPI 将 API_URL 注入到 HTML 中
func renderHTMLWithAPI(c *gin.Context, content []byte, apiUrl string) {
	if apiUrl == "" {
		c.Data(http.StatusOK, "text/html; charset=utf-8", content)
		return
	}

	// Inject localStorage setter before </body>
	// 在 </body> 前注入 localStorage 设置脚本
	script := fmt.Sprintf("<script>localStorage.setItem('API_URL', '%s');</script>", apiUrl)
	html := string(content)
	if strings.Contains(html, "</body>") {
		html = strings.Replace(html, "</body>", script+"</body>", 1)
	} else {
		html += script
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
