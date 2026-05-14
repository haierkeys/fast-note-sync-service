package domain

import (
	"context"
	"time"
)

// AuthToken defines the authentication token domain model
// AuthToken 定义认证令牌领域模型
type AuthToken struct {
	ID          int64     // Primary Key // 主键
	UID         int64     // User ID // 用户 ID
	TokenString string    // Token String/Hash // 令牌字符串/哈希
	Scope       string    // Permission Scope // 权限范围
	ClientType  string    // Client Type (e.g. webgui, obsidian) // 客户端类型
	BoundIP     string    // Bound IP Address // 绑定 IP 地址
	UserAgent   string    // User Agent // 用户代理
	Status      int64     // Status (1: Active, 0: Revoked) // 状态 (1: 活跃, 0: 注销)
	ExpiredAt   time.Time // Expiration Time // 过期时间
	IssueType   int       // Issue Type (1: Login, 2: Manual) // 签发类型 (1: 登录, 2: 手动)
	LastUsedAt  time.Time // Last Used Time // 最后使用时间
	CreatedAt   time.Time // Creation Time // 创建时间
	UpdatedAt   time.Time // Update Time // 更新时间
}

// AuthTokenLog defines the token access log domain model
// AuthTokenLog 定义令牌访问日志领域模型
type AuthTokenLog struct {
	ID            int64     // Primary Key // 主键
	TokenID       int64     // Token ID // 令牌 ID
	UID           int64     // User ID // 用户 ID
	Protocol      string    // Protocol (rest, ws, mcp) // 协议
	Client        string    // Client (webgui, obsidian) // 客户端
	ClientName    string    // Client Name // 客户端名称
	ClientVersion string    // Client Version // 客户端版本
	IP            string    // Request IP // 请求 IP
	UA            string    // User Agent // 用户代理
	StatusCode    int64     // HTTP Status Code // HTTP 状态码
	CreatedAt     time.Time // Creation Time // 创建时间
}

// AuthTokenRepository defines the AuthToken repository interface
// AuthTokenRepository 定义认证令牌仓储接口
type AuthTokenRepository interface {
	// Create creates a new auth token
	// Create 创建新的认证令牌
	Create(ctx context.Context, token *AuthToken) (*AuthToken, error)

	// GetByID gets a token by ID
	// GetByID 根据 ID 获取令牌
	GetByID(ctx context.Context, id int64) (*AuthToken, error)

	// GetByTokenString gets a token by its string
	// GetByTokenString 根据令牌字符串获取令牌
	GetByTokenString(ctx context.Context, tokenString string) (*AuthToken, error)

	// ListByUID lists all active tokens for a user
	// ListByUID 列出用户的所有活跃令牌
	ListByUID(ctx context.Context, uid int64) ([]*AuthToken, error)

	// Update updates all properties of a token
	// Update 更新令牌的所有属性
	Update(ctx context.Context, token *AuthToken) error
	
	// UpdateScope updates the scope of a token
	// UpdateScope 更新令牌的权限范围
	UpdateScope(ctx context.Context, id int64, scope string) error

	// UpdateLastUsedAt updates the last used time of a token
	// UpdateLastUsedAt 更新令牌的最后使用时间
	UpdateLastUsedAt(ctx context.Context, id int64) error

	// Revoke revokes a token
	// Revoke 注销令牌
	Revoke(ctx context.Context, id int64) error

	// RevokeAllByUID revokes all active tokens for a user
	// RevokeAllByUID 注销用户的所有活跃令牌
	RevokeAllByUID(ctx context.Context, uid int64) error

	// UpdateTokenString updates the token string (nonce) of a token
	// UpdateTokenString 更新令牌字符串（标识符）
	UpdateTokenString(ctx context.Context, id int64, tokenString string) error
}

type AuthTokenLogRepository interface {
	// Create creates a new access log
	// Create 创建新的访问日志
	Create(ctx context.Context, log *AuthTokenLog) error
	// ListByTokenID lists access logs for a specific token with pagination
	// ListByTokenID 为特定令牌列出带有分页的访问日志
	ListByTokenID(ctx context.Context, tokenID int64, page, pageSize int) ([]*AuthTokenLog, int64, error)
}
