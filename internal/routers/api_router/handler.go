// Package api_router 提供 HTTP API 路由处理器
package api_router

import (
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
)

// Handler 基础 Handler 结构体，封装 App Container
// 所有 API Handler 都应该嵌入此结构体以获得依赖注入能力
type Handler struct {
	App *app.App
	WSS *pkgapp.WebsocketServer
}

// NewHandler 创建基础 Handler 实例
func NewHandler(a *app.App) *Handler {
	return &Handler{App: a}
}

// NewHandlerWithWSS 创建带 WebSocket 服务的 Handler 实例
func NewHandlerWithWSS(a *app.App, wss *pkgapp.WebsocketServer) *Handler {
	return &Handler{App: a, WSS: wss}
}
