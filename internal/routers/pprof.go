package routers

import (
	"net/http"
	"net/http/pprof"

	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/haierkeys/obsidian-better-sync-service/internal/middleware"
	"github.com/haierkeys/obsidian-better-sync-service/internal/routers/api_router"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// DefaultPrefix url prefix of pprof
	DefaultPrefix = "/debug/pprof"
)

func NewPrivateRouter() *gin.Engine {

	r := gin.New()

	if global.Config.Server.RunMode == "debug" {
		r.Use(gin.Recovery())
	} else {
		r.Use(middleware.Recovery())
	}

	// prom监控
	r.GET("/debug/vars", api_router.Expvar)
	r.GET("metrics", gin.WrapH(promhttp.Handler()))

	if global.Config.Server.RunMode == "debug" {
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

func pprofHandler(h http.HandlerFunc) gin.HandlerFunc {
	handler := h
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
