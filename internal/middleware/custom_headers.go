// Package middleware provides custom Gin middlewares
// Package middleware 提供自定义 Gin 中间件
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// forbiddenHeaders list of security headers that cannot be overridden by custom headers
// forbiddenHeaders 无法被自定义响应头覆盖的安全响应头黑名单
var forbiddenHeaders = map[string]bool{
	"content-security-policy":   true,
	"x-frame-options":           true,
	"strict-transport-security": true,
	"x-content-type-options":    true,
}

// CustomHeaders creates middleware to set custom response headers
// CustomHeaders 创建设置自定义响应头的中间件
func CustomHeaders(headers map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for k, v := range headers {
			if !forbiddenHeaders[strings.ToLower(k)] {
				c.Header(k, v)
			}
		}
		c.Next()
	}
}
