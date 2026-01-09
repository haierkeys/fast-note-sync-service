package middleware

import (
	"github.com/haierkeys/fast-note-sync-service/pkg/app"

	"github.com/gin-gonic/gin"
)

// AppInfoWithConfig 创建带配置的应用信息中间件（支持依赖注入）
func AppInfoWithConfig(appName, appVersion string) gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Set("app_name", appName)
		c.Set("app_version", appVersion)
		c.Set("access_host", app.GetAccessHost(c))

		c.Next()
	}
}
