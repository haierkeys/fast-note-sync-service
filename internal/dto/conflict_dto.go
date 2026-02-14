// Package dto Defines data transfer objects (request parameters and response structs)
// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

// ConflictFileRequest Request parameters for creating a conflict file
// 创建冲突文件的请求参数
type ConflictFileRequest struct {
	Vault             string `json:"vault" form:"vault" binding:"required"`                         // Vault name // 保险库名称
	OriginalPath      string `json:"originalPath" form:"originalPath" binding:"required"`           // Original file path // 原始文件路径
	ClientContent     string `json:"clientContent" form:"clientContent" binding:"required"`         // Client side content // 客户端内容
	ClientContentHash string `json:"clientContentHash" form:"clientContentHash" binding:"required"` // Client side content hash // 客户端内容哈希
	Ctime             int64  `json:"ctime" form:"ctime"`                                            // Creation timestamp // 创建时间戳
	Mtime             int64  `json:"mtime" form:"mtime"`                                            // Modification timestamp // 修改时间戳
}

// ---------------- DTO / Response ----------------

// ConflictFileResponse Response for creating a conflict file
// 创建冲突文件的响应
type ConflictFileResponse struct {
	ConflictPath string `json:"conflictPath"` // Path of the created conflict file // 创建的冲突文件路径
	Message      string `json:"message"`      // Result message // 结果消息
	NoteID       int64  `json:"noteId"`       // Note ID // 笔记 ID
}
