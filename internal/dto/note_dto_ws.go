package dto

// NoteSyncRenameMessage message structure for note rename during sync
// 同步过程中笔记重命名的消息结构
type NoteSyncRenameMessage struct {
	Path        string `json:"path" form:"path" binding:"required"` // New path // 新路径
	PathHash    string `json:"pathHash" form:"pathHash"`            // New path hash // 新路径哈希
	ContentHash string `json:"contentHash" form:"contentHash"`      // Content hash // 内容哈希
	Ctime       int64  `json:"ctime" form:"ctime"`                  // Creation timestamp // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime"`                  // Modification timestamp // 修改时间戳
	Size        int64  `json:"size" form:"size"`                    // File size // 文件大小
	OldPath     string `json:"oldPath" form:"oldPath"`              // Old path // 旧路径
	OldPathHash string `json:"oldPathHash" form:"oldPathHash"`      // Old path hash // 旧路径哈希
}

// NoteSyncModifyMessage message content for note modification or creation
// NoteSyncModifyMessage 笔记修改或创建的消息内容
type NoteSyncModifyMessage struct {
	Path             string `json:"path" form:"path"`                 // Note path // 笔记路径
	PathHash         string `json:"pathHash" form:"pathHash"`         // Path hash // 路径哈希值
	Content          string `json:"content" form:"content"`           // Note content // 笔记内容
	ContentHash      string `json:"contentHash" form:"contentHash"`   // Content hash // 内容哈希
	Ctime            int64  `json:"ctime" form:"ctime"`               // Creation timestamp // 创建时间戳
	Mtime            int64  `json:"mtime" form:"mtime"`               // Modification timestamp // 修改时间戳
	UpdatedTimestamp int64  `json:"lastTime" form:"updatedTimestamp"` // Record update timestamp // 记录更新时间戳
}

// NoteSyncEndMessage message structure returned when sync ends
// NoteSyncEndMessage 同步结束时返回的信息结构
type NoteSyncEndMessage struct {
	LastTime           int64             `json:"lastTime" form:"lastTime"`                     // Current sync update time // 本次同步更新时间
	NeedUploadCount    int64             `json:"needUploadCount" form:"needUploadCount"`       // Number of notes needing upload // 需要上传的笔记数量
	NeedModifyCount    int64             `json:"needModifyCount" form:"needModifyCount"`       // Number of notes needing modification // 需要修改的数量
	NeedSyncMtimeCount int64             `json:"needSyncMtimeCount" form:"needSyncMtimeCount"` // Number of notes needing mtime sync // 需要同步修改时间的数量
	NeedDeleteCount    int64             `json:"needDeleteCount" form:"needDeleteCount"`       // Number of notes needing deletion // 需要删除的数量
	Messages           []WSQueuedMessage `json:"messages"`                                     // Merged message queue // 合并的消息队列
}

// NoteSyncNeedPushMessage server informs client of file info needing push
// NoteSyncNeedPushMessage 服务端告知客户端需要推送的文件信息
type NoteSyncNeedPushMessage struct {
	Path     string `json:"path" form:"path"`         // Note path // 笔记路径
	PathHash string `json:"pathHash" form:"pathHash"` // Path hash // 路径哈希值
}

// NoteSyncMtimeMessage message structure for updating mtime during sync
// NoteSyncMtimeMessage 同步时用于更新 mtime 的消息结构
type NoteSyncMtimeMessage struct {
	Path  string `json:"path" form:"path"`   // Note path // 笔记路径
	Ctime int64  `json:"ctime" form:"ctime"` // Creation timestamp // 创建时间戳
	Mtime int64  `json:"mtime" form:"mtime"` // Modification timestamp // 修改时间戳
}

// NoteSyncDeleteMessage message structure for note deletion
// NoteSyncDeleteMessage 笔记删除的消息结构
type NoteSyncDeleteMessage struct {
	Path     string `json:"path" form:"path"`         // Note path // 笔记路径
	PathHash string `json:"pathHash" form:"pathHash"` // Path hash // 路径哈希值
	Ctime    int64  `json:"ctime" form:"ctime"`       // Creation timestamp // 创建时间戳
	Mtime    int64  `json:"mtime" form:"mtime"`       // Modification timestamp // 修改时间戳
	Size     int64  `json:"size" form:"size"`         // File size // 文件大小
}
