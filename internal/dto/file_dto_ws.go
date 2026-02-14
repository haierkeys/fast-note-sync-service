package dto

// FileSyncModifyMessage message content for file modification or creation
// FileSyncModifyMessage 文件修改或创建的消息内容
type FileSyncModifyMessage struct {
	Path             string `json:"path" form:"path"`                 // File path // 文件路径
	PathHash         string `json:"pathHash" form:"pathHash"`         // Path hash // 路径哈希值
	ContentHash      string `json:"contentHash" form:"contentHash"`   // Content hash // 内容哈希
	Size             int64  `json:"size" form:"size"`                 // File size // 文件大小
	Ctime            int64  `json:"ctime" form:"ctime"`               // Creation timestamp // 创建时间戳
	Mtime            int64  `json:"mtime" form:"mtime"`               // Modification timestamp // 修改时间戳
	UpdatedTimestamp int64  `json:"lastTime" form:"updatedTimestamp"` // Record update timestamp // 记录更新时间戳
}

// FileSyncEndMessage defines the message structure when file sync ends
// FileSyncEndMessage 定义文件同步结束时的消息结构
type FileSyncEndMessage struct {
	LastTime           int64             `json:"lastTime" form:"lastTime"`                     // Last sync time // 最后同步时间
	NeedUploadCount    int64             `json:"needUploadCount" form:"needUploadCount"`       // Number of items needing upload // 需要上传的数量
	NeedModifyCount    int64             `json:"needModifyCount" form:"needModifyCount"`       // Number of items needing modification // 需要修改的数量
	NeedSyncMtimeCount int64             `json:"needSyncMtimeCount" form:"needSyncMtimeCount"` // Number of items needing mtime sync // 需要同步修改时间的数量
	NeedDeleteCount    int64             `json:"needDeleteCount" form:"needDeleteCount"`       // Number of items needing deletion // 需要删除的数量
	Messages           []WSQueuedMessage `json:"messages"`                                     // Merged message queue // 合并的消息队列
}

// FileSyncUploadMessage defines the message structure informing client that file upload is needed
// FileSyncUploadMessage 定义服务端通知客户端需要上传文件的消息结构
type FileSyncUploadMessage struct {
	Path      string `json:"path"`      // File path // 文件路径
	SessionID string `json:"sessionId"` // Session ID // 会话 ID
	ChunkSize int64  `json:"chunkSize"` // Chunk size // 分块大小
}

// FileSyncDownloadMessage defines the message structure informing client that file download is ready
// FileSyncDownloadMessage 定义服务端通知客户端准备下载文件的消息结构
type FileSyncDownloadMessage struct {
	Path        string `json:"path"`        // File path // 文件路径
	Ctime       int64  `json:"ctime"`       // Creation time // 创建时间
	Mtime       int64  `json:"mtime"`       // Modification time // 修改时间
	SessionID   string `json:"sessionId"`   // Session ID // 会话 ID
	ChunkSize   int64  `json:"chunkSize"`   // Chunk size // 分块大小
	TotalChunks int64  `json:"totalChunks"` // Total chunks // 总分块数
	Size        int64  `json:"size"`        // Total file size // 文件总大小
}

// FileSyncMtimeMessage defines the message structure for file metadata update
// FileSyncMtimeMessage 定义文件元数据更新消息结构
type FileSyncMtimeMessage struct {
	Path  string `json:"path"`   // File path // 文件路径
	Ctime int64  `json:"ctime" ` // Creation timestamp // 创建时间戳
	Mtime int64  `json:"mtime" ` // Modification timestamp // 修改时间戳
}

// FileSyncDeleteMessage defines the message structure for file deletion during sync
// FileSyncDeleteMessage 定义同步期间文件删除的消息结构
type FileSyncDeleteMessage struct {
	Path     string `json:"path" form:"path"`         // File path // 文件路径
	PathHash string `json:"pathHash" form:"pathHash"` // Path hash // 路径哈希值
	Ctime    int64  `json:"ctime" form:"ctime"`       // Creation timestamp // 创建时间戳
	Mtime    int64  `json:"mtime" form:"mtime"`       // Modification timestamp // 修改时间戳
	Size     int64  `json:"size" form:"size"`         // File size // 文件大小
}

// FileSyncRenameMessage message structure for file rename during sync
// 同步过程中文件重命名的消息结构
type FileSyncRenameMessage struct {
	Path        string `json:"path" form:"path" binding:"required"` // New path // 新路径
	PathHash    string `json:"pathHash" form:"pathHash"`            // New path hash // 新路径哈希
	ContentHash string `json:"contentHash" form:"contentHash"`      // Content hash // 内容哈希
	Ctime       int64  `json:"ctime" form:"ctime"`                  // Creation timestamp // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime"`                  // Modification timestamp // 修改时间戳
	Size        int64  `json:"size" form:"size"`                    // File size // 文件大小
	OldPath     string `json:"oldPath" form:"oldPath"`              // Old path // 旧路径
	OldPathHash string `json:"oldPathHash" form:"oldPathHash"`      // Old path hash // 旧路径哈希
}
