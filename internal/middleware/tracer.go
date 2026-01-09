package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/global"
)

const (
	// DefaultTraceIDHeader 默认的 Trace ID 请求头名称
	DefaultTraceIDHeader = "X-Trace-ID"
	// TraceIDKey Context 中存储 Trace ID 的键
	TraceIDKey = "trace_id"
)

// TraceMiddleware 创建请求追踪中间件
// 功能：
// 1. 从请求头获取或生成唯一的 Trace ID
// 2. 将 Trace ID 注入到 gin.Context 和 request.Context
// 3. 在响应头中返回 Trace ID
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用追踪
		if global.Config != nil && !global.Config.Tracer.Enabled {
			c.Next()
			return
		}

		// 获取配置的请求头名称
		headerName := getTraceIDHeader()

		// 尝试从请求头获取 Trace ID
		traceID := c.GetHeader(headerName)
		if traceID == "" {
			// 生成新的 Trace ID
			traceID = generateTraceID()
		}

		// 存储到 gin.Context
		c.Set(TraceIDKey, traceID)

		// 注入到 request.Context
		ctx := context.WithValue(c.Request.Context(), TraceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)

		// 添加到响应头
		c.Header(headerName, traceID)

		c.Next()
	}
}

// generateTraceID 生成唯一的 Trace ID
// 格式: {timestamp_nano}-{random_hex}
func generateTraceID() string {
	// 生成 8 字节随机数
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		// 如果随机数生成失败，使用时间戳作为后备
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return fmt.Sprintf("%d-%s",
		time.Now().UnixNano(),
		hex.EncodeToString(randomBytes)[:8])
}

// getTraceIDHeader 获取配置的 Trace ID 请求头名称
func getTraceIDHeader() string {
	if global.Config != nil && global.Config.Tracer.Header != "" {
		return global.Config.Tracer.Header
	}
	return DefaultTraceIDHeader
}

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
