package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultTraceIDHeader default Trace ID request header name // 默认的 Trace ID 请求头名称
	DefaultTraceIDHeader = "X-Trace-ID"
	// TraceIDKey Context 中存储 Trace ID 的键
	// TraceIDKey key for storing Trace ID in Context
	TraceIDKey = "trace_id"
)

// TraceMiddlewareWithConfig creates a request tracing middleware (with injected configuration)
// TraceMiddlewareWithConfig 创建请求追踪中间件（使用注入的配置）
// Functionality:
// 功能：
// 1. Get or generate a unique Trace ID from the request header
// 1. 从请求头获取或生成唯一的 Trace ID
// 2. Inject Trace ID into gin.Context and request.Context
// 2. 将 Trace ID 注入到 gin.Context 和 request.Context
// 3. Return Trace ID in the response header
// 3. 在响应头中返回 Trace ID
func TraceMiddlewareWithConfig(enabled bool, headerName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if tracing is enabled
		// 检查是否启用追踪
		if !enabled {
			c.Next()
			return
		}

		// Get configured request header name
		// 获取配置的请求头名称
		if headerName == "" {
			headerName = DefaultTraceIDHeader
		}

		// Try to get Trace ID from request header
		// 尝试从请求头获取 Trace ID
		traceID := c.GetHeader(headerName)
		if traceID == "" {
			// Generate new Trace ID
			// 生成新的 Trace ID
			traceID = generateTraceID()
		}

		// Store into gin.Context
		// 存储到 gin.Context
		c.Set(TraceIDKey, traceID)

		// Inject into request.Context
		// 注入到 request.Context
		ctx := context.WithValue(c.Request.Context(), TraceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)

		// Add to response header
		// 添加到响应头
		c.Header(headerName, traceID)

		c.Next()
	}
}

// TraceMiddleware creates request tracing middleware (enabled by default)
// TraceMiddleware 创建请求追踪中间件（默认启用）
// Deprecated: Recommended to use TraceMiddlewareWithConfig
// Deprecated: 推荐使用 TraceMiddlewareWithConfig
func TraceMiddleware() gin.HandlerFunc {
	return TraceMiddlewareWithConfig(true, DefaultTraceIDHeader)
}

// generateTraceID generates unique Trace ID
// Format: {timestamp_nano}-{random_hex}
// generateTraceID 生成唯一的 Trace ID
// 格式: {timestamp_nano}-{random_hex}
func generateTraceID() string {
	// Generate 8-byte random number
	// 生成 8 字节随机数
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		// If random number generation fails, use timestamp as backup
		// 如果随机数生成失败，使用时间戳作为后备
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return fmt.Sprintf("%d-%s",
		time.Now().UnixNano(),
		hex.EncodeToString(randomBytes)[:8])
}

// GetTraceID retrieves Trace ID from context.Context
// GetTraceID 从 context.Context 获取 Trace ID
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(TraceIDKey).(string); ok {
		return id
	}
	return ""
}

// GetTraceIDFromGin retrieves Trace ID from gin.Context
// GetTraceIDFromGin 从 gin.Context 获取 Trace ID
func GetTraceIDFromGin(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if id, exists := c.Get(TraceIDKey); exists {
		if traceID, ok := id.(string); ok {
			return traceID
		}
	}
	return ""
}
