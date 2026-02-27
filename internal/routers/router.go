package routers

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"time"

	_ "github.com/haierkeys/fast-note-sync-service/docs"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	"github.com/haierkeys/fast-note-sync-service/internal/routers/api_router"
	"github.com/haierkeys/fast-note-sync-service/internal/routers/websocket_router"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/limiter"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/lxzan/gws"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var methodLimiters = limiter.NewMethodLimiter().AddBuckets(
	limiter.BucketRule{
		Key:          "/auth",
		FillInterval: time.Second,
		Capacity:     10,
		Quantum:      10,
	},
)

func NewRouter(frontendFiles embed.FS, appContainer *app.App, uni *ut.UniversalTranslator) *gin.Engine {

	// Get configuration
	// 获取配置
	cfg := appContainer.Config()

	var wss = pkgapp.NewWebsocketServer(pkgapp.WSConfig{
		GWSOption: gws.ServerOption{
			CheckUtf8Enabled: true,
			ParallelEnabled:  true, // Enable parallel message processing
			// 开启并行消息处理
			Recovery: gws.Recovery, // Enable exception recovery
			// 开启异常恢复
			PermessageDeflate: gws.PermessageDeflate{Enabled: true}, // Enable compression
			// 开启压缩
			ParallelGolimit:    8,
			ReadMaxPayloadSize: int(util.ParseSize(cfg.App.WebSocketReadMaxPayloadSize, 1024*1024*64)), // Load from config, default 64MB
			// 从配置读取，默认 64MB
			WriteMaxPayloadSize: int(util.ParseSize(cfg.App.WebSocketWriteMaxPayloadSize, 1024*1024*64)), // Load from config, default 64MB
			// 从配置读取，默认 64MB
		},
	}, appContainer)

	// Create WebSocket Handlers (injected App Container)
	// 创建 WebSocket Handlers（注入 App Container）
	noteWSHandler := websocket_router.NewNoteWSHandler(appContainer)
	folderWSHandler := websocket_router.NewFolderWSHandler(appContainer)
	fileWSHandler := websocket_router.NewFileWSHandler(appContainer)
	settingWSHandler := websocket_router.NewSettingWSHandler(appContainer)

	// Note
	wss.Use(dto.NoteReceiveModify, noteWSHandler.NoteModify)
	wss.Use(dto.NoteReceiveDelete, noteWSHandler.NoteDelete)
	wss.Use(dto.NoteReceiveRename, noteWSHandler.NoteRename)
	wss.Use(dto.NoteReceiveRePush, noteWSHandler.NoteRePush)
	wss.Use(dto.NoteReceiveCheck, noteWSHandler.NoteModifyCheck)
	wss.Use(dto.NoteReceiveSync, noteWSHandler.NoteSync)

	// Folder
	wss.Use(dto.FolderReceiveSync, folderWSHandler.FolderSync)
	wss.Use(dto.FolderReceiveModify, folderWSHandler.FolderModify)
	wss.Use(dto.FolderReceiveDelete, folderWSHandler.FolderDelete)
	wss.Use(dto.FolderReceiveRename, folderWSHandler.FolderRename)

	// Setting
	wss.Use(dto.SettingReceiveModify, settingWSHandler.SettingModify)
	wss.Use(dto.SettingReceiveDelete, settingWSHandler.SettingDelete)
	wss.Use(dto.SettingReceiveCheck, settingWSHandler.SettingModifyCheck)
	wss.Use(dto.SettingReceiveSync, settingWSHandler.SettingSync)
	wss.Use(dto.SettingReceiveClear, settingWSHandler.SettingClear)

	// Attachment
	wss.Use(dto.FileReceiveSync, fileWSHandler.FileSync)
	wss.Use(dto.FileReceiveUploadCheck, fileWSHandler.FileUploadCheck)
	wss.Use(dto.FileReceiveRename, fileWSHandler.FileRename)
	wss.Use(dto.FileReceiveDelete, fileWSHandler.FileDelete)
	wss.Use(dto.FileReceiveChunkDownload, fileWSHandler.FileChunkDownload)
	wss.Use(dto.FileReceiveRePush, fileWSHandler.FileRePush)

	// Attachment chunk upload
	wss.UseBinary(dto.VaultFileMsgType, fileWSHandler.FileUploadChunkBinary)

	wss.UseUserVerify(noteWSHandler.UserInfo)

	frontendAssets, _ := fs.Sub(frontendFiles, "frontend/assets")
	frontendStatic, _ := fs.Sub(frontendFiles, "frontend/static")
	frontendIndexContent, _ := frontendFiles.ReadFile("frontend/index.html")

	r := gin.New()
	r.Use(middleware.Cors())

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/webgui")
	})
	r.GET("/webgui/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", frontendIndexContent)
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

	r.Group("/assets", cacheMiddleware).StaticFS("/", http.FS(frontendAssets))
	r.Group("/static", cacheMiddleware).StaticFS("/", http.FS(frontendStatic))
	r.Group("/user_static", cacheMiddleware).Static("/", userStaticPath)

	api := r.Group("/api")
	{
		api.Use(middleware.AppInfoWithConfig(app.Name, appContainer.Version().Version))
		api.Use(gin.Logger())
		api.Use(middleware.TraceMiddlewareWithConfig(cfg.Tracer.Enabled, cfg.Tracer.Header)) // Trace ID middleware
		// Trace ID 中间件
		api.Use(middleware.RateLimiter(methodLimiters))
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
		adminControlHandler := api_router.NewAdminControlHandler(appContainer)
		shareHandler := api_router.NewShareHandler(appContainer)
		storageHandler := api_router.NewStorageHandler(appContainer)
		backupHandler := api_router.NewBackupHandler(appContainer)
		gitSyncHandler := api_router.NewGitSyncHandler(appContainer)

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
		auth.Use(middleware.UserAuthTokenWithConfig(cfg.Security.AuthTokenKey))
		{
			// Create share
			// 创建分享
			auth.POST("/share", shareHandler.Create)

			// Admin config interface
			// 管理员配置接口
			auth.GET("/admin/config", adminControlHandler.GetConfig)
			auth.POST("/admin/config", adminControlHandler.UpdateConfig)
			auth.GET("/admin/systeminfo", adminControlHandler.GetSystemInfo)
			auth.GET("/admin/upgrade", adminControlHandler.Upgrade)

			auth.POST("/user/change_password", userHandler.UserChangePassword)
			auth.GET("/user/info", userHandler.UserInfo)
			auth.GET("/vault", vaultHandler.List)
			auth.POST("/vault", vaultHandler.CreateOrUpdate)
			auth.DELETE("/vault", vaultHandler.Delete)

			auth.GET("/note", noteHandler.Get)
			auth.POST("/note", noteHandler.CreateOrUpdate)
			auth.DELETE("/note", noteHandler.Delete)
			auth.PUT("/note/restore", noteHandler.Restore)
			auth.GET("/notes", noteHandler.List)

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
			auth.POST("/note/move", noteHandler.Move)

			// Note link operations
			auth.GET("/note/backlinks", noteHandler.GetBacklinks)
			auth.GET("/note/outlinks", noteHandler.GetOutlinks)

			auth.GET("/file", fileHandler.GetInfo)
			auth.OPTIONS("/file", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			auth.GET("/file/info", fileHandler.Get)
			auth.OPTIONS("/file/info", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			auth.DELETE("/file", fileHandler.Delete)
			auth.PUT("/file/restore", fileHandler.Restore)
			auth.GET("/files", fileHandler.List)
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
		}

	}

	// Swagger UI (outside auth group to ensure public access)
	// Swagger UI (放在 auth 组外，确保可以公开访问)
	r.GET("/docs/*any", func(c *gin.Context) {
		p := c.Param("any")
		if p == "" || p == "/" {
			c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
			return
		}
		ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
	})

	// Read debug page from embedded FS
	debugPageContent, _ := frontendFiles.ReadFile("docs/test_ws_debug.html")
	r.GET("/ws_debug", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", debugPageContent)
	})

	// Read swagger files from embedded FS
	swaggerJSON, _ := frontendFiles.ReadFile("docs/swagger.json")
	r.GET("/rest_doc/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/rest_doc.json")
	})
	r.GET("/rest_doc.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", swaggerJSON)
	})
	swaggerYAML, _ := frontendFiles.ReadFile("docs/swagger.yaml")
	r.GET("/rest_doc.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/x-yaml; charset=utf-8", swaggerYAML)
	})

	if cfg.Storage.LocalFS.HttpfsIsEnable && cfg.Storage.LocalFS.IsEnabled {
		r.StaticFS(cfg.Storage.LocalFS.SavePath, http.Dir(cfg.Storage.LocalFS.SavePath))
		r.OPTIONS(cfg.Storage.LocalFS.SavePath+"/*filepath", func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		})
	}

	r.NoRoute(middleware.NoFound())

	return r
}
