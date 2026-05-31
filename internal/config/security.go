package config

// SecurityConfig security configuration
// SecurityConfig 安全配置
type SecurityConfig struct {
	AuthTokenKey string `yaml:"auth-token-key" default:"fast-note-sync-Auth-Token"`
	TokenExpiry  string `yaml:"token-expiry" default:"365d"` // Token expiry, supports format: 7d (days), 24h (hours), 30m (minutes)
	// Token 过期时间，支持格式：7d（天）、24h（小时）、30m（分钟）
	ShareTokenKey string `yaml:"share-token-key" default:"fns"`
	// ShareTokenExpiry share Token expiry
	// ShareTokenExpiry 分享 Token 过期时间
	ShareTokenExpiry string `yaml:"share-token-expiry" default:"30d"`
	// WebGUILoginTokenExpiry expiry duration for WebGUI auto-issued login tokens (e.g. 7d, 24h)
	// WebGUILoginTokenExpiry WebGUI 自动签发登录 Token 的有效期（如 7d、24h）
	WebGUILoginTokenExpiry string `yaml:"webgui-login-token-expiry" default:"7d"`
	// WebGUILoginTokenBindIP whether to bind the client IP when issuing WebGUI login tokens
	// WebGUILoginTokenBindIP 签发 WebGUI 登录 Token 时是否绑定客户端 IP
	WebGUILoginTokenBindIP *bool `yaml:"webgui-login-token-bind-ip" default:"true"`
}
