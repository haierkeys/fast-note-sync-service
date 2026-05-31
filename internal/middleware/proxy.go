package middleware

import (
	"crypto/tls"
	"net"

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

		// Trust CF-Connecting-IP only when the direct TCP connection originates from
		// the loopback interface (127.0.0.1 / ::1). This is the fingerprint of a local
		// cloudflared tunnel process — the only entity that connects from loopback and
		// injects this header. An external attacker connecting directly to the origin
		// cannot fake a loopback RemoteAddr at the TCP level.
		//
		// 仅当 TCP 直连来源为回环地址（127.0.0.1 / ::1）时，才信任 CF-Connecting-IP 头。
		// 这是本地 cloudflared 进程的唯一特征——它是唯一从回环地址连接并注入该头的实体。
		// 外部攻击者直连源站时，无法在 TCP 层伪造回环地址，因此此头会被忽略。
		if cfIP := c.GetHeader("CF-Connecting-IP"); cfIP != "" {
			if remoteHost, _, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil {
				if remoteHost == "127.0.0.1" || remoteHost == "::1" {
					c.Request.RemoteAddr = cfIP + ":0"
				}
			}
		}

		c.Next()
	}
}
