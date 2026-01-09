package routers

import (
	"embed"
	"io/fs"
	"net/http"
	"time"

	_ "github.com/haierkeys/fast-note-sync-service/docs"
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	"github.com/haierkeys/fast-note-sync-service/internal/routers/api_router"
	"github.com/haierkeys/fast-note-sync-service/internal/routers/websocket_router"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/limiter"

	"github.com/gin-gonic/gin"
	"github.com/lxzan/gws"
)

var methodLimiters = limiter.NewMethodLimiter().AddBuckets(
	limiter.BucketRule{
		Key:          "/auth",
		FillInterval: time.Second,
		Capacity:     10,
		Quantum:      10,
	},
)

func NewRouter(frontendFiles embed.FS) *gin.Engine {

	var wss = app.NewWebsocketServer(app.WSConfig{
		GWSOption: gws.ServerOption{
			CheckUtf8Enabled:    true,
			ParallelEnabled:     true,                                 // 开启并行消息处理
			Recovery:            gws.Recovery,                         // 开启异常恢复
			PermessageDeflate:   gws.PermessageDeflate{Enabled: true}, // 开启压缩
			ParallelGolimit:     8,
			ReadMaxPayloadSize:  1024 * 1024 * 64, // 设置最大读取缓冲区大小 64MB
			WriteMaxPayloadSize: 1024 * 1024 * 64, // 设置最大写入缓冲区大小 64MB
		},
	})
	// 修改 创建
	wss.Use("NoteModify", websocket_router.NoteModify)
	//删除
	wss.Use("NoteDelete", websocket_router.NoteDelete)
	//重命名
	wss.Use("NoteRename", websocket_router.NoteRename)
	// 笔记检查
	wss.Use("NoteCheck", websocket_router.NoteModifyCheck)
	// 基于mtime的更新通知
	wss.Use("NoteSync", websocket_router.NoteSync)

	// 配置同步
	wss.Use("SettingModify", websocket_router.SettingModify)
	wss.Use("SettingDelete", websocket_router.SettingDelete)
	wss.Use("SettingCheck", websocket_router.SettingModifyCheck)
	wss.Use("SettingSync", websocket_router.SettingSync)

	// 附件同步
	wss.Use("FileSync", websocket_router.FileSync)
	//附件上传前检查
	wss.Use("FileUploadCheck", websocket_router.FileUploadCheck)
	//附件删除
	wss.Use("FileDelete", websocket_router.FileDelete)

	wss.Use("FileChunkDownload", websocket_router.FileChunkDownload)

	//附件上传分块
	wss.UseBinary(websocket_router.VaultFileSync, websocket_router.FileUploadChunkBinary)

	// WebGUI 配置
	wss.Use("WebGUIConfigGet", websocket_router.WebGUIConfigGet)

	wss.UseUserVerify(websocket_router.UserInfo)

	frontendAssets, _ := fs.Sub(frontendFiles, "frontend/assets")
	frontendStatic, _ := fs.Sub(frontendFiles, "frontend/static")
	frontendIndexContent, _ := frontendFiles.ReadFile("frontend/index.html")

	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", frontendIndexContent)
	})

	cacheMiddleware := func(c *gin.Context) {
		// 设置强缓存，缓存一年
		c.Header("Cache-Control", "public, s-maxage=31536000, max-age=31536000, must-revalidate")
		c.Next()
	}

	r.Group("/assets", cacheMiddleware).StaticFS("/", http.FS(frontendAssets))
	r.Group("/static", cacheMiddleware).StaticFS("/", http.FS(frontendStatic))

	api := r.Group("/api")
	{
		api.Use(middleware.AppInfo())
		api.Use(gin.Logger())
		api.Use(middleware.RateLimiter(methodLimiters))
		api.Use(middleware.ContextTimeout(time.Duration(global.Config.App.DefaultContextTimeout) * time.Second))
		api.Use(middleware.Cors())
		api.Use(middleware.Lang())
		api.Use(middleware.AccessLog())
		api.Use(middleware.Recovery())

		api.POST("/user/register", api_router.NewUser().Register)
		api.POST("/user/login", api_router.NewUser().Login)
		api.GET("/user/sync", wss.Run())

		// 添加服务端版本号接口（无需认证）
		api.GET("/version", api_router.NewVersion().ServerVersion)
		api.GET("/webgui/config", api_router.NewWebGUI().Config)

		api.Use(middleware.UserAuthToken()).GET("/admin/config", api_router.NewWebGUI().GetConfig)
		api.Use(middleware.UserAuthToken()).POST("/admin/config", api_router.NewWebGUI().UpdateConfig)

		api.Use(middleware.UserAuthToken()).POST("/user/change_password", api_router.NewUser().UserChangePassword)
		api.Use(middleware.UserAuthToken()).GET("/user/info", api_router.NewUser().UserInfo)
		api.Use(middleware.UserAuthToken()).GET("/vault", api_router.NewVault().List)
		api.Use(middleware.UserAuthToken()).POST("/vault", api_router.NewVault().CreateOrUpdate)
		api.Use(middleware.UserAuthToken()).DELETE("/vault", api_router.NewVault().Delete)

		noteApiWithWss := api_router.NewNote(wss)

		api.Use(middleware.UserAuthToken()).GET("/note", noteApiWithWss.Get)
		api.Use(middleware.UserAuthToken()).GET("/note/file", noteApiWithWss.GetFileContent)
		api.Use(middleware.UserAuthToken()).POST("/note", noteApiWithWss.CreateOrUpdate)
		api.Use(middleware.UserAuthToken()).DELETE("/note", noteApiWithWss.Delete)
		api.Use(middleware.UserAuthToken()).GET("/notes", noteApiWithWss.List)

		api.Use(middleware.UserAuthToken()).GET("/note/history", api_router.NewNoteHistory().Get)
		api.Use(middleware.UserAuthToken()).GET("/note/histories", api_router.NewNoteHistory().List)
	}

	if global.Config.App.UploadSavePath != "" {
		r.StaticFS(global.Config.App.UploadSavePath, http.Dir(global.Config.App.UploadSavePath))
	}
	r.Use(middleware.Cors())
	r.NoRoute(middleware.NoFound())

	return r
}
