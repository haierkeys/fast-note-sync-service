package config

// LogConfig log configuration
// LogConfig 日志配置
type LogConfig struct {
	// Level log level, see zapcore.ParseLevel
	// Level 日志级别，参见 zapcore.ParseLevel
	Level string `yaml:"level" default:"warn"`
	// File log file path, default stderr
	// File 日志文件路径，默认为 stderr
	File string `yaml:"file" default:"storage/logs/log.log"`
	// Production whether to enable JSON output
	// Production 是否启用 JSON 输出
	Production bool `yaml:"production"`
}
