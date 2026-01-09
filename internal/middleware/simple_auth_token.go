/**
  @author: haierkeys
  @since: 2022/9/14
  @desc:
**/

package middleware

import (
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
)

// SimpleAuthTokenWithConfig 简单 Token 认证中间件（使用注入的配置）
func SimpleAuthTokenWithConfig(authToken string) gin.HandlerFunc {
	return func(c *gin.Context) {

		if authToken == "" {
			c.Next()
			return
		}

		response := app.NewResponse(c)

		var token string

		if s, exist := c.GetQuery("authorization"); exist {
			token = s
		} else if s, exist = c.GetQuery("Authorization"); exist {
			token = s
		} else if s = c.GetHeader("authorization"); len(s) != 0 {
			token = s
		} else if s = c.GetHeader("Authorization"); len(s) != 0 {
			token = s
		}

		if token != authToken {
			response.ToResponse(code.ErrorInvalidAuthToken)
			c.Abort()
			return
		}
		c.Next()
	}
}

// SimpleAuthToken 简单 Token 认证中间件（无配置，始终通过）
// Deprecated: 推荐使用 SimpleAuthTokenWithConfig
func SimpleAuthToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
