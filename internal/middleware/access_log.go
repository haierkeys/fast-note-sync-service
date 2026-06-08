package middleware

import (
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AccessLogWithLogger creates access log middleware with logger (supports dependency injection)
// AccessLogWithLogger 创建带日志器的访问日志中间件（支持依赖注入）
func AccessLogWithLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		path := c.Request.URL.Path
		query := sanitizeQuery(c.Request.URL.RawQuery)

		startTime := time.Now()
		c.Next()

		timeCost := time.Since(startTime)

		logger.Info(path,
			zap.String("method", c.Request.Method),
			zap.String("url", path+"?"+query),
			zap.String("start-time", startTime.Format("2006-01-02 15:04:05")),
			zap.Duration("time-cost", timeCost),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	}
}

// sanitizeQuery masks sensitive parameters in query string
// sanitizeQuery 遮蔽请求查询字符串中的敏感参数值
func sanitizeQuery(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return rawQuery
	}
	sensitiveKeys := map[string]bool{
		"token":        true,
		"password":     true,
		"access_token": true,
		"share-token":  true,
	}
	changed := false
	for k := range values {
		lowerK := strings.ToLower(k)
		if sensitiveKeys[lowerK] {
			values.Set(k, "******")
			changed = true
		}
	}
	if changed {
		return values.Encode()
	}
	return rawQuery
}
