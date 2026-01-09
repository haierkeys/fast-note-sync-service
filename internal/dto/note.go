// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

// NoteUpdateCheckRequest 客户端用于检查是否需要更新的请求参数
type NoteUpdateCheckRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`
	Path        string `json:"path" form:"path" binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash" binding:""`
	Ctime       int64  `json:"ctime" form:"ctime" binding:"required"`
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required"`
}

// NoteModifyOrCreateRequest 用于创建或修改笔记的请求参数
type NoteModifyOrCreateRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`
	Path        string `json:"path" form:"path" binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash"`
	SrcPath     string `json:"srcPath" form:"srcPath"`
	SrcPathHash string `json:"srcPathHash" form:"srcPathHash"`
	BaseHash    string `json:"baseHash" form:"baseHash" binding:""`
	Content     string `json:"content" form:"content" binding:""`
	ContentHash string `json:"contentHash" form:"contentHash" binding:""`
	Ctime       int64  `json:"ctime" form:"ctime"`
	Mtime       int64  `json:"mtime" form:"mtime"`
	OldPath     string `json:"oldPath" form:"oldPath"`
	OldPathHash string `json:"oldPathHash" form:"oldPathHash"`
}

// ContentModifyRequest 专用于只修改内容的请求参数
type ContentModifyRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`
	Path        string `json:"path" form:"path" binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`
	Content     string `json:"content" form:"content" binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash" binding:"required"`
	Ctime       int64  `json:"ctime" form:"ctime" binding:"required"`
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required"`
}

// NoteDeleteRequest 删除笔记所需参数
type NoteDeleteRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash"`
}

// NoteSyncCheckRequest 同步检查单条记录的参数
type NoteSyncCheckRequest struct {
	Path        string `json:"path" form:"path"`
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash" binding:""`
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required"`
}

// NoteSyncRequest 同步请求主体
type NoteSyncRequest struct {
	Vault    string                 `json:"vault" form:"vault" binding:"required"`
	LastTime int64                  `json:"lastTime" form:"lastTime"`
	Notes    []NoteSyncCheckRequest `json:"notes" form:"notes"`
}

// ModifyMtimeFilesRequest 用于按 mtime 查询修改文件
type ModifyMtimeFilesRequest struct {
	Vault string `json:"vault" form:"vault" binding:"required"`
	Mtime int64  `json:"mtime" form:"mtime"`
}

// NoteGetRequest 用于获取单条笔记的请求参数
type NoteGetRequest struct {
	Vault     string `json:"vault" form:"vault" binding:"required"`
	Path      string `json:"path" form:"path" binding:"required"`
	PathHash  string `json:"pathHash" form:"pathHash"`
	IsRecycle bool   `json:"isRecycle" form:"isRecycle"`
}

// NoteRenameRequest 重命名笔记所需参数
type NoteRenameRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`
	Path        string `json:"path" form:"path" binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`
	OldPath     string `json:"oldPath" form:"oldPath" binding:"required"`
	OldPathHash string `json:"oldPathHash" form:"oldPathHash" binding:"required"`
}

// NoteListRequest 获取笔记列表的分页参数
type NoteListRequest struct {
	Vault     string `json:"vault" form:"vault" binding:"required"`
	Keyword   string `json:"keyword" form:"keyword"`
	IsRecycle bool   `json:"isRecycle" form:"isRecycle"`
}

// NoteWithFileLinksResponse 带有文件链接的笔记响应结构体
type NoteWithFileLinksResponse struct {
	ID               int64             `json:"id"`
	Path             string            `json:"path"`
	PathHash         string            `json:"pathHash"`
	Content          string            `json:"content"`
	ContentHash      string            `json:"contentHash"`
	FileLinks        map[string]string `json:"fileLinks"`
	Version          int64             `json:"version"`
	Ctime            int64             `json:"ctime"`
	Mtime            int64             `json:"mtime"`
	UpdatedTimestamp int64             `json:"updatedTimestamp"`
	UpdatedAt        interface{}       `json:"updatedAt"`
	CreatedAt        interface{}       `json:"createdAt"`
}

// NoteHistoryListRequest 笔记历史列表请求参数
type NoteHistoryListRequest struct {
	Vault     string `json:"vault" form:"vault" binding:"required"`
	Path      string `json:"path" form:"path" binding:"required"`
	PathHash  string `json:"pathHash" form:"pathHash"`
	IsRecycle bool   `json:"isRecycle" form:"isRecycle"`
}
