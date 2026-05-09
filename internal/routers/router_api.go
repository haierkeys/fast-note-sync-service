package routers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	"github.com/haierkeys/fast-note-sync-service/internal/routers/api_router"
	"github.com/haierkeys/fast-note-sync-service/internal/routers/mcp_router"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

func registerAPIRoutes(r *gin.Engine, appContainer *app.App, wss *pkgapp.WebsocketServer, uni *ut.UniversalTranslator) {
	cfg := appContainer.Config()
	api := r.Group("/api")
	{
		api.Use(middleware.AppInfoWithConfig(app.Name, appContainer.Version().Version))
		api.Use(gin.Logger())
		api.Use(middleware.TraceMiddlewareWithConfig(cfg.Tracer.Enabled, cfg.Tracer.Header)) // Trace ID middleware
		// Trace ID 中间件
		api.Use(middleware.RateLimiter(methodLimiters))

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

		api.Use(middleware.ContextTimeout(time.Duration(cfg.App.DefaultContextTimeout) * time.Second))
		api.Use(middleware.LangWithTranslator(uni))
		api.Use(middleware.AccessLogWithLogger(appContainer.Logger()))
		api.Use(middleware.RecoveryWithLogger(appContainer.Logger()))

		// Create Handlers (injected App Container)
		// 创建 Handlers（注入 App Container）
		userHandler := api_router.NewUserHandler(appContainer)
		vaultHandler := api_router.NewVaultHandler(appContainer)
		noteHandler := api_router.NewNoteHandler(appContainer, wss)
		folderHandler := api_router.NewFolderHandler(appContainer)
		fileHandler := api_router.NewFileHandler(appContainer, wss)
		noteHistoryHandler := api_router.NewNoteHistoryHandler(appContainer, wss)
		versionHandler := api_router.NewVersionHandler(appContainer)
		adminControlHandler := api_router.NewAdminControlHandler(appContainer, wss)
		shareHandler := api_router.NewShareHandler(appContainer, wss)
		storageHandler := api_router.NewStorageHandler(appContainer)
		backupHandler := api_router.NewBackupHandler(appContainer)
		gitSyncHandler := api_router.NewGitSyncHandler(appContainer)
		settingHandler := api_router.NewSettingHandler(appContainer, wss)
		syncLogHandler := api_router.NewSyncLogHandler(appContainer)
		tokenHandler := api_router.NewTokenHandler(appContainer)

		api.POST("/user/register", userHandler.Register)
		api.POST("/user/login", userHandler.Login)
		api.GET("/user/sync", wss.Run())

		// Add server version interface (no auth required)
		// 添加服务端版本号接口（无需认证）
		api.GET("/version", versionHandler.ServerVersion)
		api.GET("/support", versionHandler.Support)
		api.GET("/webgui/config", adminControlHandler.Config)

		// Health check interface (no auth required)
		// 健康检查接口（无需认证）
		healthHandler := api_router.NewHealthHandler(appContainer)
		api.GET("/health", healthHandler.Check)

		// Share routing group (controlled read-only access)
		// 分享路由组 (受控的只读访问)
		share := api.Group("/share")
		share.Use(middleware.ShareAuthToken(appContainer.ShareService))
		{
			share.GET("/note", shareHandler.NoteGet) // Get shared note
			// 获取分享的笔记
			share.GET("/file", shareHandler.FileGet) // Get shared file content
			// 获取分享的文件内容
		}

		// Auth routing group (authentication required)
		// 需要认证的路由组
		auth := api.Group("/")
		auth.Use(middleware.UserAuthTokenWithConfig(cfg.Security.AuthTokenKey, appContainer.TokenService))
		{
			// Create share
			// 创建分享
			auth.POST("/share", shareHandler.Create)
			auth.POST("/share/password", shareHandler.UpdatePassword)
			auth.GET("/share", shareHandler.Query)
			auth.DELETE("/share", shareHandler.Cancel)
			auth.POST("/share/short_link", shareHandler.CreateShortLink)
			auth.GET("/shares", shareHandler.List)

			// Admin config interface
			// 管理员配置接口
			auth.GET("/admin/check", adminControlHandler.CheckAdmin)
			auth.GET("/admin/config", adminControlHandler.GetConfig)
			auth.POST("/admin/config", adminControlHandler.UpdateConfig)
			auth.GET("/admin/config/user_database", adminControlHandler.GetUserDatabaseConfig)
			auth.POST("/admin/config/user_database", adminControlHandler.UpdateUserDatabaseConfig)
			auth.POST("/admin/config/user_database/test", adminControlHandler.ValidateUserDatabaseConfig)
			auth.GET("/admin/config/ngrok", adminControlHandler.GetNgrokConfig)
			auth.POST("/admin/config/ngrok", adminControlHandler.UpdateNgrokConfig)
			auth.GET("/admin/config/cloudflare", adminControlHandler.GetCloudflareConfig)
			auth.POST("/admin/config/cloudflare", adminControlHandler.UpdateCloudflareConfig)
			auth.GET("/admin/systeminfo", adminControlHandler.GetSystemInfo)
			auth.GET("/admin/upgrade", adminControlHandler.Upgrade)
			auth.GET("/admin/restart", adminControlHandler.Restart)
			auth.GET("/admin/gc", adminControlHandler.GC)
			auth.GET("/admin/ws_clients", adminControlHandler.GetWSClients)
			auth.GET("/admin/cloudflared_tunnel_download", adminControlHandler.CloudflaredTunnelDownload)

			auth.POST("/user/change_password", userHandler.UserChangePassword)
			auth.GET("/user/info", userHandler.UserInfo)
			auth.GET("/vault", vaultHandler.List)
			auth.POST("/vault", vaultHandler.CreateOrUpdate)
			auth.DELETE("/vault", vaultHandler.Delete)

			auth.GET("/note", noteHandler.Get)
			auth.POST("/note", noteHandler.CreateOrUpdate)
			auth.DELETE("/note", noteHandler.Delete)
			auth.PUT("/note/restore", noteHandler.Restore)
			auth.POST("/note/rename", noteHandler.Rename)
			auth.GET("/notes", noteHandler.List)
			auth.DELETE("/note/recycle-clear", noteHandler.RecycleClear)
			auth.GET("/notes/share-paths", shareHandler.NoteSharePaths)

			auth.GET("/folder", folderHandler.Get)
			auth.POST("/folder", folderHandler.Create)
			auth.DELETE("/folder", folderHandler.Delete)
			auth.GET("/folders", folderHandler.List)
			auth.GET("/folder/notes", folderHandler.ListNotes)
			auth.GET("/folder/files", folderHandler.ListFiles)
			auth.GET("/folder/tree", folderHandler.Tree)

			// Note edit operations
			auth.PATCH("/note/frontmatter", noteHandler.PatchFrontmatter)
			auth.POST("/note/append", noteHandler.Append)
			auth.POST("/note/prepend", noteHandler.Prepend)
			auth.POST("/note/replace", noteHandler.Replace)

			// Note link operations
			auth.GET("/note/backlinks", noteHandler.GetBacklinks)
			auth.GET("/note/outlinks", noteHandler.GetOutlinks)

			auth.GET("/file", fileHandler.GetInfo)
			auth.POST("/file", fileHandler.Upload)
			auth.OPTIONS("/file", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			auth.GET("/file/info", fileHandler.Get)
			auth.OPTIONS("/file/info", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			auth.DELETE("/file", fileHandler.Delete)
			auth.PUT("/file/restore", fileHandler.Restore)
			auth.POST("/file/rename", fileHandler.Rename)
			auth.GET("/files", fileHandler.List)
			auth.DELETE("/file/recycle-clear", fileHandler.RecycleClear)
			auth.OPTIONS("/files", func(c *gin.Context) { c.Status(http.StatusNoContent) })

			auth.GET("/note/history", noteHistoryHandler.Get)
			auth.GET("/note/histories", noteHistoryHandler.List)
			auth.PUT("/note/history/restore", noteHistoryHandler.Restore)

			auth.GET("/storage", storageHandler.List)
			auth.POST("/storage", storageHandler.CreateOrUpdate)
			auth.GET("/storage/enabled_types", storageHandler.EnabledTypes)
			auth.POST("/storage/validate", storageHandler.Validate)
			auth.DELETE("/storage", storageHandler.Delete)

			auth.GET("/backup/configs", backupHandler.GetConfigs)
			auth.POST("/backup/config", backupHandler.UpdateConfig)
			auth.DELETE("/backup/config", backupHandler.DeleteConfig)
			auth.GET("/backup/historys", backupHandler.ListHistory)
			auth.POST("/backup/execute", backupHandler.Execute)

			auth.GET("/git-sync/configs", gitSyncHandler.GetConfigs)
			auth.POST("/git-sync/config", gitSyncHandler.UpdateConfig)
			auth.DELETE("/git-sync/config", gitSyncHandler.DeleteConfig)
			auth.POST("/git-sync/validate", gitSyncHandler.Validate)
			auth.DELETE("/git-sync/config/clean", gitSyncHandler.CleanWorkspace)
			auth.POST("/git-sync/config/execute", gitSyncHandler.Execute)
			auth.GET("/git-sync/histories", gitSyncHandler.GetHistories)

			auth.GET("/setting", settingHandler.Get)
			auth.POST("/setting", settingHandler.CreateOrUpdate)
			auth.DELETE("/setting", settingHandler.Delete)
			auth.POST("/setting/rename", settingHandler.Rename)
			auth.GET("/settings", settingHandler.List)

			// Sync log routes
			// 同步日志路由
			auth.GET("/sync-logs", syncLogHandler.List)

			// Token management routes (Restricted to webgui)
			// 令牌管理路由（限制仅 webgui 访问）
			tokenGroup := auth.Group("")
			tokenGroup.Use(func(c *gin.Context) {
				client := c.GetHeader("x-client")
				if client == "" {
					client = c.Query("client")
				}
				if !strings.EqualFold(client, "webgui") {
					response := pkgapp.NewResponse(c)
					response.ToResponse(code.ErrorInvalidUserAuthToken.WithDetails("Token management is only allowed from webgui"))
					c.Abort()
					return
				}
				c.Next()
			})
			{
				tokenGroup.GET("/tokens", tokenHandler.List)
				tokenGroup.POST("/token", tokenHandler.Create)
				tokenGroup.PUT("/token/:id", tokenHandler.UpdateScope)
				tokenGroup.DELETE("/token/:id", tokenHandler.Revoke)
				tokenGroup.GET("/token/:id/logs", tokenHandler.ListLogs)
			}
		}
	}
}
