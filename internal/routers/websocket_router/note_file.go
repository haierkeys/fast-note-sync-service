package websocket_router

import (
	"github.com/haierkeys/obsidian-better-sync-service/pkg/app"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/code"
)

func HandleUpload(c *app.WebsocketClient, msg *app.WebSocketMessage) {

	c.ToResponse(code.Success)
}
