package routers

import (
	"embed"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/limiter"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/lxzan/gws"
)

var methodLimiters = limiter.NewMethodLimiter().AddBuckets(
	limiter.BucketRule{
		Key:          "/auth",
		FillInterval: time.Second,
		Capacity:     10,
		Quantum:      10,
	},
	limiter.BucketRule{
		Key:          "/api/user/login",
		FillInterval: time.Second,
		Capacity:     5,
		Quantum:      1,
	},
	limiter.BucketRule{
		Key:          "/api/share/verify",
		FillInterval: time.Second,
		Capacity:     10,
		Quantum:      1,
	},
	limiter.BucketRule{
		Key:          "/api/share/note",
		FillInterval: time.Second,
		Capacity:     10,
		Quantum:      1,
	},
)

func NewRouter(frontendFiles embed.FS, appContainer *app.App, uni *ut.UniversalTranslator) *gin.Engine {

	// Get configuration
	// 获取配置
	cfg := appContainer.Config()

	var wss = pkgapp.NewWebsocketServer(pkgapp.WSConfig{
		GWSOption: gws.ServerOption{
			CheckUtf8Enabled: cfg.App.WebSocketCheckUtf8Enabled,
			ParallelEnabled:  cfg.App.WebSocketParallelEnabled, // Enable parallel message processing from config
			// 从配置开启并行消息处理
			Recovery: gws.Recovery, // Enable exception recovery
			// 开启异常恢复
			PermessageDeflate: gws.PermessageDeflate{
				Enabled:               cfg.App.WebSocketCompressionEnabled,
				Level:                 cfg.App.WebSocketCompressionLevel,
				Threshold:             cfg.App.WebSocketCompressionThreshold,
				ServerContextTakeover: true,
				ClientContextTakeover: true,
			}, // Enable compression from config
			// 从配置开启压缩
			ParallelGolimit:    cfg.App.WebSocketParallelGolimit,
			ReadMaxPayloadSize: int(util.ParseSize(cfg.App.WebSocketReadMaxPayloadSize, 1024*1024*64)), // Load from config, default 64MB
			// 从配置读取，默认 64MB
			WriteMaxPayloadSize: int(util.ParseSize(cfg.App.WebSocketWriteMaxPayloadSize, 1024*1024*64)), // Load from config, default 64MB
			// 从配置读取，默认 64MB
		},
	}, appContainer)
	appContainer.SetWSS(wss)

	// Initialize WebSocket routes
	// 初始化 WebSocket 路由
	initWebSocketRoutes(wss, appContainer)

	r := gin.New()
	r.Use(middleware.Proxy())
	r.Use(middleware.Cors())

	// Register Static routes
	// 注册静态资源路由
	// If independent ports are configured, the main port only provides necessary static files (for API/compatibility)
	// 如果配置了独立端口，则主端口仅提供必要的静态资源，不再提供 Web/Share 页面访问
	registerStaticFiles(r, frontendFiles, appContainer)
	if cfg.Server.WebGuiPort == "" {
		registerWebGuiRoutes(r, frontendFiles, appContainer)
	}
	if cfg.Server.SharePort == "" {
		registerShareRoutes(r, frontendFiles, appContainer)
	}

	// Register API routes
	// 注册 API 路由
	registerAPIRoutes(r, appContainer, wss, uni)

	// Register OpenAPI/Swagger routes
	// 注册 OpenAPI/Swagger 路由
	registerOpenAPIRoutes(r, frontendFiles)

	r.NoRoute(middleware.NoFound())

	return r
}

func NewWebGuiRouter(frontendFiles embed.FS, appContainer *app.App) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Proxy())
	r.Use(middleware.Cors())

	registerStaticFiles(r, frontendFiles, appContainer)
	registerWebGuiRoutes(r, frontendFiles, appContainer)

	r.NoRoute(middleware.NoFound())
	return r
}

func NewShareRouter(frontendFiles embed.FS, appContainer *app.App) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Proxy())
	r.Use(middleware.Cors())

	registerStaticFiles(r, frontendFiles, appContainer)
	registerShareRoutes(r, frontendFiles, appContainer)

	r.NoRoute(middleware.NoFound())
	return r
}
