package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AccessLogWithLogger 创建带日志器的访问日志中间件（支持依赖注入）
func AccessLogWithLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

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
