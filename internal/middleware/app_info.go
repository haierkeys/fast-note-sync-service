package middleware

import (
	"github.com/haierkeys/fast-note-sync-service/pkg/app"

	"github.com/gin-gonic/gin"
)

// AppInfoWithConfig creates middleware to set application information (supports dependency injection)
// AppInfoWithConfig 创建设置应用信息的中间件（支持依赖注入）
func AppInfoWithConfig(appName, appVersion string) gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Set("app_name", appName)
		c.Set("app_version", appVersion)
		c.Set("access_host", app.GetAccessHost(c))

		c.Next()
	}
}
