package config

// ServerConfig server configuration
// ServerConfig 服务器配置
type ServerConfig struct {
	// RunMode run mode
	// RunMode 运行模式
	RunMode string `yaml:"run-mode" default:"release"`
	// HttpPort HTTP port
	// HttpPort HTTP 端口
	HttpPort string `yaml:"http-port" default:":9000"`
	// ReadTimeout read timeout (seconds)
	// ReadTimeout 读取超时（秒）
	ReadTimeout int `yaml:"read-timeout" default:"60"`
	// WriteTimeout write timeout (seconds)
	// WriteTimeout 写入超时（秒）
	WriteTimeout int `yaml:"write-timeout" default:"60"`
	// PrivateHttpListen private HTTP listen address
	// PrivateHttpListen 私有 HTTP 监听地址
	PrivateHttpListen string `yaml:"private-http-listen"`
	// MCPSSEPingInterval MCP SSE ping interval (seconds)
	// MCPSSEPingInterval MCP SSE 保活心跳间隔（秒）
	MCPSSEPingInterval int `yaml:"mcp-sse-ping-interval" default:"30"`
}
