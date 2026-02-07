// Package dto Defines data transfer objects (request parameters and response structs)
// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// FileUpdateCheckRequest Client request parameters for checking if updates are needed
// 客户端用于检查是否需要更新的请求参数
type FileUpdateCheckRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`
	Path        string `json:"path" form:"path" binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash" binding:""`
	Size        int64  `json:"size" form:"size" binding:""`
	Ctime       int64  `json:"ctime" form:"ctime" binding:"required"`
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required"`
}

// FileUpdateRequest Request parameters for creating or modifying a file
// 用于创建或修改文件的请求参数
type FileUpdateRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`
	Path        string `json:"path" form:"path" binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash"`
	ContentHash string `json:"contentHash" form:"contentHash" binding:""`
	SavePath    string `json:"savePath" form:"savePath" binding:""`
	Size        int64  `json:"size" form:"size"`
	Ctime       int64  `json:"ctime" form:"ctime"`
	Mtime       int64  `json:"mtime" form:"mtime"`
}

// FileDeleteRequest Parameters required for deleting a file
// 删除文件所需参数
type FileDeleteRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash" binding:"required"`
}

// FileRestoreRequest parameters for restoring a file
// FileRestoreRequest 恢复文件请求参数
type FileRestoreRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash"`
}

// FileSyncCheckRequest/ Parameters for checking synchronization of a single record
// 同步检查单条记录的参数
type FileSyncCheckRequest struct {
	Path        string `json:"path" form:"path"`
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash" binding:""`
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required"`
	Size        int64  `json:"size" form:"size"`
}

type FileSyncDelFile struct {
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash" binding:"required"`
}

// FileSyncRequest Synchronization request body
// 同步请求主体
type FileSyncRequest struct {
	Vault        string                 `json:"vault" form:"vault" binding:"required"`
	LastTime     int64                  `json:"lastTime" form:"lastTime"`
	Files        []FileSyncCheckRequest `json:"files" form:"files"`
	DelFiles     []FileSyncDelFile      `json:"delFiles" form:"delFiles"`
	MissingFiles []FileSyncDelFile      `json:"missingFiles" form:"missingFiles"`
}

// FileUploadCompleteRequest Parameters for file upload completion
// 文件上传完成参数
type FileUploadCompleteRequest struct {
	SessionID string `json:"sessionId" binding:"required"`
}

// FileGetRequest Request parameters for retrieving a single file
// 用于获取单条文件的请求参数
type FileGetRequest struct {
	Vault     string `json:"vault" form:"vault" binding:"required"`
	Path      string `json:"path" form:"path" binding:"required"`
	PathHash  string `json:"pathHash" form:"pathHash"`
	IsRecycle bool   `json:"isRecycle" form:"isRecycle"`
}

// FileListRequest Pagination parameters for retrieving the file list
// 获取文件列表的分页参数
type FileListRequest struct {
	Vault     string `json:"vault" form:"vault" binding:"required"`
	Keyword   string `json:"keyword" form:"keyword"`
	IsRecycle bool   `json:"isRecycle" form:"isRecycle"`
	SortBy    string `json:"sortBy" form:"sortBy"`       // Sorting field: mtime(default), ctime, path
	SortOrder string `json:"sortOrder" form:"sortOrder"` // Sorting order: desc(default), asc
}

// FileRenameRequest Parameters required for renaming a file
// 重命名文件所需参数
type FileRenameRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`
	Path        string `json:"path" form:"path" binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`
	OldPath     string `json:"oldPath" form:"oldPath" binding:"required"`
	OldPathHash string `json:"oldPathHash" form:"oldPathHash" binding:"required"`
}

// FileDTO File Data Transfer Object
// FileDTO 文件数据传输对象
type FileDTO struct {
	ID               int64      `json:"-"`
	Action           string     `json:"-"`
	Path             string     `json:"path" form:"path"`
	PathHash         string     `json:"pathHash" form:"pathHash"`
	ContentHash      string     `json:"contentHash" form:"contentHash"`
	SavePath         string     `json:"-"`
	Size             int64      `json:"size" form:"size"`
	Ctime            int64      `json:"ctime" form:"ctime"`
	Mtime            int64      `json:"mtime" form:"mtime"`
	UpdatedTimestamp int64      `json:"lastTime" form:"updatedTimestamp"`
	UpdatedAt        timex.Time `json:"-"`
	CreatedAt        timex.Time `json:"-"`
}
