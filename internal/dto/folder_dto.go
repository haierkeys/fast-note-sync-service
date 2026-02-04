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
	UpdatedTimestamp int64      `json:"lastTime" form:"updatedTimestamp"`
	UpdatedAt        timex.Time `json:"-"`
	CreatedAt        timex.Time `json:"-"`
}

// FolderListRequest 获取文件夹列表的请求参数
type FolderListRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path"`
	PathHash string `json:"pathHash" form:"pathHash"`
}

// FolderCreateRequest 创建文件夹请求参数
type FolderCreateRequest struct {
	Vault string `json:"vault" form:"vault" binding:"required"`
	Path  string `json:"path" form:"path" binding:"required"`
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
	Vault   string `json:"vault" form:"vault" binding:"required"`
	Path    string `json:"path" form:"path" binding:"required"`
	OldPath string `json:"oldPath" form:"oldPath" binding:"required"`
}

// FolderContentRequest 获取文件夹内容的请求参数
type FolderContentRequest struct {
	Vault     string `json:"vault" form:"vault" binding:"required"`
	Path      string `json:"path" form:"path"`
	PathHash  string `json:"pathHash" form:"pathHash"`
	Page      int    `json:"page" form:"page"`
	PageSize  int    `json:"pageSize" form:"pageSize"`
	SortBy    string `json:"sortBy" form:"sortBy"`
	SortOrder string `json:"sortOrder" form:"sortOrder"`
}
