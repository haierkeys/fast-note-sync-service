package routers

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
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

	if *cfg.Storage.LocalFS.HttpfsIsEnable && cfg.Storage.LocalFS.IsEnabled {
		r.StaticFS(cfg.Storage.LocalFS.SavePath, http.Dir(cfg.Storage.LocalFS.SavePath))
		r.OPTIONS(cfg.Storage.LocalFS.SavePath+"/*filepath", func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		})
	}

	// 附件模拟静态访问入口，主要提供给 nginx 访问
	// Attachment mock static access endpoint, mainly used for nginx proxy
	r.GET("/attachments/:uid/:vault/*filepath", func(c *gin.Context) {
		// 1. 强制启用校验
		// 1. Mandatory enable check
		if !cfg.AttachmentStatic.IsEnable {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// 2. 强制白名单校验
		// 2. Mandatory whitelist checks
		if len(cfg.AttachmentStatic.AllowedVaults) == 0 || len(cfg.AttachmentStatic.AllowedTypes) == 0 {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		uidStr := c.Param("uid")
		uid, err := strconv.ParseInt(uidStr, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		vaultRaw := c.Param("vault")
		filepathRaw := c.Param("filepath")

		// URL 解码支持中文名称
		// URL decode to support Chinese names
		vault, err := url.PathUnescape(vaultRaw)
		if err != nil {
			vault = vaultRaw
		}
		filepathParam, err := url.PathUnescape(filepathRaw)
		if err != nil {
			filepathParam = filepathRaw
		}

		filepathParam = strings.TrimPrefix(filepathParam, "/")
		if vault == "" || filepathParam == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// 3. 校验用户和库白名单
		// 3. Validate user and vault whitelist
		vaults, ok := cfg.AttachmentStatic.AllowedVaults[uidStr]
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		vaultAllowed := false
		for _, v := range vaults {
			if v == vault {
				vaultAllowed = true
				break
			}
		}
		if !vaultAllowed {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// 4. 校验允许访问的文件后缀类型
		// 4. Validate allowed file extensions
		ext := filepath.Ext(filepathParam)
		ext = strings.ToLower(ext)
		typeAllowed := false
		for _, t := range cfg.AttachmentStatic.AllowedTypes {
			if strings.ToLower(t) == ext {
				typeAllowed = true
				break
			}
		}
		if !typeAllowed {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// 5. 调用文件服务获取物理路径并返回
		// 5. Call file service to get physical path and serve
		fileSvc := appContainer.GetFileService("", "", "")
		ctx := c.Request.Context()
		params := &dto.FileGetRequest{
			Vault:    vault,
			Path:     filepathParam,
			PathHash: util.EncodeHash32(filepathParam),
		}

		savePath, contentType, mtime, etag, fileName, err := fileSvc.GetContentInfo(ctx, uid, params)
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		file, err := os.Open(savePath)
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		defer file.Close()

		if contentType != "" {
			c.Header("Content-Type", contentType)
		}
		// 设置强缓存，缓存一年
		// Set strong cache for one year
		c.Header("Cache-Control", "public, s-maxage=31536000, max-age=31536000, must-revalidate")
		if etag != "" {
			c.Header("ETag", etag)
		}

		http.ServeContent(c.Writer, c.Request, fileName, time.UnixMilli(mtime), file)
	})
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
	var script string
	if apiUrl == "" {
		script = "<script>localStorage.removeItem('API_URL');</script>"
	} else {
		// Inject localStorage setter before </body>
		// 在 </body> 前注入 localStorage 设置脚本
		script = fmt.Sprintf("<script>localStorage.setItem('API_URL', '%s');</script>", apiUrl)
	}

	html := string(content)
	if strings.Contains(html, "</body>") {
		html = strings.Replace(html, "</body>", script+"</body>", 1)
	} else {
		html += script
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
