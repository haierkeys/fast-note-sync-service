package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryWithLogger 创建带日志器的 Recovery 中间件（支持依赖注入）
func RecoveryWithLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		defer func() {
			if err := recover(); err != nil {
				var errorMsg string
				switch err.(type) {
				case string:
					errorMsg = err.(string)
				case error:
					// 记录 error 类型的错误
					logger.Error("Recovered from panic",
						zap.Int("status", c.Writer.Status()),
						zap.String("router", path),
						zap.String("method", c.Request.Method),
						zap.String("query", query),
						zap.String("ip", c.ClientIP()),
						zap.String("user-agent", c.Request.UserAgent()),
						zap.String("request", c.Request.PostForm.Encode()),
						zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()), // 记录错误的上下文
						zap.Error(err.(error)),                                               // 错误信息
						zap.String("stack", string(debug.Stack())),                           // 错误堆栈
					)
					errorMsg = err.(error).Error()
				default:
					// 如果是其它类型的 panic（如非错误类型的 panic）
					logger.Error("Recovered from unknown panic",
						zap.Int("status", c.Writer.Status()),
						zap.String("router", path),
						zap.String("method", c.Request.Method),
						zap.String("query", query),
						zap.String("ip", c.ClientIP()),
						zap.String("user-agent", c.Request.UserAgent()),
						zap.String("request", c.Request.PostForm.Encode()),
						zap.String("panic_value", fmt.Sprintf("%v", err)), // 记录 panic 的值
						zap.String("stack", string(debug.Stack())),        // 错误堆栈
					)
				}

				// 返回统一的错误响应
				app.NewResponse(c).ToResponse(code.ErrorServerInternal.WithDetails(errorMsg))
			}
		}()

		c.Next()
	}
}
