package middleware

import (
	"embed"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// StaticCompressMiddleware returns a middleware that supports pre-compressed files (.br, .gz).
// StaticCompressMiddleware 返回一个支持预压缩文件（.br, .gz）的中间件。
// It only handles requests for the frontend directory.
// 它仅处理对 frontend 目录的请求。
func StaticCompressMiddleware(frontendFiles embed.FS) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Only handle /assets/ and /static/ paths which are served from frontendFiles
		// 仅处理从 frontendFiles 提供的 /assets/ 和 /static/ 路径
		var internalPath string
		if strings.HasPrefix(path, "/assets/") {
			internalPath = "frontend/assets" + strings.TrimPrefix(path, "/assets")
		} else if strings.HasPrefix(path, "/static/") {
			internalPath = "frontend/static" + strings.TrimPrefix(path, "/static")
		} else {
			c.Next()
			return
		}

		acceptEncoding := c.GetHeader("Accept-Encoding")

		// Priority: Brotli (.br) > Gzip (.gz)
		// 优先级：Brotli (.br) > Gzip (.gz)

		if strings.Contains(acceptEncoding, "br") {
			brPath := internalPath + ".br"
			if _, err := frontendFiles.Open(brPath); err == nil {
				c.Header("Content-Encoding", "br")
				c.Header("Vary", "Accept-Encoding")
				// Set Content-Type based on the original file extension
				// 根据原始文件后缀设置 Content-Type
				setContentType(c, internalPath)
				c.FileFromFS(brPath, http.FS(frontendFiles))
				c.Abort()
				return
			}
		}

		if strings.Contains(acceptEncoding, "gzip") {
			gzPath := internalPath + ".gz"
			if _, err := frontendFiles.Open(gzPath); err == nil {
				c.Header("Content-Encoding", "gzip")
				c.Header("Vary", "Accept-Encoding")
				setContentType(c, internalPath)
				c.FileFromFS(gzPath, http.FS(frontendFiles))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func setContentType(c *gin.Context, path string) {
	ext := filepath.Ext(path)
	switch ext {
	case ".js":
		c.Header("Content-Type", "application/javascript")
	case ".css":
		c.Header("Content-Type", "text/css")
	case ".html":
		c.Header("Content-Type", "text/html; charset=utf-8")
	case ".json":
		c.Header("Content-Type", "application/json; charset=utf-8")
	case ".svg":
		c.Header("Content-Type", "image/svg+xml")
	case ".png":
		c.Header("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		c.Header("Content-Type", "image/jpeg")
	case ".gif":
		c.Header("Content-Type", "image/gif")
	case ".woff":
		c.Header("Content-Type", "font/woff")
	case ".woff2":
		c.Header("Content-Type", "font/woff2")
	case ".ttf":
		c.Header("Content-Type", "font/ttf")
	}
}
