package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {

	return func(c *gin.Context) {

		var domain string
		if s, exist := c.GetQuery("domain"); exist {
			domain = s
		} else {
			domain = c.GetHeader("domain")
		}

		if domain != "" && !strings.HasPrefix(domain, "http"+"://") {
			xForwardedProto := c.GetHeader("X-Forwarded-Proto")
			if xForwardedProto == "https" {
				domain = "https" + "://" + domain
			} else {
				domain = "http" + "://" + domain
			}
		}

		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With,  AccessToken, X-CSRF-Token, Authorization, Debug, Domain, Token, Lang, Content-Type, Content-Length,  Accept")

		if domain != "" {
			c.Header("Access-Control-Allow-Origin", domain)
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		// 允许放行OPTIONS请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}
