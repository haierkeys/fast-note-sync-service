// Package app provides application container, encapsulates all dependencies and services
// Package app 提供应用容器，封装所有依赖和服务
package app

import (
	"os"
	"path/filepath"
	"time"

	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/haierkeys/fast-note-sync-service/pkg/workerpool"
	"github.com/haierkeys/fast-note-sync-service/pkg/writequeue"

	"github.com/creasty/defaults"
	"github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// AppConfig application configuration
// AppConfig 应用配置
type AppConfig struct {
	File string `yaml:"-"` // config file path, not serialized
	// 配置文件路径, 不序列化

	Server   ServerConfig   `yaml:"server"`
	App      AppSettings    `yaml:"app"`
	Security SecurityConfig `yaml:"security"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
	User     UserConfig     `yaml:"user"`
	Tracer   TracerConfig   `yaml:"tracer"`

	Storage config.StorageConfig `yaml:"storage"`
	WebGUI  WebGUIConfig         `yaml:"webgui"`
}

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
}

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
}

// DatabaseConfig database configuration
// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	// Type database type
	// Type 数据库类型
	Type string `yaml:"type" default:"sqlite"`
	// Path SQLite database file path
	// Path SQLite 数据库文件路径
	Path string `yaml:"path" default:"storage/database/db.sqlite3"`
	// UserName username
	// UserName 用户名
	UserName string `yaml:"username"`
	// Password password
	// Password 密码
	Password string `yaml:"password"`
	// Host host
	// Host 主机
	Host string `yaml:"host"`
	// Name database name
	// Name 数据库名
	Name string `yaml:"name"`
	// TablePrefix table prefix
	// TablePrefix 表前缀
	TablePrefix string `yaml:"table-prefix"`
	// AutoMigrate whether to enable auto migration
	// AutoMigrate 是否启用自动迁移
	AutoMigrate bool `yaml:"auto-migrate"`
	// Charset charset
	// Charset 字符集
	Charset string `yaml:"charset"`
	// ParseTime whether to parse time
	// ParseTime 是否解析时间
	ParseTime bool `yaml:"parse-time"`
	// MaxIdleConns maximum idle connections, default 10
	// MaxIdleConns 最大闲置连接数，默认 10
	MaxIdleConns int `yaml:"max-idle-conns" default:"10"`
	// MaxOpenConns maximum open connections, default 100
	// MaxOpenConns 最大打开连接数，默认 100
	MaxOpenConns int `yaml:"max-open-conns" default:"100"`
	// ConnMaxLifetime maximum connection lifetime, supports format: 30m (minutes), 1h (hours), default 30m
	// ConnMaxLifetime 连接最大生命周期，支持格式：30m（分钟）、1h（小时），默认 30m
	ConnMaxLifetime string `yaml:"conn-max-lifetime" default:"30m"`
	// ConnMaxIdleTime maximum idle connection lifetime, supports format: 10m (minutes), 1h (hours), default 10m
	// ConnMaxIdleTime 空闲连接最大生命周期，支持格式：10m（分钟）、1h（小时），默认 10m
	ConnMaxIdleTime string `yaml:"conn-max-idle-time" default:"10m"`
}

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

// AppSettings application settings
// AppSettings 应用设置
type AppSettings struct {
	// DefaultPageSize default page size
	// DefaultPageSize 默认页面大小
	DefaultPageSize int `yaml:"default-page-size" default:"10"`
	// MaxPageSize maximum page size
	// MaxPageSize 最大页面大小
	MaxPageSize int `yaml:"max-page-size" default:"100"`
	// DefaultContextTimeout default context timeout duration
	// DefaultContextTimeout 默认上下文超时时间
	DefaultContextTimeout int `yaml:"default-context-timeout" default:"60"`
	// LogSavePath log save path
	// LogSavePath 日志保存路径
	LogSavePath string `yaml:"log-save-fileurl"`
	// LogFile log filename
	// LogFile 日志文件名
	LogFile string `yaml:"log-file"`
	// TempPath upload temporary path
	// TempPath 上传临时路径
	TempPath string `yaml:"temp-path" default:"storage/temp"`
	// IsReturnSussess whether to return success info
	// IsReturnSussess 是否返回成功信息
	IsReturnSussess bool `yaml:"is-return-sussess" default:"false"`
	// SoftDeleteRetentionTime retention time for soft deleted notes
	// SoftDeleteRetentionTime 软删除笔记保留时间
	SoftDeleteRetentionTime string `yaml:"soft-delete-retention-time" default:"7d"`
	// HistoryKeepVersions number of historical versions to keep, default 100
	// HistoryKeepVersions 历史记录保留版本数，默认 100
	HistoryKeepVersions int `yaml:"history-keep-versions" default:"100"`
	// HistorySaveDelay historical record save delay time, supports format: 10s (seconds), 1m (minutes), default 10s
	// HistorySaveDelay 历史记录保存延迟时间，支持格式：10s（秒）、1m（分钟），默认 10s
	HistorySaveDelay string `yaml:"history-save-delay" default:"10s"`
	// UploadSessionTimeout file upload session timeout duration
	// UploadSessionTimeout 文件上传会话超时时间
	UploadSessionTimeout string `yaml:"upload-session-timeout" default:"1d"`
	// FileChunkSize file chunk size
	// FileChunkSize 文件分片大小
	FileChunkSize string `yaml:"file-chunk-size" default:"512KB"`
	// DownloadSessionTimeout file chunk download timeout duration
	// DownloadSessionTimeout 文件分片下载超时时间
	DownloadSessionTimeout string `yaml:"download-session-timeout" default:"1h"`
	// DefaultAPIFolder API default note folder, automatically add this prefix when note path does not contain "/"
	// DefaultAPIFolder API默认笔记文件夹，当笔记路径不包含"/"时自动添加此前缀
	// DefaultAPIFolder string `yaml:"default-api-folder" default:""`

	// Worker Pool configurations
	// Worker Pool 配置
	WorkerPoolMaxWorkers int `yaml:"worker-pool-max-workers" default:"100"`
	WorkerPoolQueueSize  int `yaml:"worker-pool-queue-size" default:"1000"`

	// Write Queue configurations
	// Write Queue 配置
	WriteQueueCapacity int    `yaml:"write-queue-capacity" default:"100"`
	WriteQueueTimeout  string `yaml:"write-queue-timeout" default:"30s"`
	WriteQueueIdleTime string `yaml:"write-queue-idle-time" default:"10m"`

	// WebSocket configurations
	// WebSocket 配置
	WebSocketReadMaxPayloadSize  string `yaml:"ws-read-max-payload-size" default:"128MB"`
	WebSocketWriteMaxPayloadSize string `yaml:"ws-write-max-payload-size" default:"128MB"`
}

// WebGUIConfig Web GUI configuration
// WebGUIConfig Web GUI 配置
type WebGUIConfig struct {
	FontSet string `yaml:"font-set" json:"fontSet" default:""`
}

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

// LoadConfig loads configuration from file
// LoadConfig 从文件加载配置
// returns configuration instance and absolute path of configuration file
// 返回配置实例和配置文件的绝对路径
func LoadConfig(f string) (*AppConfig, string, error) {
	realpath, err := filepath.Abs(f)
	if err != nil {
		return nil, "", err
	}
	realpath = filepath.Clean(realpath)

	c := new(AppConfig)
	c.File = realpath

	// Set default values
	// 设置默认值
	if err := defaults.Set(c); err != nil {
		return nil, realpath, errors.Wrap(err, "set default config failed")
	}

	file, err := os.ReadFile(realpath)
	if err != nil {
		return nil, realpath, errors.Wrap(err, "read config file failed")
	}

	err = yaml.Unmarshal(file, c)
	if err != nil {
		return nil, realpath, errors.Wrap(err, "parse config file failed")
	}

	// Set default values again to fill fields that exist in YAML but have empty values
	// 再次设置默认值，以填充 YAML 中存在但值为空的字段
	// defaults.Set filled only when the field is the zero value of the type
	// defaults.Set 只有在字段为该类型的零值时才会填充
	if err := defaults.Set(c); err != nil {
		return nil, realpath, errors.Wrap(err, "re-set default config failed")
	}

	return c, realpath, nil
}

// Save saves configuration to file
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

// GetWorkerPoolConfig gets Worker Pool configuration
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

// GetWriteQueueConfig gets Write Queue configuration
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

// GetTokenExpiry gets Token expiry duration
// GetTokenExpiry 获取 Token 过期时间
func (c *AppConfig) GetTokenExpiry() time.Duration {
	if expiry, err := util.ParseDuration(c.Security.TokenExpiry); err == nil {
		return expiry
	}
	return 365 * 24 * time.Hour // Theoretically will not reach here because of default values
	// 理论上不会走到这里，因为有默认值
}

// GetShareTokenExpiry gets share Token expiry duration
// GetShareTokenExpiry 获取分享 Token 过期时间
func (c *AppConfig) GetShareTokenExpiry() time.Duration {
	if expiry, err := util.ParseDuration(c.Security.ShareTokenExpiry); err == nil {
		return expiry
	}
	return 30 * 24 * time.Hour // Theoretically will not reach here because of default values
	// 理论上不会走到这里，因为有默认值
}
