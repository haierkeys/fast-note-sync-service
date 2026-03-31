package config

// NgrokConfig ngrok configuration
// NgrokConfig ngrok 配置
type NgrokConfig struct {
	// Enabled whether to enable ngrok tunnel
	Enabled bool `yaml:"enabled" default:"false"`
	// AuthToken ngrok auth token
	AuthToken string `yaml:"auth-token"`
	// Domain ngrok custom domain (optional)
	Domain string `yaml:"domain"`
}

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
