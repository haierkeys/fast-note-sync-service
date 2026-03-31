package config

// TracerConfig request tracing configuration
// TracerConfig 请求追踪配置
type TracerConfig struct {
	// Enabled whether tracing is enabled
	// Enabled 是否启用追踪
	Enabled bool `yaml:"enabled" default:"true"`
	// Header tracing ID request header name, default X-Trace-ID
	// Header 追踪 ID 请求头名称，默认 X-Trace-ID
	Header string `yaml:"header" default:"X-Trace-ID"`
}
