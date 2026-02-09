// Package websocket_router provides WebSocket router handlers
// Package websocket_router 提供 WebSocket 路由处理器
package websocket_router

import (
	"context"
	"strings"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"go.uber.org/zap"
)

// WSHandler WebSocket base Handler struct, encapsulates App Container
// All WebSocket Handlers should embed this struct to gain dependency injection capability
// WSHandler WebSocket 基础 Handler 结构体，封装 App Container
// 所有 WebSocket Handler 都应该嵌入此结构体以获得依赖注入能力
type WSHandler struct {
	App *app.App
}

// NewWSHandler creates WebSocket base Handler instance
// NewWSHandler 创建 WebSocket 基础 Handler 实例
func NewWSHandler(a *app.App) *WSHandler {
	return &WSHandler{App: a}
}

// logError records error log, including Trace ID
// Directly use WebsocketClient.TraceID field, avoiding fetching from potentially invalid HTTP context
// logError 记录错误日志，包含 Trace ID
// 直接使用 WebsocketClient.TraceID 字段，避免从可能失效的 HTTP context 获取
func (h *WSHandler) logError(c *pkgapp.WebsocketClient, method string, err error) {
	traceID := ""
	if c != nil {
		traceID = c.TraceID
	}

	// If connection closed error and context canceled, downgrade log level
	// 如果是连接关闭导致的错误且 context 已取消，降级日志级别
	if isNetworkClosedError(err) && c != nil && c.Context().Err() != nil {
		h.logDebug(c, method, zap.Error(err))
		return
	}

	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}

// logDebug records debug log, including Trace ID
// logDebug 记录调试日志，包含 Trace ID
func (h *WSHandler) logDebug(c *pkgapp.WebsocketClient, method string, fields ...zap.Field) {
	traceID := ""
	if c != nil {
		traceID = c.TraceID
	}
	allFields := append([]zap.Field{zap.String("traceId", traceID)}, fields...)
	h.App.Logger().Debug(method, allFields...)
}

// logInfo records info log, including Trace ID
// Directly use WebsocketClient.TraceID field, avoiding fetching from potentially invalid HTTP context
// logInfo 记录信息日志，包含 Trace ID
// 直接使用 WebsocketClient.TraceID 字段，避免从可能失效的 HTTP context 获取
func (h *WSHandler) logInfo(c *pkgapp.WebsocketClient, method string, fields ...zap.Field) {
	traceID := ""
	if c != nil {
		traceID = c.TraceID
	}
	allFields := append([]zap.Field{zap.String("traceId", traceID)}, fields...)
	h.App.Logger().Info(method, allFields...)
}

// logWarn records warning log, including Trace ID
// Directly use WebsocketClient.TraceID field, avoiding fetching from potentially invalid HTTP context
// logWarn 记录警告日志，包含 Trace ID
// 直接使用 WebsocketClient.TraceID 字段，避免从可能失效的 HTTP context 获取
func (h *WSHandler) logWarn(c *pkgapp.WebsocketClient, method string, fields ...zap.Field) {
	traceID := ""
	if c != nil {
		traceID = c.TraceID
	}
	allFields := append([]zap.Field{zap.String("traceId", traceID)}, fields...)
	h.App.Logger().Warn(method, allFields...)
}

// respondError unified error response method
// Records error log and sends error response with Details to client
// respondError 统一错误响应方法
// 记录错误日志并发送包含 Details 的错误响应给客户端
func (h *WSHandler) respondError(c *pkgapp.WebsocketClient, codeErr *code.Code, err error, method string) {
	h.logError(c, method, err)
	// If err is already a *code.Code, it might have more details than codeErr
	if cErr, ok := err.(*code.Code); ok {
		c.ToResponse(cErr)
		return
	}
	c.ToResponse(codeErr.WithDetails(err.Error()))
}

// respondErrorWithData unified error response method with data
// Records error log and sends error response with Details and Data to client
// respondErrorWithData 带数据的统一错误响应方法
// 记录错误日志并发送包含 Details 和 Data 的错误响应给客户端
func (h *WSHandler) respondErrorWithData(c *pkgapp.WebsocketClient, codeErr *code.Code, err error, data interface{}, method string) {
	h.logError(c, method, err)
	c.ToResponse(codeErr.WithDetails(err.Error()).WithData(data))
}

// GetTraceID retrieves Trace ID from WebSocket client
// Directly use WebsocketClient.TraceID field, avoiding fetching from potentially invalid HTTP context
// GetTraceID 从 WebSocket 客户端获取 Trace ID
// 直接使用 WebsocketClient.TraceID 字段，避免从可能失效的 HTTP context 获取
func GetTraceID(c *pkgapp.WebsocketClient) string {
	if c == nil {
		return ""
	}
	return c.TraceID
}

// LogErrorWithLogger records error log, including Trace ID (uses injected logger)
// LogErrorWithLogger 记录错误日志，包含 Trace ID（使用注入的 logger）
func LogErrorWithLogger(logger *zap.Logger, c *pkgapp.WebsocketClient, method string, err error) {
	traceID := GetTraceID(c)

	// If connection closed error and context canceled, downgrade log level
	// 如果是连接关闭导致的错误且 context 已取消，降级日志级别
	if isNetworkClosedError(err) && c != nil && c.Context().Err() != nil {
		allFields := []zap.Field{zap.String("traceId", traceID), zap.Error(err)}
		logger.Debug(method, allFields...)
		return
	}

	logger.Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}

// LogInfoWithLogger records info log, including Trace ID (uses injected logger)
// LogInfoWithLogger 记录信息日志，包含 Trace ID（使用注入的 logger）
func LogInfoWithLogger(logger *zap.Logger, c *pkgapp.WebsocketClient, method string, fields ...zap.Field) {
	traceID := GetTraceID(c)
	allFields := append([]zap.Field{zap.String("traceId", traceID)}, fields...)
	logger.Info(method, allFields...)
}

// LogWarnWithLogger records warning log, including Trace ID (uses injected logger)
// LogWarnWithLogger 记录警告日志，包含 Trace ID（使用注入的 logger）
func LogWarnWithLogger(logger *zap.Logger, c *pkgapp.WebsocketClient, method string, fields ...zap.Field) {
	traceID := GetTraceID(c)
	allFields := append([]zap.Field{zap.String("traceId", traceID)}, fields...)
	logger.Warn(method, allFields...)
}

// isNetworkClosedError checks if the error is related to network closure
// isNetworkClosedError 检查是否为网络关闭相关的错误
func isNetworkClosedError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "use of closed network connection") ||
		strings.Contains(msg, "connection reset by peer") ||
		strings.Contains(msg, "broken pipe") ||
		err == context.Canceled
}
