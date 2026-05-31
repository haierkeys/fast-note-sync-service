// Package middleware provides custom Gin middlewares
// Package middleware 提供自定义 Gin 中间件
package middleware

import (
	"github.com/gin-gonic/gin"
)

// CustomHeaders creates middleware to set custom response headers
// CustomHeaders 创建设置自定义响应头的中间件
func CustomHeaders(headers map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for k, v := range headers {
			c.Header(k, v)
		}
		c.Next()
	}
}
