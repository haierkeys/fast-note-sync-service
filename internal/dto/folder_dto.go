package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// FolderDTO 文件夹数据传输对象
type FolderDTO struct {
	ID               int64      `json:"-" form:"id"`
	Action           string     `json:"-" form:"action"`
	Path             string     `json:"path" form:"path"`
	PathHash         string     `json:"pathHash" form:"pathHash"`
	Level            int64      `json:"-" form:"level"`
	FID              int64      `json:"-" form:"fid"`
	Ctime            int64      `json:"ctime" form:"ctime"`
	Mtime            int64      `json:"mtime" form:"mtime"`
	UpdatedTimestamp int64      `json:"lastTime" form:"updatedTimestamp"`
	UpdatedAt        timex.Time `json:"updatedAt"`
	CreatedAt        timex.Time `json:"createdAt"`
}

// FolderGetRequest 获取文件夹的请求参数
type FolderGetRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path"`
	PathHash string `json:"pathHash" form:"pathHash"`
}

// FolderListRequest 获取文件夹列表的请求参数
type FolderListRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path"`
	PathHash string `json:"pathHash" form:"pathHash"`
}

// FolderCreateRequest 创建文件夹请求参数
type FolderCreateRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash"`
}

// FolderDeleteRequest 删除文件夹请求参数
type FolderDeleteRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash"`
}

// FolderSyncCheckRequest 同步检查单条记录的参数
type FolderSyncCheckRequest struct {
	Path     string `json:"path" form:"path"`
	PathHash string `json:"pathHash" form:"pathHash" binding:"required"`
	Mtime    int64  `json:"mtime" form:"mtime" binding:"required"`
}

// FolderSyncDelFolder 同步删除/缺失文件夹的参数
type FolderSyncDelFolder struct {
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash" binding:"required"`
}

// FolderSyncRequest 同步请求主体
type FolderSyncRequest struct {
	Vault          string                   `json:"vault" form:"vault" binding:"required"`
	LastTime       int64                    `json:"lastTime" form:"lastTime"`
	Folders        []FolderSyncCheckRequest `json:"folders" form:"folders"`
	DelFolders     []FolderSyncDelFolder    `json:"delFolders" form:"delFolders"`
	MissingFolders []FolderSyncDelFolder    `json:"missingFolders" form:"missingFolders"`
}

// FolderRenameRequest 文件夹重命名请求参数
type FolderRenameRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`
	Path        string `json:"path" form:"path" binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`
	OldPath     string `json:"oldPath" form:"oldPath" binding:"required"`
	OldPathHash string `json:"oldPathHash" form:"oldPathHash" binding:"required"`
}

// FolderContentRequest 获取文件夹内容的请求参数
type FolderContentRequest struct {
	Vault     string `json:"vault" form:"vault" binding:"required"`
	Path      string `json:"path" form:"path"`
	PathHash  string `json:"pathHash" form:"pathHash"`
	SortBy    string `json:"sortBy" form:"sortBy"`
	SortOrder string `json:"sortOrder" form:"sortOrder"`
}

// FolderTreeRequest 获取文件夹树的请求参数
type FolderTreeRequest struct {
	Vault string `json:"vault" form:"vault" binding:"required"`
	Depth int    `json:"depth" form:"depth"` // 0 or negative = unlimited
}

// FolderTreeNode 文件夹树节点
type FolderTreeNode struct {
	Path      string            `json:"path"`
	Name      string            `json:"name"`
	NoteCount int               `json:"noteCount"`
	FileCount int               `json:"fileCount"`
	Children  []*FolderTreeNode `json:"children,omitempty"`
}

// FolderTreeResponse 文件夹树响应
type FolderTreeResponse struct {
	Folders       []*FolderTreeNode `json:"folders"`
	RootNoteCount int               `json:"rootNoteCount"`
	RootFileCount int               `json:"rootFileCount"`
}
