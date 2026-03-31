package config

// UserConfig user configuration
// UserConfig 用户配置
type UserConfig struct {
	// RegisterIsEnable whether registration is enabled
	// RegisterIsEnable 注册是否启用
	RegisterIsEnable bool `yaml:"register-is-enable"`
	// AdminUID admin UID, 0 means no restriction on admin access
	// AdminUID 管理员 UID，0 表示不限制管理员访问
	AdminUID int `yaml:"admin-uid" default:"0"`
}
