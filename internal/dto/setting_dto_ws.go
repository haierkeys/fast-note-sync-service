package dto

// SettingSyncModifyMessage message content for setting modification or creation during sync
// 同步期间配置修改或创建的消息内容
type SettingSyncModifyMessage struct {
	Vault            string `json:"vault" form:"vault"`               // Vault name // 保险库名称
	Path             string `json:"path" form:"path"`                 // Setting path // 配置路径
	PathHash         string `json:"pathHash" form:"pathHash"`         // Path hash // 路径哈希值
	Content          string `json:"content" form:"content"`           // Setting content // 配置内容
	ContentHash      string `json:"contentHash" form:"contentHash"`   // Content hash // 内容哈希
	Ctime            int64  `json:"ctime" form:"ctime"`               // Creation timestamp // 创建时间戳
	Mtime            int64  `json:"mtime" form:"mtime"`               // Modification timestamp // 修改时间戳
	UpdatedTimestamp int64  `json:"lastTime" form:"updatedTimestamp"` // Record update timestamp // 记录更新时间戳
}

// SettingSyncEndMessage defines the setting sync end message structure
// SettingSyncEndMessage 定义配置同步结束的消息结构
type SettingSyncEndMessage struct {
	LastTime           int64             `json:"lastTime" form:"lastTime"`                     // Last sync time // 最后同步时间
	NeedUploadCount    int64             `json:"needUploadCount" form:"needUploadCount"`       // Number of settings needing upload // 需要上传的数量
	NeedModifyCount    int64             `json:"needModifyCount" form:"needModifyCount"`       // Number of settings needing modification // 需要修改的数量
	NeedSyncMtimeCount int64             `json:"needSyncMtimeCount" form:"needSyncMtimeCount"` // Number of settings needing mtime sync // 需要同步修改时间的数量
	NeedDeleteCount    int64             `json:"needDeleteCount" form:"needDeleteCount"`       // Number of settings needing deletion // 需要删除的数量
	Messages           []WSQueuedMessage `json:"messages"`                                     // Merged message queue // 合并的消息队列
}

// SettingSyncNeedUploadMessage defines the message structure informing client that setting upload is needed during sync
// SettingSyncNeedUploadMessage 同步期间服务端通知客户端需要上传配置的消息结构
type SettingSyncNeedUploadMessage struct {
	Path string `json:"path" form:"path"` // Setting path // 配置路径
}

// SettingSyncMtimeMessage defines the message structure for setting modification time sync during sync
// SettingSyncMtimeMessage 同步期间配置元数据更新消息结构
type SettingSyncMtimeMessage struct {
	Path  string `json:"path" form:"path"`   // Setting path // 配置路径
	Ctime int64  `json:"ctime" form:"ctime"` // Creation timestamp // 创建时间戳
	Mtime int64  `json:"mtime" form:"mtime"` // Modification timestamp // 修改时间戳
}

// SettingSyncDeleteMessage defines the message structure for setting deletion during sync
// SettingSyncDeleteMessage 同步期间配置删除的消息结构
type SettingSyncDeleteMessage struct {
	Path string `json:"path" form:"path"` // Setting path // 配置路径
}
