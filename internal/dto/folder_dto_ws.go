package dto

// FolderSyncEndMessage defines the folder sync end message structure
// FolderSyncEndMessage 定义文件夹同步结束的消息结构
type FolderSyncEndMessage struct {
	LastTime        int64             `json:"lastTime"`        // Current sync update time // 本次同步更新时间
	NeedModifyCount int64             `json:"needModifyCount"` // Number of folders needing modification // 需要修改的文件夹数量
	NeedDeleteCount int64             `json:"needDeleteCount"` // Number of folders needing deletion // 需要删除的文件夹数量
	Messages        []WSQueuedMessage `json:"messages"`        // Merged message queue // 合并的消息队列
}

// FolderSyncRenameMessage message structure for folder rename during sync
// 同步过程中文件夹重命名的消息结构
type FolderSyncRenameMessage struct {
	Path        string `json:"path" form:"path" binding:"required"` // New path // 新路径
	PathHash    string `json:"pathHash" form:"pathHash"`            // New path hash // 新路径哈希
	Ctime       int64  `json:"ctime" form:"ctime"`                  // Creation timestamp // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime"`                  // Modification timestamp // 修改时间戳
	OldPath     string `json:"oldPath" form:"oldPath"`              // Old path // 旧路径
	OldPathHash string `json:"oldPathHash" form:"oldPathHash"`      // Old path hash // 旧路径哈希
}

// FolderSyncDeleteMessage message structure for folder deletion during sync
// FolderSyncDeleteMessage 同步期间文件夹删除的消息结构
type FolderSyncDeleteMessage struct {
	Path     string `json:"path" form:"path"`         // Folder path // 文件夹路径
	PathHash string `json:"pathHash" form:"pathHash"` // Path hash // 路径哈希值
	Ctime    int64  `json:"ctime" form:"ctime"`       // Creation timestamp // 创建时间戳
	Mtime    int64  `json:"mtime" form:"mtime"`       // Modification timestamp // 修改时间戳
}

// FolderSyncModifyMessage message content for folder modification or creation during sync
// 同步期间文件夹修改或创建的消息内容
type FolderSyncModifyMessage struct {
	Path             string `json:"path" form:"path"`                 // Folder path // 文件夹路径
	PathHash         string `json:"pathHash" form:"pathHash"`         // Path hash // 路径哈希值
	Ctime            int64  `json:"ctime" form:"ctime"`               // Creation timestamp // 创建时间戳
	Mtime            int64  `json:"mtime" form:"mtime"`               // Modification timestamp // 修改时间戳
	UpdatedTimestamp int64  `json:"lastTime" form:"updatedTimestamp"` // Record update timestamp // 记录更新时间戳
}
