package middleware

import (
	"crypto/tls"

	"github.com/gin-gonic/gin"
)

// Proxy handles proxy headers and restores original request information
// Proxy 处理代理头部并恢复原始请求信息
func Proxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Detect protocol from X-Forwarded-Proto
		// 从 X-Forwarded-Proto 检测协议
		proto := c.GetHeader("X-Forwarded-Proto")
		if proto == "" {
			if c.Request.TLS != nil {
				proto = "https"
			} else {
				proto = "http"
			}
		}

		// Update Request URL Scheme
		// 更新请求 URL 的 Scheme
		c.Request.URL.Scheme = proto

		// If protocol is https but TLS is nil (due to proxy termination),
		// we "fake" a TLS state to satisfy libraries that check r.TLS != nil
		// 如果协议是 https 但 TLS 为 nil（由于代理终止），
		// 我们“伪造”一个 TLS 状态以满足检查 r.TLS != nil 的库
		if proto == "https" && c.Request.TLS == nil {
			c.Request.TLS = &tls.ConnectionState{}
		}

		// Detect host from X-Forwarded-Host
		// 从 X-Forwarded-Host 检测主机名
		if host := c.GetHeader("X-Forwarded-Host"); host != "" {
			c.Request.Host = host
		}

		c.Next()
	}
}
