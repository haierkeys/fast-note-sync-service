package middleware

import (
	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/app"

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
