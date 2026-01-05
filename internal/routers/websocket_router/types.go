package websocket_router

import "github.com/haierkeys/fast-note-sync-service/pkg/code"

// queuedMessage 表示待发送的消息队列项
// 用于在同步过程中收集消息,在 SyncEnd 消息发送后统一批量发送
type queuedMessage struct {
	response    *code.Code
	messageType string
}
