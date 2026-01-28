package routers

import (
	"net/http"
	"net/http/pprof"

	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	"github.com/haierkeys/fast-note-sync-service/internal/routers/api_router"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	// DefaultPrefix url prefix of pprof
	DefaultPrefix = "/debug/pprof"
)

// NewPrivateRouterWithLogger creates private router (using injected logger)
// NewPrivateRouterWithLogger 创建私有路由（使用注入的日志器）
func NewPrivateRouterWithLogger(runMode string, logger *zap.Logger) *gin.Engine {

	r := gin.New()

	if runMode == "debug" {
		r.Use(gin.Recovery())
	} else {
		r.Use(middleware.RecoveryWithLogger(logger))
	}

	// prom monitoring
	// prom监控
	r.GET("/debug/vars", api_router.Expvar)
	r.GET("metrics", gin.WrapH(promhttp.Handler()))

	if runMode == "debug" {
		p := r.Group("pprof")
		{
			p.GET("/", pprofHandler(pprof.Index))
			p.GET("/cmdline", pprofHandler(pprof.Cmdline))
			p.GET("/profile", pprofHandler(pprof.Profile))
			p.POST("/symbol", pprofHandler(pprof.Symbol))
			p.GET("/symbol", pprofHandler(pprof.Symbol))
			p.GET("/trace", pprofHandler(pprof.Trace))
			p.GET("/allocs", pprofHandler(pprof.Handler("allocs").ServeHTTP))
			p.GET("/block", pprofHandler(pprof.Handler("block").ServeHTTP))
			p.GET("/goroutine", pprofHandler(pprof.Handler("goroutine").ServeHTTP))
			p.GET("/heap", pprofHandler(pprof.Handler("heap").ServeHTTP))
			p.GET("/mutex", pprofHandler(pprof.Handler("mutex").ServeHTTP))
			p.GET("/threadcreate", pprofHandler(pprof.Handler("threadcreate").ServeHTTP))
		}
	}

	return r
}

// NewPrivateRouterWithConfig creates private router (using injected config, using nop logger)
// NewPrivateRouterWithConfig 创建私有路由（使用注入的配置，使用 nop logger）
// Deprecated: Recommended to use NewPrivateRouterWithLogger
// Deprecated: 推荐使用 NewPrivateRouterWithLogger
func NewPrivateRouterWithConfig(runMode string) *gin.Engine {
	return NewPrivateRouterWithLogger(runMode, zap.NewNop())
}

// NewPrivateRouter creates private router (using default release mode)
// NewPrivateRouter 创建私有路由（使用默认 release 模式）
// Deprecated: Recommended to use NewPrivateRouterWithLogger
// Deprecated: 推荐使用 NewPrivateRouterWithLogger
func NewPrivateRouter() *gin.Engine {
	return NewPrivateRouterWithConfig("release")
}

func pprofHandler(h http.HandlerFunc) gin.HandlerFunc {
	handler := h
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
