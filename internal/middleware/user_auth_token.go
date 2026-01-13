package middleware

import (
	"strings"

	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
)

// UserAuthTokenWithConfig 用户 Token 认证中间件（使用注入的密钥）
// 支持 Authorization: Bearer <token> 格式和 URL 参数 token（用于图片等资源请求）
func UserAuthTokenWithConfig(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := app.NewResponse(c)
		var token string

		// 优先从 Authorization header 获取（标准 Bearer 格式）
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 如果 header 中没有，尝试从 URL 参数获取（用于图片等资源请求）
		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			response.ToResponse(code.ErrorNotUserAuthToken)
			c.Abort()
			return
		}

		user, err := app.ParseTokenWithKey(token, secretKey)
		if err != nil {
			response.ToResponse(code.ErrorInvalidUserAuthToken)
			c.Abort()
			return
		}

		c.Set("user_token", user)
		c.Next()
	}
}

// UserAuthToken 用户 Token 认证中间件（无密钥，始终失败）
// Deprecated: 推荐使用 UserAuthTokenWithConfig
func UserAuthToken() gin.HandlerFunc {
	return UserAuthTokenWithConfig("")
}
