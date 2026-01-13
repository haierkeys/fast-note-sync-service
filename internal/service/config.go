// Package service 实现业务逻辑层
package service

// ServiceConfig 服务层配置
type ServiceConfig struct {
	// User 用户相关配置
	User UserServiceConfig

	// App 应用相关配置
	App AppServiceConfig
}

// UserServiceConfig 用户服务配置
type UserServiceConfig struct {
	// RegisterIsEnable 注册是否启用
	RegisterIsEnable bool
}

// AppServiceConfig 应用服务配置
type AppServiceConfig struct {
	// SoftDeleteRetentionTime 软删除保留时间
	// 支持格式：7d（天）、24h（小时）、30m（分钟）、0 或空表示不自动清理
	SoftDeleteRetentionTime string
	// HistoryKeepVersions 历史记录保留版本数
	HistoryKeepVersions int
	// HistorySaveDelay 历史记录保存延迟时间
	// 支持格式：10s（秒）、1m（分钟），默认 10s
	HistorySaveDelay string
}
