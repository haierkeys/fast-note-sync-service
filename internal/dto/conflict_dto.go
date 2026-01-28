// Package dto Defines data transfer objects (request parameters and response structs)
// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

// ConflictFileRequest Request parameters for creating a conflict file
// 创建冲突文件的请求参数
type ConflictFileRequest struct {
	Vault             string `json:"vault" form:"vault" binding:"required"`
	OriginalPath      string `json:"originalPath" form:"originalPath" binding:"required"`
	ClientContent     string `json:"clientContent" form:"clientContent" binding:"required"`
	ClientContentHash string `json:"clientContentHash" form:"clientContentHash" binding:"required"`
	Ctime             int64  `json:"ctime" form:"ctime"`
	Mtime             int64  `json:"mtime" form:"mtime"`
}

// ConflictFileResponse Response for creating a conflict file
// 创建冲突文件的响应
type ConflictFileResponse struct {
	ConflictPath string `json:"conflictPath"`
	Message      string `json:"message"`
	NoteID       int64  `json:"noteId"`
}
