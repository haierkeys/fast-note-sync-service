package middleware

import (
	"bytes"
	"time"

	"github.com/haierkeys/obsidian-better-sync-service/global"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AccessLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w AccessLogWriter) Write(p []byte) (int, error) {
	if n, err := w.body.Write(p); err != nil {
		return n, err
	}
	return w.ResponseWriter.Write(p)
}

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {

		bodyWriter := &AccessLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bodyWriter

		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		startTime := time.Now()
		c.Next()

		timeCost := time.Since(startTime)
		statusCode, isExists := c.Get("status_code")

		if isExists {
			global.Log().Info(path,
				zap.Int("status", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("fileurl", path),
				zap.String("query", query),
				zap.String("ip", c.ClientIP()),
				zap.String("start-time", startTime.Format("2006-01-02 15:04:05")),
				zap.String("user-agent", c.Request.UserAgent()),
				zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
				zap.Duration("time-cost", timeCost),
				zap.Int("status_code", statusCode.(int)),
				zap.String("request", c.Request.PostForm.Encode()),
				zap.String("response", bodyWriter.body.String()),
			)
		} else {
			global.Log().Info(path,
				zap.Int("status", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("fileurl", path),
				zap.String("query", query),
				zap.String("ip", c.ClientIP()),
				zap.String("start-time", startTime.Format("2006-01-02 15:04:05")),
				zap.String("user-agent", c.Request.UserAgent()),
				zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
				zap.Duration("time-cost", timeCost),
				zap.String("request", c.Request.PostForm.Encode()),
				zap.String("response", bodyWriter.body.String()),
			)
		}

	}
}
