package dto

// FolderSyncEndMessage defines the folder sync end message structure
// FolderSyncEndMessage 定义文件夹同步结束的消息结构。
type FolderSyncEndMessage struct {
	LastTime        int64             `json:"lastTime"`
	NeedModifyCount int64             `json:"needModifyCount"`
	NeedDeleteCount int64             `json:"needDeleteCount"`
	Messages        []WSQueuedMessage `json:"messages"`
}

type FolderSyncRenameMessage struct {
	Path        string `json:"path" form:"path" binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash"`
	Ctime       int64  `json:"ctime" form:"ctime"`
	Mtime       int64  `json:"mtime" form:"mtime"`
	OldPath     string `json:"oldPath" form:"oldPath"`
	OldPathHash string `json:"oldPathHash" form:"oldPathHash"`
}

// FolderSyncDeleteMessage message structure for folder deletion
// FolderSyncDeleteMessage 文件夹删除的消息结构
type FolderSyncDeleteMessage struct {
	Path     string `json:"path" form:"path"`         // Path info (file path) // 路径信息（文件路径）
	PathHash string `json:"pathHash" form:"pathHash"` // Path hash for fast lookup // 路径哈希值，用于快速查找
	Ctime    int64  `json:"ctime" form:"ctime"`       // Creation timestamp // 创建时间戳
	Mtime    int64  `json:"mtime" form:"mtime"`       // Modification timestamp // 修改时间戳
}

type FolderSyncModifyMessage struct {
	Path             string `json:"path" form:"path"`
	PathHash         string `json:"pathHash" form:"pathHash"`
	Ctime            int64  `json:"ctime" form:"ctime"`
	Mtime            int64  `json:"mtime" form:"mtime"`
	UpdatedTimestamp int64  `json:"lastTime" form:"updatedTimestamp"`
}
