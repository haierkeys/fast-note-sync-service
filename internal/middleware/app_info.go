package middleware

import (
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"

	"github.com/gin-gonic/gin"
)

func AppInfo() gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Set("app_name", global.Name)
		c.Set("app_version", global.Version)
		c.Set("access_host", app.GetAccessHost(c))

		c.Next()
	}
}
