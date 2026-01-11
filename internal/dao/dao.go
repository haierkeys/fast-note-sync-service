package dao

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/haierkeys/fast-note-sync-service/pkg/writequeue"

	"github.com/glebarez/sqlite"
	"github.com/haierkeys/gormTracing"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"go.uber.org/zap"
)

// DatabaseConfig 数据库配置（用于依赖注入）
type DatabaseConfig struct {
	// Type 数据库类型
	Type string
	// Path SQLite 数据库文件路径
	Path string
	// UserName 用户名
	UserName string
	// Password 密码
	Password string
	// Host 主机
	Host string
	// Name 数据库名
	Name string
	// TablePrefix 表前缀
	TablePrefix string
	// AutoMigrate 是否启用自动迁移
	AutoMigrate bool
	// Charset 字符集
	Charset string
	// ParseTime 是否解析时间
	ParseTime bool
	// MaxIdleConns 最大闲置连接数，默认 10
	MaxIdleConns int
	// MaxOpenConns 最大打开连接数，默认 100
	MaxOpenConns int
	// ConnMaxLifetime 连接最大生命周期，支持格式：30m（分钟）、1h（小时），默认 30m
	ConnMaxLifetime string
	// ConnMaxIdleTime 空闲连接最大生命周期，支持格式：10m（分钟）、1h（小时），默认 10m
	ConnMaxIdleTime string
	// RunMode 运行模式（用于日志级别控制）
	RunMode string
}

// Dao 数据访问对象，封装数据库操作
type Dao struct {
	Db       *gorm.DB
	KeyDb    map[string]*gorm.DB
	ctx      context.Context
	onceKeys sync.Map
	mu       sync.RWMutex // 保护 KeyDb 的并发访问

	// 注入的依赖
	config        *DatabaseConfig
	logger        *zap.Logger
	writeQueueMgr *writequeue.Manager
}

// DaoOption 用于配置 Dao 的选项函数
type DaoOption func(*Dao)

// WithConfig 设置数据库配置
func WithConfig(cfg *DatabaseConfig) DaoOption {
	return func(d *Dao) {
		d.config = cfg
	}
}

// WithLogger 设置日志器
func WithLogger(logger *zap.Logger) DaoOption {
	return func(d *Dao) {
		d.logger = logger
	}
}

// WithWriteQueueManager 设置写队列管理器
func WithWriteQueueManager(wqm *writequeue.Manager) DaoOption {
	return func(d *Dao) {
		d.writeQueueMgr = wqm
	}
}

type daoDBCustomKey interface {
	GetKey(uid int64) string
}

// New 创建 Dao 实例（支持依赖注入）
// db: 主数据库连接
// ctx: 上下文
// opts: 可选配置项
func New(db *gorm.DB, ctx context.Context, opts ...DaoOption) *Dao {
	d := &Dao{
		Db:    db,
		ctx:   ctx,
		KeyDb: make(map[string]*gorm.DB),
	}

	// 应用选项
	for _, opt := range opts {
		opt(d)
	}

	// 如果没有提供 logger，使用 nop logger
	if d.logger == nil {
		d.logger = zap.NewNop()
	}

	return d
}

// Logger 获取日志器
func (d *Dao) Logger() *zap.Logger {
	if d.logger != nil {
		return d.logger
	}
	return zap.NewNop()
}

// Config 获取数据库配置
func (d *Dao) Config() *DatabaseConfig {
	return d.config
}

// WriteQueueManager 获取写队列管理器
func (d *Dao) WriteQueueManager() *writequeue.Manager {
	return d.writeQueueMgr
}

func (d *Dao) UseQueryWithFunc(f func(*gorm.DB), key ...string) *query.Query {
	db := d.UseKey(key...)
	if db == nil {
		keyName := "default"
		if len(key) > 0 {
			keyName = key[0]
		}
		panic(fmt.Sprintf("数据库实例为 nil (key=%s),请检查数据库配置和连接", keyName))
	}
	f(db)
	return query.Use(db)
}

func (d *Dao) UseQueryWithOnceFunc(f func(*gorm.DB), onceKey string, key ...string) *query.Query {
	db := d.UseKey(key...)
	if db == nil {
		keyName := "default"
		if len(key) > 0 {
			keyName = key[0]
		}
		panic(fmt.Sprintf("数据库实例为 nil (key=%s, onceKey=%s),请检查数据库配置和连接", keyName, onceKey))
	}
	if _, loaded := d.onceKeys.LoadOrStore(onceKey, true); !loaded {
		f(db)
	}
	return query.Use(db)
}

func (d *Dao) UseQuery(key ...string) *query.Query {
	db := d.UseKey(key...)
	if db == nil {
		keyName := "default"
		if len(key) > 0 {
			keyName = key[0]
		}
		panic(fmt.Sprintf("数据库实例为 nil (key=%s),请检查数据库配置和连接", keyName))
	}
	return query.Use(db)
}

func (d *Dao) UseKey(key ...string) *gorm.DB {
	var db *gorm.DB
	if len(key) > 0 {
		db = d.UseDb(key[0])
	} else {
		db = d.Db
	}
	return db
}

func (d *Dao) UserDB(uid int64) *gorm.DB {
	key := "user_" + strconv.FormatInt(uid, 10)
	b := d.UseKey(key)
	return b
}

// getDBConfig 获取数据库配置
func (d *Dao) getDBConfig() DatabaseConfig {
	if d.config != nil {
		return *d.config
	}
	return DatabaseConfig{}
}

func (d *Dao) UseDb(key string) *gorm.DB {
	// 使用读锁检查是否已存在
	d.mu.RLock()
	if db, ok := d.KeyDb[key]; ok {
		d.mu.RUnlock()
		return db
	}
	d.mu.RUnlock()

	// 获取配置
	c := d.getDBConfig()

	if c.Type == "mysql" {
		c.Name = c.Name + "_" + key
	} else if c.Type == "sqlite" {
		ext := filepath.Ext(c.Path)
		c.Path = c.Path[:len(c.Path)-len(ext)] + "_" + key + ext
	}

	dbNew, err := NewDBEngineWithConfig(c, d.Logger())
	if err != nil {
		d.Logger().Error("UseDb failed", zap.String("key", key), zap.Error(err))
		return nil
	}

	// 使用写锁存储
	d.mu.Lock()
	// 双重检查
	if existingDb, ok := d.KeyDb[key]; ok {
		d.mu.Unlock()
		// 关闭新创建的连接
		if sqlDB, err := dbNew.DB(); err == nil {
			sqlDB.Close()
		}
		return existingDb
	}
	d.KeyDb[key] = dbNew
	d.mu.Unlock()

	return dbNew
}

// NewDBEngineWithConfig 创建数据库引擎（支持依赖注入）
// 函数名: NewDBEngineWithConfig
// 函数使用说明: 根据配置创建并初始化 GORM 数据库引擎,配置连接池参数和日志模式。
// 参数说明:
//   - c DatabaseConfig: 数据库配置
//   - zapLogger *zap.Logger: 日志器（可选，为 nil 时使用默认日志）
//
// 返回值说明:
//   - *gorm.DB: 数据库连接实例
//   - error: 出错时返回错误
func NewDBEngineWithConfig(c DatabaseConfig, zapLogger *zap.Logger) (*gorm.DB, error) {

	var db *gorm.DB
	var err error

	db, err = gorm.Open(useDiaWithConfig(c), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   c.TablePrefix, // 表名前缀，`User` 的表名应该是 `t_users`
			SingularTable: true,          // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `t_user`
		},
	})
	if err != nil {
		return nil, err
	}

	// 根据运行模式设置日志级别
	if c.RunMode == "debug" {
		db.Config.Logger = logger.Default.LogMode(logger.Info)
	} else {
		db.Config.Logger = logger.Default.LogMode(logger.Silent)
	}

	// 获取通用数据库对象 sql.DB ，然后使用其提供的功能
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数（带默认值）
	// MaxIdleConns: 默认 10
	maxIdleConns := c.MaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = 10
	}
	sqlDB.SetMaxIdleConns(maxIdleConns)

	// MaxOpenConns: 默认 100
	maxOpenConns := c.MaxOpenConns
	if maxOpenConns == 0 {
		maxOpenConns = 100
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)

	// ConnMaxLifetime: 默认 30 分钟
	connMaxLifetime := 30 * time.Minute
	if c.ConnMaxLifetime != "" {
		if parsed, err := util.ParseDuration(c.ConnMaxLifetime); err == nil {
			connMaxLifetime = parsed
		}
	}
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	// ConnMaxIdleTime: 默认 10 分钟
	connMaxIdleTime := 10 * time.Minute
	if c.ConnMaxIdleTime != "" {
		if parsed, err := util.ParseDuration(c.ConnMaxIdleTime); err == nil {
			connMaxIdleTime = parsed
		}
	}
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	_ = db.Use(&gormTracing.OpentracingPlugin{})

	return db, nil

}

// useDiaWithConfig 获取数据库方言（支持依赖注入）
// 函数名: useDiaWithConfig
// 函数使用说明: 根据数据库配置返回对应的 GORM 方言(MySQL 或 SQLite)。
// 参数说明:
//   - c DatabaseConfig: 数据库配置
//
// 返回值说明:
//   - gorm.Dialector: GORM 数据库方言
func useDiaWithConfig(c DatabaseConfig) gorm.Dialector {
	if c.Type == "mysql" {
		return mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&parseTime=%t&loc=Local",
			c.UserName,
			c.Password,
			c.Host,
			c.Name,
			c.Charset,
			c.ParseTime,
		))
	} else if c.Type == "sqlite" {

		filepath.Dir(c.Path)

		if !fileurl.IsExist(c.Path) {
			fileurl.CreatePath(c.Path, os.ModePerm)
		}

		absDb, err := filepath.Abs(c.Path)
		if err != nil {
			panic(err)
		}
		dbSlash := "/" + strings.TrimPrefix(filepath.ToSlash(absDb), "/")
		connStr := "file://" + dbSlash + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(10000)"
		// connStr = "file:///" + dbSlash + "?_foreign_keys=1&_journal_mode=WAL&_synchronous=NORMAL&_busy_timeout=10000&_mutex=no"
		// connStr := c.Path + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(10000)"

		return sqlite.Open(connStr)
	}
	return nil

}

// WithRetry 封装数据库操作的重试逻辑，主要用于解决 SQLite "database is locked" 问题
func (d *Dao) WithRetry(fn func() error) error {
	maxRetries := 5
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		// 检查是否为 SQLite 锁定错误
		errStr := err.Error()
		if strings.Contains(errStr, "database is locked") || strings.Contains(errStr, "SQLITE_BUSY") {
			// 指数退避或固定延迟
			time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
			continue
		}
		return err // 其他错误直接返回
	}
	return err
}

// ExecuteWrite 执行写操作（通过写队列串行化）
// 写操作会被串行化执行，同一用户的写操作按 FIFO 顺序处理
// ctx: 上下文，用于超时和取消控制
// uid: 用户 ID，用于确定写队列
// fn: 写操作函数，接收用户数据库连接
// 返回值: 写操作的错误
// 注意: 必须通过 WithWriteQueueManager 注入写队列管理器
func (d *Dao) ExecuteWrite(ctx context.Context, uid int64, r daoDBCustomKey, fn func(*gorm.DB) error) error {
	if d.writeQueueMgr == nil {
		return fmt.Errorf("writeQueueMgr is nil, must inject via WithWriteQueueManager")
	}

	return d.writeQueueMgr.Execute(ctx, uid, func() error {
		db := d.UseKey(r.GetKey(uid))
		if db == nil {
			return fmt.Errorf("database connection is nil (uid=%d)", uid)
		}
		return fn(db.WithContext(ctx))
	})
}

// ExecuteRead 执行读操作（直接执行，不经过写队列）
// 读操作不需要串行化，可以并发执行
// ctx: 上下文，用于超时和取消控制
// uid: 用户 ID，用于获取用户数据库连接
// fn: 读操作函数，接收用户数据库连接
// 返回值: 读操作的错误
func (d *Dao) ExecuteRead(ctx context.Context, uid int64, r daoDBCustomKey, fn func(*gorm.DB) error) error {
	db := d.UseKey(r.GetKey(uid))
	if db == nil {
		return fmt.Errorf("database connection is nil (uid=%d)", uid)
	}
	return fn(db.WithContext(ctx))
}

// ExecuteWriteWithRetry 执行写操作（通过写队列串行化，带重试）
// 结合写队列和重试逻辑，用于处理 SQLite 并发写入问题
// ctx: 上下文，用于超时和取消控制
// uid: 用户 ID，用于确定写队列
// fn: 写操作函数，接收用户数据库连接
// 返回值: 写操作的错误
func (d *Dao) ExecuteWriteWithRetry(ctx context.Context, uid int64, r daoDBCustomKey, fn func(*gorm.DB) error) error {
	return d.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		return d.WithRetry(func() error {
			return fn(db)
		})
	})
}

// getDbKeyByModel 根据模型名称获取对应的数据库连接 Key
func (d *Dao) getDbKeyByModelName(uid int64, modelKey string) string {
	if uid <= 0 {
		return "" // 主数据库
	}

	switch modelKey {
	case "User":
		return "" // 用户表在主库
	case "Note":
		return NewNoteRepository(d).(daoDBCustomKey).GetKey(uid)
	case "File":
		return NewFileRepository(d).(daoDBCustomKey).GetKey(uid)
	case "Setting":
		return NewSettingRepository(d).(daoDBCustomKey).GetKey(uid)
	case "NoteHistory":
		return NewNoteHistoryRepository(d).(daoDBCustomKey).GetKey(uid)
	case "Vault":
		return NewVaultRepository(d).(daoDBCustomKey).GetKey(uid)
	default:
		return ""
	}
}

func (d *Dao) AutoMigrate(uid int64, modelKey string) error {
	// 1. 如果 modelKey 为空，说明是“全量迁移”，按模型分别路由迁移
	if modelKey == "" {
		models := []string{"User", "Note", "File", "Setting", "NoteHistory", "Vault"}
		for _, m := range models {
			if err := d.AutoMigrate(uid, m); err != nil {
				return err
			}
		}
		return nil
	}

	dbKey := d.getDbKeyByModelName(uid, modelKey)
	b := d.UseKey(dbKey)

	if b == nil {
		return fmt.Errorf("database connection is nil for model %s (uid=%d, dbKey=%s)", modelKey, uid, dbKey)
	}
	return model.AutoMigrate(b, modelKey)
}

// user 获取用户查询对象（内部方法）
func (d *Dao) user() *query.Query {
	return d.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "User")
	}, "user#user")
}

// GetAllUserUIDs 获取所有用户的UID
// 返回值说明:
//   - []int64: 用户UID列表
//   - error: 出错时返回错误
func (d *Dao) GetAllUserUIDs() ([]int64, error) {
	var uids []int64
	u := d.user().User
	err := u.WithContext(d.ctx).Select(u.UID).Where(u.IsDeleted.Eq(0)).Scan(&uids)
	if err != nil {
		return nil, err
	}
	return uids, nil
}
