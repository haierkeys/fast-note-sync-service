package websocket_router

import (
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

// WebGUIWSHandler WebGUI WebSocket 处理器
type WebGUIWSHandler struct {
	App *app.App
}

// NewWebGUIWSHandler 创建 WebGUIWSHandler 实例
func NewWebGUIWSHandler(a *app.App) *WebGUIWSHandler {
	return &WebGUIWSHandler{App: a}
}

// WebGUIConfigGet 处理获取 WebGUI 配置的 WebSocket 消息
func (h *WebGUIWSHandler) WebGUIConfigGet(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	cfg := h.App.Config()
	c.ToResponse(code.Success.WithData(cfg.WebGUI), "WebGUIConfigGet")
}
