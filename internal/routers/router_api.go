package routers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	"github.com/haierkeys/fast-note-sync-service/internal/routers/api_router"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
)

func registerAPIRoutes(r *gin.Engine, appContainer *app.App, wss *pkgapp.WebsocketServer, uni *ut.UniversalTranslator) {
	cfg := appContainer.Config()
	api := r.Group("/api")
	{
		api.Use(middleware.AppInfoWithConfig(app.Name, appContainer.Version().Version))
		api.Use(gin.Logger())
		api.Use(middleware.TraceMiddlewareWithConfig(*cfg.Tracer.Enabled, cfg.Tracer.Header)) // Trace ID middleware
		// Trace ID 中间件
		api.Use(middleware.RateLimiter(methodLimiters))

		// MCP routes
		registerMCPRoutes(api, appContainer, wss)

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

		// No-auth WebGUI restricted routes
		// 免认证但仅限 WebGUI 访问的路由组
		noAuthWebgui := api.Group("")
		noAuthWebgui.Use(pkgapp.RequireWebGUI())
		{
			noAuthWebgui.POST("/user/register", userHandler.Register)
			noAuthWebgui.POST("/user/login", userHandler.Login)
			noAuthWebgui.GET("/webgui/config", adminControlHandler.Config)
		}
		api.GET("/user/sync", wss.Run())

		// Add server version interface (no auth required)
		// 添加服务端版本号接口（无需认证）
		api.GET("/version", versionHandler.ServerVersion)
		api.GET("/support", versionHandler.Support)

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
			auth.POST("/auth/logout", userHandler.Logout)
			auth.POST("/share", shareHandler.Create)
			auth.POST("/share/password", shareHandler.UpdatePassword)
			auth.GET("/share", shareHandler.Query)
			auth.DELETE("/share", shareHandler.Cancel)
			auth.POST("/share/short_link", shareHandler.CreateShortLink)
			auth.GET("/shares", shareHandler.List)

			// Admin config interface
			// 管理员配置接口
			auth.GET("/admin/upgrade", adminControlHandler.Upgrade)
			auth.GET("/admin/ws_clients", adminControlHandler.GetWSClients)
			auth.DELETE("/admin/ws_client/:traceId", adminControlHandler.KickWSClient)

			auth.GET("/user/info", userHandler.UserInfo)

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

			auth.GET("/setting", settingHandler.Get)
			auth.POST("/setting", settingHandler.CreateOrUpdate)
			auth.DELETE("/setting", settingHandler.Delete)
			auth.POST("/setting/rename", settingHandler.Rename)
			auth.GET("/settings", settingHandler.List)

			// WebGUI restricted routes
			// 仅限 WebGUI 访问的路由组
			webguiGroup := auth.Group("")
			webguiGroup.Use(pkgapp.RequireWebGUI())
			{
				// User management routes
				// 用户管理接口
				webguiGroup.POST("/user/change_password", userHandler.UserChangePassword)

				// Vault management routes
				// 笔记库管理接口
				webguiGroup.GET("/vault", vaultHandler.List)
				webguiGroup.POST("/vault", vaultHandler.CreateOrUpdate)
				webguiGroup.DELETE("/vault", vaultHandler.Delete)
				webguiGroup.POST("/vault/rebuild-index", vaultHandler.RebuildIndex)

				// Admin config interface
				// 管理员配置接口
				webguiGroup.GET("/admin/check", adminControlHandler.CheckAdmin)
				webguiGroup.GET("/admin/config", adminControlHandler.GetConfig)
				webguiGroup.POST("/admin/config", adminControlHandler.UpdateConfig)
				webguiGroup.GET("/admin/config/user_database", adminControlHandler.GetUserDatabaseConfig)
				webguiGroup.POST("/admin/config/user_database", adminControlHandler.UpdateUserDatabaseConfig)
				webguiGroup.POST("/admin/config/user_database/test", adminControlHandler.ValidateUserDatabaseConfig)
				webguiGroup.GET("/admin/config/cloudflare", adminControlHandler.GetCloudflareConfig)
				webguiGroup.POST("/admin/config/cloudflare", adminControlHandler.UpdateCloudflareConfig)
				webguiGroup.GET("/admin/systeminfo", adminControlHandler.GetSystemInfo)
				webguiGroup.GET("/admin/restart", adminControlHandler.Restart)
				webguiGroup.GET("/admin/gc", adminControlHandler.GC)
				webguiGroup.GET("/admin/cloudflared_tunnel_download", adminControlHandler.CloudflaredTunnelDownload)

				// Storage management routes
				// 存储配置接口
				webguiGroup.GET("/storage", storageHandler.List)
				webguiGroup.POST("/storage", storageHandler.CreateOrUpdate)
				webguiGroup.GET("/storage/enabled_types", storageHandler.EnabledTypes)
				webguiGroup.POST("/storage/validate", storageHandler.Validate)
				webguiGroup.DELETE("/storage", storageHandler.Delete)

				// Backup routes
				// 本地备份接口
				webguiGroup.GET("/backup/configs", backupHandler.GetConfigs)
				webguiGroup.POST("/backup/config", backupHandler.UpdateConfig)
				webguiGroup.DELETE("/backup/config", backupHandler.DeleteConfig)
				webguiGroup.GET("/backup/historys", backupHandler.ListHistory)
				webguiGroup.POST("/backup/execute", backupHandler.Execute)

				// Git sync routes
				// Git 同步接口
				webguiGroup.GET("/git-sync/configs", gitSyncHandler.GetConfigs)
				webguiGroup.POST("/git-sync/config", gitSyncHandler.UpdateConfig)
				webguiGroup.DELETE("/git-sync/config", gitSyncHandler.DeleteConfig)
				webguiGroup.POST("/git-sync/validate", gitSyncHandler.Validate)
				webguiGroup.DELETE("/git-sync/config/clean", gitSyncHandler.CleanWorkspace)
				webguiGroup.POST("/git-sync/config/execute", gitSyncHandler.Execute)
				webguiGroup.GET("/git-sync/histories", gitSyncHandler.GetHistories)

				// Sync log routes
				// 同步日志路由
				webguiGroup.GET("/sync-logs", syncLogHandler.List)

				// Token management routes
				// 令牌管理路由
				webguiGroup.GET("/tokens", tokenHandler.List)
				webguiGroup.POST("/token", tokenHandler.Create)
				webguiGroup.PUT("/token/:id", tokenHandler.Update)
				webguiGroup.DELETE("/token/:id", tokenHandler.Revoke)
				webguiGroup.POST("/token/:id/rotate", tokenHandler.Rotate)
				webguiGroup.GET("/token/:id/logs", tokenHandler.ListLogs)
			}
		}
	}
}
