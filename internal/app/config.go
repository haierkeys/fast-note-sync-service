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

	Server       config.ServerConfig     `yaml:"server"`
	App          config.AppSettings      `yaml:"app"`
	Security     config.SecurityConfig   `yaml:"security"`
	Database     config.DatabaseConfig   `yaml:"database"`
	UserDatabase config.DatabaseConfig   `yaml:"user-database"`
	Log          config.LogConfig        `yaml:"log"`
	User         config.UserConfig       `yaml:"user"`
	Tracer       config.TracerConfig     `yaml:"tracer"`
	ShortLink    config.ShortLinkConfig  `yaml:"short-link"`
	Storage      config.StorageConfig    `yaml:"storage"`
	Git          config.GitConfig        `yaml:"git"`
	WebGUI       config.WebGUIConfig     `yaml:"webgui"`
	Ngrok        config.NgrokConfig      `yaml:"ngrok"`
	Cloudflare   config.CloudflareConfig `yaml:"cloudflare"`
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
