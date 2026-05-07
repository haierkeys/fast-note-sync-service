// Package service implements the business logic layer
// Package service 实现业务逻辑层
package service

// ServiceConfig service layer configuration
// ServiceConfig 服务层配置
type ServiceConfig struct {
	User UserServiceConfig // User related config // 用户相关配置
	App  AppServiceConfig  // App related config // 应用相关配置
}

// UserServiceConfig user service configuration
// UserServiceConfig 用户服务配置
type UserServiceConfig struct {
	RegisterIsEnable bool // Whether registration is enabled // 注册是否启用
	AdminUID         int  // Admin UID, 0 means no restriction // 管理员 UID，0 表示不限制
}

// AppServiceConfig app service configuration
// AppServiceConfig 应用服务配置
type AppServiceConfig struct {
	SoftDeleteRetentionTime string                 // Soft delete retention time (e.g., 7d, 24h, 30m, 0/empty for no cleanup) // 软删除保留时间（支持格式：7d、24h、30m、0 或空表示不自动清理）
	HistoryKeepVersions     int                    // History versions to keep // 历史记录保留版本数
	HistorySaveDelay        string                 // History save delay (e.g., 10s, 1m, default 10s) // 历史记录保存延迟时间（支持格式：10s、1m，默认 10s）
	ShareTokenExpiry        string                 // Share token expiry // 分享 Token 过期时间
	ShortLink               ShortLinkServiceConfig // Short link configuration // 短链配置
}

// ShortLinkServiceConfig short link service configuration
// ShortLinkServiceConfig 短链服务配置
type ShortLinkServiceConfig struct {
	BaseURL  string // Base URL // 基础 URL
	APIKey   string // API Key // API 密钥
	Password string // Password // 密码
	Cloaking bool   // Cloaking // 遮盖
}
