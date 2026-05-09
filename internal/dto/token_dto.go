package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// TokenIssueRequest defines the request to manually issue a new token
// TokenIssueRequest 定义手动签发新令牌的请求
type TokenIssueRequest struct {
	ClientType  string `json:"clientType" binding:"required"` // Client Type (e.g., obsidian, mobile) // 客户端类型
	Scope       string `json:"scope" binding:"required"`      // Permission Scope // 权限范围
	ExpiredDays int    `json:"expiredDays" binding:"min=1"`   // Expired days // 过期天数
}

// TokenUpdateRequest defines the request to update a token's scope
// TokenUpdateRequest 定义更新令牌权限范围的请求
type TokenUpdateRequest struct {
	Scope string `json:"scope" binding:"required"` // Permission Scope // 权限范围
}

// TokenResponse defines the response structure for a token
// TokenResponse 定义令牌的响应结构
type TokenResponse struct {
	ID         int64      `json:"id"`
	Scope      string     `json:"scope"`
	ClientType string     `json:"clientType"`
	BoundIP    string     `json:"boundIp"`
	UserAgent  string     `json:"userAgent"`
	ExpiredAt  timex.Time `json:"expiredAt"`
	CreatedAt  timex.Time `json:"createdAt"`
}

// TokenCreateResponse defines the response structure when creating a token
// TokenCreateResponse 定义创建令牌时的响应结构
type TokenCreateResponse struct {
	TokenResponse
	TokenString string `json:"token"` // The actual JWT token // 实际的 JWT 令牌
}
