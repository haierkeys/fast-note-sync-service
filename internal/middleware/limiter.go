package middleware

import (
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/limiter"

	"github.com/gin-gonic/gin"
)

// RateLimiter creates rate limiting middleware (supports dependency injection)
// RateLimiter 创建限流中间件（支持依赖注入）
func RateLimiter(l limiter.Face) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := l.Key(c)
		if bucket, ok := l.GetBucket(key); ok {
			count := bucket.TakeAvailable(1)
			if count == 0 {
				response := app.NewResponse(c)
				response.ToResponse(code.ErrorTooManyRequests)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
