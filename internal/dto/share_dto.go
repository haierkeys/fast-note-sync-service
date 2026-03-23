package dto

import (
	"time"

	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
)

// ShareCreateRequest Request parameters for creating a share
// 创建分享请求
type ShareCreateRequest struct {
	Vault    string `json:"vault" binding:"required" example:"defaultVault"` // Vault name // 保险库名称
	Path     string `json:"path" binding:"required" example:"ReadMe.md"`     // Resource path // 资源路径
	PathHash string `json:"pathHash" binding:"required" example:"hash123"`   // Resource path Hash // 资源路径哈希
	Password string `json:"password" example:"123456"`                      // Share password // 分享密码
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
	Vault    string `json:"vault" binding:"required" example:"defaultVault"` // Vault name // 保险库名称
	ID       int64  `json:"id" example:"1"`                                  // Share ID (optional) // 分享 ID (可选)
	Path     string `json:"path" example:"ReadMe.md"`                        // Resource path (optional) // 资源路径 (可选)
	PathHash string `json:"pathHash" example:"hash123"`                      // Resource path Hash (optional) // 资源路径哈希 (可选)
}

// ShareResourceRequest Request parameters for retrieving a shared resource
// 分享资源获取请求
type ShareResourceRequest struct {
	ID       int64  `json:"id" form:"id" binding:"required" example:"1"` // Resource ID // 资源 ID
	Password string `json:"password" form:"password" example:"123456"`    // Share password // 分享密码
}

// SharePasswordUpdateRequest Request parameters for updating share password
// 更新分享密码请求
type SharePasswordUpdateRequest struct {
	Vault    string `json:"vault" binding:"required" example:"test"`         // Vault name // 保险库名称
	Path     string `json:"path" binding:"required" example:"未命名.md"`        // Resource path // 资源路径
	PathHash string `json:"pathHash" binding:"required" example:"-677306325"` // Resource path Hash // 资源路径哈希
	Password string `json:"password" example:"123456"`                      // New password // 新密码
}

// ShareShortLinkCreateRequest Request parameters for creating a short link
// 创建短链请求
type ShareShortLinkCreateRequest struct {
	Vault    string `json:"vault" binding:"required" example:"work"` // Vault name // 库名
	Path     string `json:"path" binding:"required" example:"notes/todo.md"` // Path // 路径
	PathHash string `json:"pathHash" binding:"required" example:"..."` // Path hash // 路径哈希
	IsForce  bool   `json:"is_force" example:"false"` // Whether to force regeneration // 是否强制重新生成
}

// ShareListRequest Request parameters for listing shares
// 分享列表请求
type ShareListRequest struct {
	pkgapp.PaginationRequest
	SortBy    string `json:"sort_by" form:"sort_by" example:"created_at"`     // Sort field: created_at, updated_at, expires_at // 排序字段: created_at, updated_at, expires_at
	SortOrder string `json:"sort_order" form:"sort_order" example:"desc"` // Sort direction: asc or desc // 排序方向: asc 或 desc
}

// ---------------- DTO / Response ----------------

// ShareCreateResponse Response for creating a share
// 创建分享响应
type ShareCreateResponse struct {
	ID        int64     `json:"id"`         // ID of the note or file table (primary resource ID) // 笔记或文件表 ID（主资源 ID）
	Type      string    `json:"type"`       // Resource type: note or file // 资源类型：笔记（note）或文件（file）
	Token     string    `json:"token"`      // Share Token // 分享 Token
	ExpiresAt time.Time `json:"expires_at"` // Expiration time // 过期时间
	ShortLink string    `json:"short_link"` // Short link // 短链
}

// ShareListItem Represents a share item in list
// 分享列表项
type ShareListItem struct {
	ID           int64               `json:"id"`             // Share ID // 分享记录 ID
	UID          int64               `json:"uid"`            // User ID // 用户 ID
	Title        string              `json:"title"`          // Resource title (note title or file name) // 资源标题（笔记标题或文件名）
	URL          string              `json:"url"`            // Share URL (path format: /id/token) // 分享 URL (路径格式: /id/token)
	Resources    map[string][]string `json:"res"`            // Authorized resources // 资源授权列表
	NotePath     string              `json:"notePath"`       // Note path, for frontend share filter matching // 笔记路径，用于前端分享筛选匹配
	Status       int64               `json:"status"`         // Status: 1-Active, 2-Cancelled // 状态: 1-有效, 2-已撤销
	ViewCount    int64               `json:"view_count"`     // View count // 访问次数
	LastViewedAt time.Time           `json:"last_viewed_at"` // Last viewed time // 最后访问时间
	ExpiresAt    time.Time           `json:"expires_at"`     // Expiration time // 过期时间
	ShortLink    string              `json:"short_link"`     // Short link // 短链
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

// ShareListResponse Response for listing shares
// 分享列表响应
type ShareListResponse struct {
	Items []*ShareListItem `json:"items"` // Share list // 分享列表
}
