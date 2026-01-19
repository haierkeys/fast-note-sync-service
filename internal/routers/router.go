package routers

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"time"

	_ "github.com/haierkeys/fast-note-sync-service/docs"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	"github.com/haierkeys/fast-note-sync-service/internal/routers/api_router"
	"github.com/haierkeys/fast-note-sync-service/internal/routers/websocket_router"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/limiter"

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

	// 获取配置
	cfg := appContainer.Config()

	var wss = pkgapp.NewWebsocketServer(pkgapp.WSConfig{
		GWSOption: gws.ServerOption{
			CheckUtf8Enabled:    true,
			ParallelEnabled:     true,                                 // 开启并行消息处理
			Recovery:            gws.Recovery,                         // 开启异常恢复
			PermessageDeflate:   gws.PermessageDeflate{Enabled: true}, // 开启压缩
			ParallelGolimit:     8,
			ReadMaxPayloadSize:  1024 * 1024 * 64, // 设置最大读取缓冲区大小 64MB
			WriteMaxPayloadSize: 1024 * 1024 * 64, // 设置最大写入缓冲区大小 64MB
		},
	}, appContainer)

	// 创建 WebSocket Handlers（注入 App Container）
	noteWSHandler := websocket_router.NewNoteWSHandler(appContainer)
	fileWSHandler := websocket_router.NewFileWSHandler(appContainer)
	settingWSHandler := websocket_router.NewSettingWSHandler(appContainer)

	// 修改 创建
	wss.Use("NoteModify", noteWSHandler.NoteModify)
	//删除
	wss.Use("NoteDelete", noteWSHandler.NoteDelete)
	//重命名
	wss.Use("NoteRename", noteWSHandler.NoteRename)
	// 笔记检查
	wss.Use("NoteCheck", noteWSHandler.NoteModifyCheck)
	// 基于mtime的更新通知
	wss.Use("NoteSync", noteWSHandler.NoteSync)

	// 配置同步
	wss.Use("SettingModify", settingWSHandler.SettingModify)
	wss.Use("SettingDelete", settingWSHandler.SettingDelete)
	wss.Use("SettingCheck", settingWSHandler.SettingModifyCheck)
	wss.Use("SettingSync", settingWSHandler.SettingSync)

	// 附件同步
	wss.Use("FileSync", fileWSHandler.FileSync)
	//附件上传前检查
	wss.Use("FileUploadCheck", fileWSHandler.FileUploadCheck)
	//附件删除
	wss.Use("FileDelete", fileWSHandler.FileDelete)

	wss.Use("FileChunkDownload", fileWSHandler.FileChunkDownload)

	//附件上传分块
	wss.UseBinary(websocket_router.VaultFileSync, fileWSHandler.FileUploadChunkBinary)

	// WebGUI 配置（使用注入的配置）
	webGUIWSHandler := websocket_router.NewWebGUIWSHandler(appContainer)
	wss.Use("WebGUIConfigGet", webGUIWSHandler.WebGUIConfigGet)

	wss.UseUserVerify(noteWSHandler.UserInfo)

	frontendAssets, _ := fs.Sub(frontendFiles, "frontend/assets")
	frontendStatic, _ := fs.Sub(frontendFiles, "frontend/static")
	frontendIndexContent, _ := frontendFiles.ReadFile("frontend/index.html")

	r := gin.New()

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
		api.Use(middleware.TraceMiddlewareWithConfig(cfg.Tracer.Enabled, cfg.Tracer.Header)) // Trace ID 中间件
		api.Use(middleware.RateLimiter(methodLimiters))
		api.Use(middleware.ContextTimeout(time.Duration(cfg.App.DefaultContextTimeout) * time.Second))
		api.Use(middleware.Cors())
		api.Use(middleware.LangWithTranslator(uni))
		api.Use(middleware.AccessLogWithLogger(appContainer.Logger()))
		api.Use(middleware.RecoveryWithLogger(appContainer.Logger()))

		// 创建 Handlers（注入 App Container）
		userHandler := api_router.NewUserHandler(appContainer)
		vaultHandler := api_router.NewVaultHandler(appContainer)
		noteHandler := api_router.NewNoteHandler(appContainer, wss)
		fileHandler := api_router.NewFileHandler(appContainer, wss)
		noteHistoryHandler := api_router.NewNoteHistoryHandler(appContainer, wss)
		versionHandler := api_router.NewVersionHandler(appContainer)
		webGUIHandler := api_router.NewWebGUIHandler(appContainer)
		shareHandler := api_router.NewShareHandler(appContainer)

		api.POST("/user/register", userHandler.Register)
		api.POST("/user/login", userHandler.Login)
		api.GET("/user/sync", wss.Run())

		// 添加服务端版本号接口（无需认证）
		api.GET("/version", versionHandler.ServerVersion)
		api.GET("/webgui/config", webGUIHandler.Config)

		// 分享路由组 (受控的只读访问)
		share := api.Group("/share")
		share.Use(middleware.ShareAuthToken(appContainer.ShareService))
		{
			share.GET("/note", shareHandler.NoteGet) // 获取分享的笔记
			share.GET("/file", shareHandler.FileGet) // 获取分享的文件内容
		}

		// 需要认证的路由组
		auth := api.Group("/")
		auth.Use(middleware.UserAuthTokenWithConfig(cfg.Security.AuthTokenKey))
		{
			// 创建分享
			auth.POST("/share", shareHandler.Create)

			// 管理员配置接口
			auth.GET("/admin/config", webGUIHandler.GetConfig)
			auth.POST("/admin/config", webGUIHandler.UpdateConfig)

			auth.POST("/user/change_password", userHandler.UserChangePassword)
			auth.GET("/user/info", userHandler.UserInfo)
			auth.GET("/vault", vaultHandler.List)
			auth.POST("/vault", vaultHandler.CreateOrUpdate)
			auth.DELETE("/vault", vaultHandler.Delete)

			auth.GET("/note", noteHandler.Get)
			auth.GET("/note/file", fileHandler.GetContent)
			auth.POST("/note", noteHandler.CreateOrUpdate)
			auth.DELETE("/note", noteHandler.Delete)
			auth.PUT("/note/restore", noteHandler.Restore)
			auth.GET("/notes", noteHandler.List)

			auth.GET("/file", fileHandler.GetContent)
			auth.DELETE("/file", fileHandler.Delete)
			auth.GET("/files", fileHandler.List)

			auth.GET("/note/history", noteHistoryHandler.Get)
			auth.GET("/note/histories", noteHistoryHandler.List)
			auth.PUT("/note/history/restore", noteHistoryHandler.Restore)
		}

		// Swagger UI (放在 auth 组外，确保可以公开访问)
		api.GET("/docs/*any", func(c *gin.Context) {
			p := c.Param("any")
			if p == "" || p == "/" {
				c.Redirect(http.StatusMovedPermanently, "/api/docs/index.html")
				return
			}
			ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
		})
	}

	if cfg.App.UploadSavePath != "" {
		r.StaticFS(cfg.App.UploadSavePath, http.Dir(cfg.App.UploadSavePath))
	}
	r.Use(middleware.Cors())
	r.NoRoute(middleware.NoFound())

	return r
}
