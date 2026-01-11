// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// FileUpdateCheckRequest 客户端用于检查是否需要更新的请求参数
type FileUpdateCheckRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`
	Path        string `json:"path" form:"path" binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash" binding:""`
	Size        int64  `json:"size" form:"size" binding:""`
	Ctime       int64  `json:"ctime" form:"ctime" binding:"required"`
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required"`
}

// FileUpdateRequest 用于创建或修改文件的请求参数
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

// FileDeleteRequest 删除文件所需参数
type FileDeleteRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash"`
}

// FileSyncCheckRequest 同步检查单条记录的参数
type FileSyncCheckRequest struct {
	Path        string `json:"path" form:"path"`
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash" binding:""`
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required"`
	Size        int64  `json:"size" form:"size"`
}

// FileSyncRequest 同步请求主体
type FileSyncRequest struct {
	Vault    string                 `json:"vault" form:"vault" binding:"required"`
	LastTime int64                  `json:"lastTime" form:"lastTime"`
	Files    []FileSyncCheckRequest `json:"files" form:"files"`
}

// FileUploadCompleteRequest 文件上传完成参数
type FileUploadCompleteRequest struct {
	SessionID string `json:"sessionId" binding:"required"`
}

// FileGetRequest 用于获取单条文件的请求参数
type FileGetRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash"`
}

// FileListRequest 获取文件列表的分页参数
type FileListRequest struct {
	Vault string `json:"vault" form:"vault" binding:"required"`
}


// FileDTO 文件数据传输对象
type FileDTO struct {
	ID               int64      `json:"id" form:"id"`
	Action           string     `json:"-" form:"action"`
	Path             string     `json:"path" form:"path"`
	PathHash         string     `json:"pathHash" form:"pathHash"`
	ContentHash      string     `json:"contentHash" form:"contentHash"`
	SavePath         string     `json:"savePath" form:"savePath"`
	Size             int64      `json:"size" form:"size"`
	Ctime            int64      `json:"ctime" form:"ctime"`
	Mtime            int64      `json:"mtime" form:"mtime"`
	UpdatedTimestamp int64      `json:"lastTime" form:"updatedTimestamp"`
	UpdatedAt        timex.Time `json:"-"`
	CreatedAt        timex.Time `json:"-"`
}
