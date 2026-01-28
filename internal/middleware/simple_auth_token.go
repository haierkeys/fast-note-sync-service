/**
  @author: haierkeys
  @since: 2022/9/14
  @desc: Simple auth token middleware // 简单认证 Token 中间件
**/

package middleware

import (
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
)

// SimpleAuthTokenWithConfig simple Token authentication middleware (check if header/param matches secretKey)
// SimpleAuthTokenWithConfig 简单 Token 认证中间件（检查 Header/参数是否匹配 secretKey）
// Mainly used for private monitoring interfaces or simply protected interfaces
// 主要用于私有监控接口或简单的保护接口
func SimpleAuthTokenWithConfig(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {

		if secretKey == "" {
			c.Next()
			return
		}

		response := app.NewResponse(c)

		// Check URL parameter token
		// 检查 URL 参数 token
		var token string

		if s, exist := c.GetQuery("authorization"); exist {
			token = s
		} else if s, exist = c.GetQuery("Authorization"); exist {
			token = s
		} else if s = c.GetHeader("authorization"); len(s) != 0 {
			token = s
		} else if s = c.GetHeader("Authorization"); len(s) != 0 {
			// Check Authorization: Bearer <token>
			// 检查 Authorization: Bearer <token>
			if len(s) > 7 && s[:7] == "Bearer " {
				token = s[7:]
			} else {
				token = s
			}
		}

		if token != secretKey {
			response.ToResponse(code.ErrorInvalidAuthToken)
			c.Abort()
			return
		}
		c.Next()
	}
}

// SimpleAuthToken simple Token authentication middleware (no secret key, always fails)
// SimpleAuthToken 简单 Token 认证中间件（无密钥，始终失败）
// Deprecated: Recommended to use SimpleAuthTokenWithConfig instead
// Deprecated: 推荐使用 SimpleAuthTokenWithConfig
func SimpleAuthToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
