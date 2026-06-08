package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// Cors creates CORS middleware
// Cors 创建跨域中间件
func Cors(allowedOrigins []string, extApiUrl string) gin.HandlerFunc {
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[strings.ToLower(strings.TrimRight(o, "/"))] = struct{}{}
	}

	// If allowedOrigins is empty, infer from extApiUrl as default origin
	// 如果 allowedOrigins 为空，则从 extApiUrl 中推断同源域作为默认允许的 Origin
	if len(allowedOrigins) == 0 && extApiUrl != "" {
		if u, err := url.Parse(extApiUrl); err == nil && u.Scheme != "" && u.Host != "" {
			inferred := u.Scheme + "://" + u.Host
			originSet[strings.ToLower(inferred)] = struct{}{}
		}
	}

	return func(c *gin.Context) {

		origin := c.GetHeader("Origin")
		allowedOrigin := ""
		isHealthCheck := c.Request.URL.Path == "/api/health"

		if isHealthCheck {
			allowedOrigin = "*"
		} else if origin != "" {
			norm := strings.ToLower(strings.TrimRight(origin, "/"))
			if _, ok := originSet[norm]; ok {
				allowedOrigin = origin
			} else if isBuiltinOrigin(origin) {
				allowedOrigin = origin
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, X-CSRF-Token, X-Client, X-Client-Name, X-Client-Version, X-Default-Vault-Name, AccessToken, Authorization, Debug, Domain, Token, Share-Token, Lang, Content-Type, Content-Length, Accept")

		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
			// When Access-Control-Allow-Origin is *, Access-Control-Allow-Credentials cannot be true
			// 当 Access-Control-Allow-Origin 为 * 时，Access-Control-Allow-Credentials 不能为 true
			if allowedOrigin != "*" {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		// Allow OPTIONS requests to pass
		// 允许放行OPTIONS请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

// isBuiltinOrigin checks if the origin is a builtin mobile application scheme
// isBuiltinOrigin 校验 origin 是否为内置的移动应用 Scheme
func isBuiltinOrigin(origin string) bool {
	return strings.HasPrefix(origin, "app://") || strings.HasPrefix(origin, "capacitor://")
}
