package dto

import (
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
)

// TokenIssueRequest defines the request to manually issue a new token
// TokenIssueRequest 定义手动签发新令牌的请求
type TokenIssueRequest struct {
	ClientType  string `json:"clientType" binding:"required"` // Client Type (e.g., obsidian, mobile) // 客户端类型
	Scope       string `json:"scope"`                         // Permission Scope (legacy or protocols) // 权限范围
	Protocol    string `json:"protocol"`                      // Optional: explicit protocol dimension (p:) // 可选：明确的协议维度
	Client      string `json:"client"`                        // Optional: explicit client dimension (c:) // 可选：明确的客户端维度
	Function    string `json:"function"`                      // Optional: explicit function dimension (f:) // 可选：明确的功能维度
	ExpiredDays int    `json:"expiredDays" binding:"min=1"`   // Expired days // 过期天数
	BoundIP     string `json:"boundIp"`                       // Optional: Bound IP // 可选：绑定 IP
	UserAgent   string `json:"userAgent"`                     // Optional: User Agent // 可选：User Agent
	Vaults      string `json:"vaults"`                        // Optional: Restrict Vaults (comma-separated) // 可选：限制笔记库（逗号分隔）
}

// TokenUpdateRequest defines the request to update a token's properties
// TokenUpdateRequest 定义更新令牌属性的请求
type TokenUpdateRequest struct {
	ClientType  string `json:"clientType"`  // Client Type // 客户端类型
	Scope       string `json:"scope"`       // Permission Scope // 权限范围
	Protocol    string `json:"protocol"`    // Optional: explicit protocol dimension (p:) // 可选：明确的协议维度
	Client      string `json:"client"`      // Optional: explicit client dimension (c:) // 可选：明确的客户端维度
	Function    string `json:"function"`    // Optional: explicit function dimension (f:) // 可选：明确的功能维度
	ExpiredDays int    `json:"expiredDays"` // Expired days // 过期天数
	BoundIP     string `json:"boundIp"`     // Optional: Bound IP // 可选：绑定 IP
	UserAgent   string `json:"userAgent"`   // Optional: User Agent // 可选：User Agent
	Vaults      string `json:"vaults"`      // Optional: Restrict Vaults (comma-separated) // 可选：限制笔记库（逗号分隔）
}

// TokenActiveClient defines the information for an active token client
// TokenActiveClient 定义活跃令牌客户端的信息
type TokenActiveClient struct {
	Name     string          `json:"name"`
	Platform map[string]bool `json:"platform"`
}

// TokenResponse defines the response structure for a token
// TokenResponse 定义令牌的响应结构
type TokenResponse struct {
	ID         int64      `json:"id"`                   // Token ID // 令牌 ID
	Scope      string     `json:"scope"`                // Permission Scope // 权限范围
	ClientType string     `json:"clientType"`           // Client Type // 客户端类型
	BoundIP    string     `json:"boundIp"`              // Bound IP // 绑定 IP
	UserAgent  string     `json:"userAgent"`            // User Agent // 用户代理
	Vaults     string     `json:"vaults"`               // Restrict Vaults // 限制笔记库
	IssueType  int        `json:"issueType"`            // Issue Type // 签发类型
	LastUsedAt timex.Time `json:"lastUsedAt"`           // Last Used Time // 最后使用时间
	ExpiredAt  timex.Time `json:"expiredAt"`            // Expiration Time // 过期时间
	CreatedAt  timex.Time `json:"createdAt"`            // Creation Time // 创建时间
	IsWsOnline    bool                `json:"isWsOnline"`    // Is WS Online // WS 是否在线
	ActiveClients []string            `json:"activeClients"` // Active Clients // 活跃客户端
}

// TokenCreateResponse defines the response structure when creating a token
// TokenCreateResponse 定义创建令牌时的响应结构
type TokenCreateResponse struct {
	TokenResponse
	TokenString string `json:"token"` // The actual JWT token // 实际的 JWT 令牌
}

// TokenLogResponse defines the response structure for a token access log
// TokenLogResponse 定义令牌访问日志的响应结构
type TokenLogResponse struct {
	ID            int64      `json:"id"`
	Protocol      string     `json:"protocol"`
	Client        string     `json:"client"`
	ClientName    string     `json:"clientName"`
	ClientVersion string     `json:"clientVersion"`
	IP            string     `json:"ip"`
	UA            string     `json:"ua"`
	StatusCode    int64      `json:"statusCode"`
	CreatedAt     timex.Time `json:"createdAt"`
}

// TokenLogListRequest defines the request to list token logs
// TokenLogListRequest 定义列出令牌日志的请求
type TokenLogListRequest struct {
	pkgapp.PaginationRequest
}
