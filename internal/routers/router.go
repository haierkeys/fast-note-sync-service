package routers

import (
	"embed"
	"io/fs"
	"net/http"
	"time"

	_ "github.com/haierkeys/obsidian-better-sync-service/docs"
	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/haierkeys/obsidian-better-sync-service/internal/middleware"
	"github.com/haierkeys/obsidian-better-sync-service/internal/routers/api_router"
	"github.com/haierkeys/obsidian-better-sync-service/internal/routers/websocket_router"

	"github.com/haierkeys/obsidian-better-sync-service/pkg/app"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/limiter"

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

	var wss = app.NewWebsocketServer(app.WebsocketServerConfig{
		GWSOption: gws.ServerOption{
			CheckUtf8Enabled:  true,
			ParallelEnabled:   true,                                 // 开启并行消息处理
			Recovery:          gws.Recovery,                         // 开启异常恢复
			PermessageDeflate: gws.PermessageDeflate{Enabled: true}, // 开启压缩
			// ReadMaxPayloadSize:    1024 * 1024 * 16,                     // 设置最大读取缓冲区大小
			// WriteMaxPayloadSize:   1024 * 1024 * 16,                     // 设置最大写入缓冲区大小
		},
	})

	wss.Use("FileCreate", websocket_router.FileCreate)

	frontendAssets, _ := fs.Sub(frontendFiles, "frontend/assets")
	frontendIndexContent, _ := frontendFiles.ReadFile("frontend/index.html")

	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", frontendIndexContent)
	})
	r.StaticFS("/assets", http.FS(frontendAssets))

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

		userApiR := api.Group("/user")
		{
			userApiR.GET("/sync", wss.Run())
		}
	}
	r.Use(middleware.Cors())
	r.NoRoute(middleware.NoFound())

	return r
}
