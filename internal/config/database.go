package config

// DatabaseConfig database configuration
// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	// Type 数据库类型 (mysql, postgres, sqlite)
	Type string `yaml:"type" default:"sqlite"`
	// Path SQLite 数据库 file path
	Path string `yaml:"path" default:"storage/database/db.sqlite3"`
	// UserName 用户名
	UserName string `yaml:"username"`
	// Password 密码
	Password string `yaml:"password"`
	// Host 主机
	Host string `yaml:"host"`
	// Port 端口
	Port int `yaml:"port"`
	// Name 数据库名
	Name string `yaml:"name"`
	// SSLMode SSL 模式 (仅限 postgres)
	SSLMode string `yaml:"ssl-mode"`
	// TablePrefix 表前缀
	TablePrefix string `yaml:"table-prefix"`
	// AutoMigrate 是否启用自动迁移
	AutoMigrate bool `yaml:"auto-migrate"`
	// Charset 字符集
	Charset string `yaml:"charset"`
	// ParseTime 是否解析时间
	ParseTime bool `yaml:"parse-time"`
	// MaxIdleConns 最大闲置连接数，默认 10
	MaxIdleConns int `yaml:"max-idle-conns" default:"10"`
	// MaxOpenConns 最大打开连接数，默认 100
	MaxOpenConns int `yaml:"max-open-conns" default:"100"`
	// ConnMaxLifetime 连接最大生命周期
	ConnMaxLifetime string `yaml:"conn-max-lifetime" default:"30m"`
	// ConnMaxIdleTime 空闲连接最大生命周期
	ConnMaxIdleTime string `yaml:"conn-max-idle-time" default:"10m"`
	// EnableWriteQueue 是否启用写队列，默认值为真
	EnableWriteQueue *bool `yaml:"enable-write-queue" default:"true"`
	// MaxWriteConcurrency 当 EnableWriteQueue 为 false 时，最大并发写入数，0 或负数表示不限制
	MaxWriteConcurrency int `yaml:"max-write-concurrency"`
	// RunMode 运行模式 (从 dao 层整合)
	RunMode string `yaml:"-"`
}
