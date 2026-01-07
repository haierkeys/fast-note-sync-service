package websocket_router

import (
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

// WebGUIConfigGet 处理获取 WebGUI 配置的 WebSocket 消息
func WebGUIConfigGet(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	c.ToResponse(code.Success.Clone().WithData(global.Config.WebGUI), "WebGUIConfigGet")
}
