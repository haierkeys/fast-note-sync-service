// Package app 提供应用容器，封装所有依赖和服务
package app

import (
	"os"
	"path/filepath"
	"time"

	"github.com/haierkeys/fast-note-sync-service/pkg/storage/local_fs"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/haierkeys/fast-note-sync-service/pkg/workerpool"
	"github.com/haierkeys/fast-note-sync-service/pkg/writequeue"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// AppConfig 应用配置
type AppConfig struct {
	File     string          `yaml:"-"` // 配置文件路径，不序列化
	Server   ServerConfig    `yaml:"server"`
	Log      LogConfig       `yaml:"log"`
	Database DatabaseConfig  `yaml:"database"`
	App      AppSettings     `yaml:"app"`
	User     UserConfig      `yaml:"user"`
	Security SecurityConfig  `yaml:"security"`
	LocalFS  local_fs.Config `yaml:"local-fs"`
	WebGUI   WebGUIConfig    `yaml:"webgui"`
	Tracer   TracerConfig    `yaml:"tracer"`
}

// LogConfig 日志配置
type LogConfig struct {
	// Level 日志级别，参见 zapcore.ParseLevel
	Level string `yaml:"level"`
	// File 日志文件路径，默认为 stderr
	File string `yaml:"file"`
	// Production 是否启用 JSON 输出
	Production bool `yaml:"production"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	// RunMode 运行模式
	RunMode string `yaml:"run-mode"`
	// HttpPort HTTP 端口
	HttpPort string `yaml:"http-port"`
	// ReadTimeout 读取超时（秒）
	ReadTimeout int `yaml:"read-timeout"`
	// WriteTimeout 写入超时（秒）
	WriteTimeout int `yaml:"write-timeout"`
	// PrivateHttpListen 私有 HTTP 监听地址
	PrivateHttpListen string `yaml:"private-http-listen"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	AuthToken    string `yaml:"auth-token"`
	AuthTokenKey string `yaml:"auth-token-key"`
	TokenExpiry  string `yaml:"token-expiry"` // Token 过期时间，支持格式：7d（天）、24h（小时）、30m（分钟）
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	// Type 数据库类型
	Type string `yaml:"type"`
	// Path SQLite 数据库文件路径
	Path string `yaml:"path"`
	// UserName 用户名
	UserName string `yaml:"username"`
	// Password 密码
	Password string `yaml:"password"`
	// Host 主机
	Host string `yaml:"host"`
	// Name 数据库名
	Name string `yaml:"name"`
	// TablePrefix 表前缀
	TablePrefix string `yaml:"table-prefix"`
	// AutoMigrate 是否启用自动迁移
	AutoMigrate bool `yaml:"auto-migrate"`
	// Charset 字符集
	Charset string `yaml:"charset"`
	// ParseTime 是否解析时间
	ParseTime bool `yaml:"parse-time"`
	// MaxIdleConns 最大闲置连接数，默认 10
	MaxIdleConns int `yaml:"max-idle-conns"`
	// MaxOpenConns 最大打开连接数，默认 100
	MaxOpenConns int `yaml:"max-open-conns"`
	// ConnMaxLifetime 连接最大生命周期，支持格式：30m（分钟）、1h（小时），默认 30m
	ConnMaxLifetime string `yaml:"conn-max-lifetime"`
	// ConnMaxIdleTime 空闲连接最大生命周期，支持格式：10m（分钟）、1h（小时），默认 10m
	ConnMaxIdleTime string `yaml:"conn-max-idle-time"`
}

// UserConfig 用户配置
type UserConfig struct {
	// RegisterIsEnable 注册是否启用
	RegisterIsEnable bool `yaml:"register-is-enable"`
	// AdminUID 管理员 UID，0 表示不限制管理员访问
	AdminUID int `yaml:"admin-uid"`
}

// AppSettings 应用设置
type AppSettings struct {
	// DefaultPageSize 默认页面大小
	DefaultPageSize int `yaml:"default-page-size"`
	// MaxPageSize 最大页面大小
	MaxPageSize int `yaml:"max-page-size"`
	// DefaultContextTimeout 默认上下文超时时间
	DefaultContextTimeout int `yaml:"default-context-timeout"`
	// LogSavePath 日志保存路径
	LogSavePath string `yaml:"log-save-fileurl"`
	// LogFile 日志文件名
	LogFile string `yaml:"log-file"`
	// TempPath 上传临时路径
	TempPath string `yaml:"temp-path"`
	// UploadSavePath 上传保存路径
	UploadSavePath string `yaml:"upload-save-path"`
	// IsReturnSussess 是否返回成功信息
	IsReturnSussess bool `yaml:"is-return-sussess"`
	// SoftDeleteRetentionTime 软删除笔记保留时间
	SoftDeleteRetentionTime string `yaml:"soft-delete-retention-time"`
	// HistoryKeepVersions 历史记录保留版本数，默认 100
	HistoryKeepVersions int `yaml:"history-keep-versions"`
	// UploadSessionTimeout 文件上传会话超时时间
	UploadSessionTimeout string `yaml:"upload-session-timeout"`
	// FileChunkSize 文件分片大小
	FileChunkSize string `yaml:"file-chunk-size"`

	// Worker Pool 配置
	WorkerPoolMaxWorkers int `yaml:"worker-pool-max-workers"`
	WorkerPoolQueueSize  int `yaml:"worker-pool-queue-size"`

	// Write Queue 配置
	WriteQueueCapacity int    `yaml:"write-queue-capacity"`
	WriteQueueTimeout  string `yaml:"write-queue-timeout"`
	WriteQueueIdleTime string `yaml:"write-queue-idle-time"`
}

// WebGUIConfig Web GUI 配置
type WebGUIConfig struct {
	FontSet string `yaml:"font-set" json:"fontSet"`
}

// TracerConfig 请求追踪配置
type TracerConfig struct {
	// Enabled 是否启用追踪
	Enabled bool `yaml:"enabled"`
	// Header 追踪 ID 请求头名称，默认 X-Trace-ID
	Header string `yaml:"header"`
}

// LoadConfig 从文件加载配置
// 返回配置实例和配置文件的绝对路径
func LoadConfig(f string) (*AppConfig, string, error) {
	realpath, err := filepath.Abs(f)
	if err != nil {
		return nil, "", err
	}
	realpath = filepath.Clean(realpath)

	c := new(AppConfig)
	c.File = realpath

	file, err := os.ReadFile(realpath)
	if err != nil {
		return nil, realpath, errors.Wrap(err, "read config file failed")
	}

	err = yaml.Unmarshal(file, c)
	if err != nil {
		return nil, realpath, errors.Wrap(err, "parse config file failed")
	}

	return c, realpath, nil
}

// Save 保存配置到文件
func (c *AppConfig) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return errors.Wrap(err, "marshal config failed")
	}

	err = os.WriteFile(c.File, data, 0644)
	if err != nil {
		return errors.Wrap(err, "write config file failed")
	}

	return nil
}

// GetWorkerPoolConfig 获取 Worker Pool 配置
func (c *AppConfig) GetWorkerPoolConfig() workerpool.Config {
	cfg := workerpool.DefaultConfig()

	if c.App.WorkerPoolMaxWorkers > 0 {
		cfg.MaxWorkers = c.App.WorkerPoolMaxWorkers
	}
	if c.App.WorkerPoolQueueSize > 0 {
		cfg.QueueSize = c.App.WorkerPoolQueueSize
	}

	return cfg
}

// GetWriteQueueConfig 获取 Write Queue 配置
func (c *AppConfig) GetWriteQueueConfig() writequeue.Config {
	cfg := writequeue.DefaultConfig()

	if c.App.WriteQueueCapacity > 0 {
		cfg.QueueCapacity = c.App.WriteQueueCapacity
	}
	if c.App.WriteQueueTimeout != "" {
		if timeout, err := util.ParseDuration(c.App.WriteQueueTimeout); err == nil {
			cfg.WriteTimeout = timeout
		}
	}
	if c.App.WriteQueueIdleTime != "" {
		if idleTime, err := util.ParseDuration(c.App.WriteQueueIdleTime); err == nil {
			cfg.IdleTimeout = idleTime
		}
	}

	return cfg
}

// GetTokenExpiry 获取 Token 过期时间
func (c *AppConfig) GetTokenExpiry() time.Duration {
	if c.Security.TokenExpiry != "" {
		if expiry, err := util.ParseDuration(c.Security.TokenExpiry); err == nil {
			return expiry
		}
	}
	return 7 * 24 * time.Hour // 默认 7 天
}
