package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryWithLogger creates recovery middleware with logger (supports dependency injection)
// RecoveryWithLogger 创建带日志器的恢复中间件（支持依赖注入）
func RecoveryWithLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		defer func() {
			if err := recover(); err != nil {
				var errorMsg string
				switch val := err.(type) {
				case string:
					errorMsg = val
				case error:
					// Record error type errors
					// 记录 error 类型的错误
					logger.Error("Recovered from panic",
						zap.Int("status", c.Writer.Status()),
						zap.String("router", path),
						zap.String("method", c.Request.Method),
						zap.String("query", query),
						zap.String("ip", c.ClientIP()),
						zap.String("user-agent", c.Request.UserAgent()),
						zap.String("request", c.Request.PostForm.Encode()),
						zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()), // Record error context
						// 记录错误的上下文
						zap.Error(val), // Error info
						zap.String("stack", string(debug.Stack())), // Error stack
						// 错误堆栈
					)
					errorMsg = val.Error()
				default:
					// Other types of panic (non-error type panic)
					// 如果是其它类型的 panic（如非错误类型的 panic）
					logger.Error("Recovered from unknown panic",
						zap.Int("status", c.Writer.Status()),
						zap.String("router", path),
						zap.String("method", c.Request.Method),
						zap.String("query", query),
						zap.String("ip", c.ClientIP()),
						zap.String("user-agent", c.Request.UserAgent()),
						zap.String("request", c.Request.PostForm.Encode()),
						zap.String("panic_value", fmt.Sprintf("%v", val)), // Record panic value
						// 记录 panic 的值
						zap.String("stack", string(debug.Stack())), // Error stack
						// 错误堆栈
					)
				}

				// Return unified error response
				// 返回统一的错误响应
				app.NewResponse(c).ToResponse(code.ErrorServerInternal.WithDetails(errorMsg))
			}
		}()

		c.Next()
	}
}
