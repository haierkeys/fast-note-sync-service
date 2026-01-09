// Package websocket_router 提供 WebSocket 路由处理器
package websocket_router

import (
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
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
	h.App.Logger.Error(method,
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
	h.App.Logger.Info(method, allFields...)
}

// GetTraceID 从 WebSocket 客户端获取 Trace ID
// 直接使用 WebsocketClient.TraceID 字段，避免从可能失效的 HTTP context 获取
func GetTraceID(c *pkgapp.WebsocketClient) string {
	if c == nil {
		return ""
	}
	return c.TraceID
}

// ============================================
// 辅助函数：为现有的函数式 handlers 提供 Trace ID 支持
// ============================================

// LogErrorWithTrace 记录错误日志，包含 Trace ID（用于函数式 handlers）
func LogErrorWithTrace(c *pkgapp.WebsocketClient, method string, err error) {
	traceID := GetTraceID(c)
	global.Logger.Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}

// LogInfoWithTrace 记录信息日志，包含 Trace ID（用于函数式 handlers）
func LogInfoWithTrace(c *pkgapp.WebsocketClient, method string, fields ...zap.Field) {
	traceID := GetTraceID(c)
	allFields := append([]zap.Field{zap.String("traceId", traceID)}, fields...)
	global.Logger.Info(method, allFields...)
}

// LogWarnWithTrace 记录警告日志，包含 Trace ID（用于函数式 handlers）
func LogWarnWithTrace(c *pkgapp.WebsocketClient, method string, fields ...zap.Field) {
	traceID := GetTraceID(c)
	allFields := append([]zap.Field{zap.String("traceId", traceID)}, fields...)
	global.Logger.Warn(method, allFields...)
}
