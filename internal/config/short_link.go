package config

// ShortLinkConfig short link configuration
// ShortLinkConfig 短链配置
type ShortLinkConfig struct {
	BaseURL  string `yaml:"base-url" default:"https://sink.cool"`
	APIKey   string `yaml:"api-key" default:"SinkCool"`
	Password string `yaml:"password" default:""`
	Cloaking bool   `yaml:"cloaking" default:"false"`
}
