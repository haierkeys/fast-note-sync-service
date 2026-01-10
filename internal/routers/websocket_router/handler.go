// Package websocket_router 提供 WebSocket 路由处理器
package websocket_router

import (
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"go.uber.org/zap"
)

// WSHandler WebSocket 基础 Handler 结构体，封装 App Container
// 所有 WebSocket Handler 都应该嵌入此结构体以获得依赖注入能力
type WSHandler struct {
	App *app.App
}

// NewWSHandler 创建 WebSocket 基础 Handler 实例
func NewWSHandler(a *app.App) *WSHandler {
	return &WSHandler{App: a}
}

// logError 记录错误日志，包含 Trace ID
// 直接使用 WebsocketClient.TraceID 字段，避免从可能失效的 HTTP context 获取
func (h *WSHandler) logError(c *pkgapp.WebsocketClient, method string, err error) {
	traceID := ""
	if c != nil {
		traceID = c.TraceID
	}
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}

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

// respondError 统一错误响应方法
// 记录错误日志并发送包含 Details 的错误响应给客户端
func (h *WSHandler) respondError(c *pkgapp.WebsocketClient, codeErr *code.Code, err error, method string) {
	h.logError(c, method, err)
	c.ToResponse(codeErr.WithDetails(err.Error()))
}

// respondErrorWithData 带数据的统一错误响应方法
// 记录错误日志并发送包含 Details 和 Data 的错误响应给客户端
func (h *WSHandler) respondErrorWithData(c *pkgapp.WebsocketClient, codeErr *code.Code, err error, data interface{}, method string) {
	h.logError(c, method, err)
	c.ToResponse(codeErr.WithDetails(err.Error()).WithData(data))
}

// GetTraceID 从 WebSocket 客户端获取 Trace ID
// 直接使用 WebsocketClient.TraceID 字段，避免从可能失效的 HTTP context 获取
func GetTraceID(c *pkgapp.WebsocketClient) string {
	if c == nil {
		return ""
	}
	return c.TraceID
}

// LogErrorWithLogger 记录错误日志，包含 Trace ID（使用注入的 logger）
func LogErrorWithLogger(logger *zap.Logger, c *pkgapp.WebsocketClient, method string, err error) {
	traceID := GetTraceID(c)
	logger.Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}

// LogInfoWithLogger 记录信息日志，包含 Trace ID（使用注入的 logger）
func LogInfoWithLogger(logger *zap.Logger, c *pkgapp.WebsocketClient, method string, fields ...zap.Field) {
	traceID := GetTraceID(c)
	allFields := append([]zap.Field{zap.String("traceId", traceID)}, fields...)
	logger.Info(method, allFields...)
}

// LogWarnWithLogger 记录警告日志，包含 Trace ID（使用注入的 logger）
func LogWarnWithLogger(logger *zap.Logger, c *pkgapp.WebsocketClient, method string, fields ...zap.Field) {
	traceID := GetTraceID(c)
	allFields := append([]zap.Field{zap.String("traceId", traceID)}, fields...)
	logger.Warn(method, allFields...)
}
