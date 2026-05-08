package routers

import (
	"embed"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/haierkeys/fast-note-sync-service/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func registerOpenAPIRoutes(r *gin.Engine, frontendFiles embed.FS) {
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
	swaggerJSON, _ := frontendFiles.ReadFile("docs/swagger.yaml")
	r.GET("/openapi/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/openapi.json")
	})
	r.GET("/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", swaggerJSON)
	})
	swaggerYAML, _ := frontendFiles.ReadFile("docs/swagger.yaml")
	r.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/x-yaml; charset=utf-8", swaggerYAML)
	})
}
