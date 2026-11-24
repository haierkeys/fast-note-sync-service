package config

// UserConfig user configuration
// UserConfig 用户配置
type UserConfig struct {
	// PasswordLoginIsEnable defaults to true; a pointer preserves explicit false across defaults.Set.
	// PasswordLoginIsEnable 默认开启；指针保留显式 false，避免被 defaults.Set 覆盖。
	PasswordLoginIsEnable *bool `yaml:"password-login-is-enable" default:"true"`
	// RegisterIsEnable whether registration is enabled
	// RegisterIsEnable 注册是否启用
	RegisterIsEnable bool `yaml:"register-is-enable"`
	// AdminUID admin UID, 0 means no restriction on admin access
	// AdminUID 管理员 UID，0 表示不限制管理员访问
	AdminUID int `yaml:"admin-uid" default:"0"`
}

// PasswordLoginEnabled treats an absent option as enabled for backward compatibility.
// PasswordLoginEnabled 未配置时默认开启，以保持向后兼容。
func (c UserConfig) PasswordLoginEnabled() bool {
	return c.PasswordLoginIsEnable == nil || *c.PasswordLoginIsEnable
}
