package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Cors creates CORS middleware
// Cors 创建跨域中间件
func Cors() gin.HandlerFunc {

	return func(c *gin.Context) {

		origin := c.GetHeader("Origin")
		allowedOrigin := ""
		if origin != "" {
			if strings.HasPrefix(origin, "app://") ||
				strings.HasPrefix(origin, "capacitor://") ||
				strings.HasPrefix(origin, "http://") ||
				strings.HasPrefix(origin, "https://") {
				allowedOrigin = origin
			}
		}

		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, X-CSRF-Token, X-Client, X-Client-Name, X-Client-Version, X-Default-Vault-Name, AccessToken, Authorization, Debug, Domain, Token, Share-Token, Lang, Content-Type, Content-Length, Accept")

		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
		}

		// Allow OPTIONS requests to pass
		// 允许放行OPTIONS请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}
