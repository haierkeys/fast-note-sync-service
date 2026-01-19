package dto

import "time"

// ShareCreateRequest 创建分享请求
type ShareCreateRequest struct {
	Vault    string `json:"vault" binding:"required" example:"defaultVault"` // 仓库名称
	Path     string `json:"path" binding:"required"`                         // 资源路径
	PathHash string `json:"path_hash" binding:"required"`                    // 资源路径 Hash
}

// ShareCreateResponse 创建分享响应
type ShareCreateResponse struct {
	ID        int64     `json:"id"`         // 笔记或附件表的 ID (主资源 ID)
	Type      string    `json:"type"`       // 资源类型: note 或 file
	Token     string    `json:"token"`      // 分享 Token
	ExpiresAt time.Time `json:"expires_at"` // 过期时间
}

// ShareResourceRequest 分享资源获取请求
type ShareResourceRequest struct {
	ID int64 `json:"id" form:"id" binding:"required"` // 资源 ID
}
