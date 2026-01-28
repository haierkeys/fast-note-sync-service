package dto

import "time"

// ShareCreateRequest Request parameters for creating a share
// 创建分享请求
type ShareCreateRequest struct {
	Vault    string `json:"vault" binding:"required" example:"defaultVault"` // Vault name
	Path     string `json:"path" binding:"required"`                         // Resource path
	PathHash string `json:"path_hash" binding:"required"`                    // Resource path Hash
}

// ShareCreateResponse Response for creating a share
// 创建分享响应
type ShareCreateResponse struct {
	ID        int64     `json:"id"`         // ID of the note or file table (primary resource ID)
	Type      string    `json:"type"`       // Resource type: note or file
	Token     string    `json:"token"`      // Share Token
	ExpiresAt time.Time `json:"expires_at"` // Expiration time
}

// ShareResourceRequest Request parameters for retrieving a shared resource
// 分享资源获取请求
type ShareResourceRequest struct {
	ID int64 `json:"id" form:"id" binding:"required"` // Resource ID
}
