package websocket_router

// queuedMessage represents a message item to be sent
// Used to collect messages during sync process, and sent together in SyncEnd message
// queuedMessage 表示待发送的消息项
// 用于在同步过程中收集消息,在 SyncEnd 消息中统一合并发送
type queuedMessage struct {
	Action string `json:"action"` // Message type // 消息类型
	Data   any    `json:"data"`   // Message content // 消息内容
}
