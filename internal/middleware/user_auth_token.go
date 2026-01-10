package middleware

import (
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
)

// UserAuthTokenWithConfig 用户 Token 认证中间件（使用注入的密钥）
func UserAuthTokenWithConfig(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string
		response := app.NewResponse(c)

		if s, exist := c.GetQuery("authorization"); exist {
			token = s
		} else if s, exist := c.GetQuery("Authorization"); exist {
			token = s
		} else if s := c.GetHeader("authorization"); len(s) != 0 {
			token = s
		} else if s := c.GetHeader("Authorization"); len(s) != 0 {
			token = s
		} else if s, exist := c.GetQuery("token"); exist {
			token = s
		} else if s, exist := c.GetQuery("Token"); exist {
			token = s
		} else if s = c.GetHeader("token"); len(s) != 0 {
			token = s
		} else if s = c.GetHeader("Token"); len(s) != 0 {
			token = s
		}

		if token == "" {
			response.ToResponse(code.ErrorNotUserAuthToken)
			c.Abort()
			return
		}

		if user, err := app.ParseTokenWithKey(token, secretKey); err != nil {
			response.ToResponse(code.ErrorInvalidUserAuthToken)
			c.Abort()
			return
		} else {
			c.Set("user_token", user)
		}

		c.Next()
	}
}

// UserAuthToken 用户 Token 认证中间件（无密钥，始终失败）
// Deprecated: 推荐使用 UserAuthTokenWithConfig
func UserAuthToken() gin.HandlerFunc {
	return UserAuthTokenWithConfig("")
}
