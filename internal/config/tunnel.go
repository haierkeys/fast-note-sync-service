package config



// CloudflareConfig cloudflare configuration
// CloudflareConfig cloudflare 配置
type CloudflareConfig struct {
	// Enabled whether to enable cloudflare tunnel
	Enabled bool `yaml:"enabled" default:"false"`
	// Token cloudflare tunnel token
	Token string `yaml:"token"`
	// LogEnabled whether to enable cloudflare tunnel logging
	LogEnabled bool `yaml:"log-enabled" default:"false"`
}
