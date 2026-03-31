package config

// DatabaseConfig database configuration
// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type                string `yaml:"type" default:"sqlite"`                                  // database type (mysql, postgres, sqlite) // 数据库类型 (mysql, postgres, sqlite)
	Path                string `yaml:"path" default:"storage/database/db.sqlite3"`             // SQLite database file path // SQLite 数据库文件路径
	UserName            string `yaml:"username"`                                               // database login username // 数据库登录用户名
	Password            string `yaml:"password"`                                               // database login password // 数据库登录密码
	Host                string `yaml:"host"`                                                   // database host // 数据库主机地址
	Port                int    `yaml:"port"`                                                   // database port // 数据库端口
	Name                string `yaml:"name"`                                                   // database name // 数据库名
	SSLMode             string `yaml:"ssl-mode"`                                               // SSL mode (postgres only) // SSL 模式 (仅限 postgres)
	TablePrefix         string `yaml:"table-prefix"`                                           // database table prefix // 数据库表前缀
	Schema              string `yaml:"schema"`                                                 // database schema (postgres only) // 数据库 Schema (仅限 postgres)
	AutoMigrate         bool   `yaml:"auto-migrate" default:"true"`                            // whether to enable automatic migration // 是否启用自动迁移
	Charset             string `yaml:"charset"`                                                // database charset // 数据库字符集
	ParseTime           bool   `yaml:"parse-time"`                                             // whether to parse time // 是否解析时间
	MaxIdleConns        int    `yaml:"max-idle-conns" default:"10"`                            // maximum number of idle connections // 最大闲置连接数，默认 10
	MaxOpenConns        int    `yaml:"max-open-conns" default:"100"`                           // maximum number of open connections // 最大打开连接数，默认 100
	ConnMaxLifetime     string `yaml:"conn-max-lifetime" default:"30m"`                        // maximum connection lifetime // 连接最大生命周期
	ConnMaxIdleTime     string `yaml:"conn-max-idle-time" default:"10m"`                        // maximum idle connection lifetime // 空闲连接最大生命周期
	EnableWriteQueue    *bool  `yaml:"enable-write-queue" default:"true"`                      // whether to enable write queue // 是否启用写队列，默认值为真
	MaxWriteConcurrency int    `yaml:"max-write-concurrency"`                                  // maximum concurrent write operations when write queue is disabled // 当 EnableWriteQueue 为 false 时，最大并发写入数，0 或负数表示不限制
	RunMode             string `yaml:"-"`                                                      // run mode (integrated from dao layer) // 运行模式 (从 dao 层整合)
}
