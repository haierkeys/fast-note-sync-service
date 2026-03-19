package dto

import "time"

// ShareCreateRequest Request parameters for creating a share
// 创建分享请求
type ShareCreateRequest struct {
	Vault    string `json:"vault" binding:"required" example:"defaultVault"` // Vault name // 保险库名称
	Path     string `json:"path" binding:"required" example:"ReadMe.md"`     // Resource path // 资源路径
	PathHash string `json:"pathHash" binding:"required" example:"hash123"`   // Resource path Hash // 资源路径哈希
}

// ShareQueryRequest Request parameters for querying a share
// 查询分享请求
type ShareQueryRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required" example:"defaultVault"` // Vault name // 保险库名称
	Path     string `json:"path" form:"path" binding:"required" example:"ReadMe.md"`     // Resource path // 资源路径
	PathHash string `json:"pathHash" form:"pathHash" binding:"required" example:"hash123"`   // Resource path Hash // 资源路径哈希
}

// ShareCancelRequest Request parameters for cancelling a share
// 取消分享请求
type ShareCancelRequest struct {
	Vault    string `json:"vault" example:"defaultVault"` // Vault name (required only when cancelling by path) // 保险库名称（按路径取消时必填）
	ID       int64  `json:"id" example:"1"`               // Share ID (cancel by ID when > 0) // 分享 ID（大于 0 时按 ID 取消）
	Path     string `json:"path" example:"ReadMe.md"`     // Resource path (optional) // 资源路径 (可选)
	PathHash string `json:"pathHash" example:"hash123"`   // Resource path Hash (optional) // 资源路径哈希 (可选)
}

// ShareResourceRequest Request parameters for retrieving a shared resource
// 分享资源获取请求
type ShareResourceRequest struct {
	ID int64 `json:"id" form:"id" binding:"required" example:"1"` // Resource ID // 资源 ID
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

// ShareNoteInfo Note information attached to a share list item
// ShareNoteInfo 分享列表项关联的笔记信息
type ShareNoteInfo struct {
	ID      int64  `json:"id"`      // Note ID // 笔记 ID
	Path    string `json:"path"`    // Note path // 笔记路径
	Ctime   int64  `json:"ctime"`   // Creation timestamp (ms) // 创建时间戳（毫秒）
	Mtime   int64  `json:"mtime"`   // Modification timestamp (ms) // 修改时间戳（毫秒）
	Version int64  `json:"version"` // Version number // 版本号
}

// ShareListItem Represents a share item in list
// 分享列表项
type ShareListItem struct {
	ID           int64               `json:"id"`             // Share ID // 分享记录 ID
	UID          int64               `json:"uid"`            // User ID // 用户 ID
	Resources    map[string][]string `json:"res"`            // Authorized resources // 资源授权列表
	Status       int64               `json:"status"`         // Status: 1-Active, 2-Cancelled // 状态: 1-有效, 2-已撤销
	ViewCount    int64               `json:"view_count"`     // View count // 访问次数
	LastViewedAt time.Time           `json:"last_viewed_at"` // Last viewed time // 最后访问时间
	ExpiresAt    time.Time           `json:"expires_at"`     // Expiration time // 过期时间
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	NoteInfo     *ShareNoteInfo      `json:"note_info,omitempty"` // Note details (only for note shares) // 笔记详情（仅笔记分享时有值）
}

// ShareListResponse Response for listing shares
// 分享列表响应
type ShareListResponse struct {
	Items []*ShareListItem `json:"items"` // Share list // 分享列表
}
