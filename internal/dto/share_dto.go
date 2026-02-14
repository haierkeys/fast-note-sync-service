package dto

import "time"

// ShareCreateRequest Request parameters for creating a share
// 创建分享请求
type ShareCreateRequest struct {
	Vault    string `json:"vault" binding:"required" example:"defaultVault"` // Vault name // 保险库名称
	Path     string `json:"path" binding:"required"`                         // Resource path // 资源路径
	PathHash string `json:"path_hash" binding:"required"`                    // Resource path Hash // 资源路径哈希
}

// ShareResourceRequest Request parameters for retrieving a shared resource
// 分享资源获取请求
type ShareResourceRequest struct {
	ID int64 `json:"id" form:"id" binding:"required"` // Resource ID // 资源 ID
}

// ---------------- DTO / Response ----------------

// ShareCreateResponse Response for creating a share
// 创建分享响应
type ShareCreateResponse struct {
	ID        int64     `json:"id"`         // ID of the note or file table (primary resource ID) // 笔记或文件表 ID（主资源 ID）
	Type      string    `json:"type"`       // Resource type: note or file // 资源类型：笔记（note）或文件（file）
	Token     string    `json:"token"`      // Share Token // 分享 Token
	ExpiresAt time.Time `json:"expires_at"` // Expiration time // 过期时间
}
