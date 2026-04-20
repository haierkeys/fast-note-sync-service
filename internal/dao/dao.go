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

	"github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/haierkeys/fast-note-sync-service/pkg/writequeue"

	"github.com/glebarez/sqlite"
	"github.com/haierkeys/gormTracing"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

// DatabaseConfig database configuration (for dependency injection)
// DatabaseConfig 数据库配置（用于依赖注入）
// DatabaseConfig is now imported from internal/config

type dbEntry struct {
	db       *gorm.DB
	lastUsed time.Time
}

// Dao data access object, encapsulates database operations
// Dao 数据访问对象，封装数据库操作
type Dao struct {
	Db       *gorm.DB
	KeyDb    map[string]*dbEntry
	ctx      context.Context
	onceKeys sync.Map
	mu       sync.RWMutex // protects concurrent access to KeyDb // 保护 KeyDb 的并发访问

	poolSemaphores sync.Map // map[string]*semaphore.Weighted 针对不同配置的并发控制

	// 注入的依赖
	config        *config.DatabaseConfig
	userConfig    *config.DatabaseConfig
	logger        *zap.Logger
	writeQueueMgr *writequeue.Manager
}

// DaoOption option function for configuring Dao
// DaoOption 用于配置 Dao 的选项函数
type DaoOption func(*Dao)

// WithConfig sets database configuration
// WithConfig 设置数据库配置
func WithConfig(cfg *config.DatabaseConfig) DaoOption {
	return func(d *Dao) {
		d.config = cfg
	}
}

// WithUserDatabaseConfig sets user database configuration
// WithUserDatabaseConfig 设置用户数据库配置
func WithUserDatabaseConfig(cfg *config.DatabaseConfig) DaoOption {
	return func(d *Dao) {
		d.userConfig = cfg
	}
}

// WithLogger sets logger
// WithLogger 设置日志器
func WithLogger(logger *zap.Logger) DaoOption {
	return func(d *Dao) {
		d.logger = logger
	}
}

// WithWriteQueueManager sets write queue manager
// WithWriteQueueManager 设置写队列管理器
func WithWriteQueueManager(wqm *writequeue.Manager) DaoOption {
	return func(d *Dao) {
		d.writeQueueMgr = wqm
	}
}

type daoDBCustomKey interface {
	GetKey(uid int64) string
}

// ModelConfig describes the database routing information for a model
// ModelConfig 描述一个模型的数据库路由信息
type ModelConfig struct {
	Name        string
	RepoFactory func(d *Dao) daoDBCustomKey
	IsMainDB    bool
}

var modelConfigs []ModelConfig

// RegisterModel called by each Repository file in init()
// RegisterModel 供各 Repository 文件在 init() 中调用
func RegisterModel(cfg ModelConfig) {
	modelConfigs = append(modelConfigs, cfg)
}

// New creates Dao instance (supports dependency injection)
// db: Main database connection // db: 主数据库连接
// ctx: Context // ctx: 上下文
// opts: Optional configuration items // opts: 可选配置项
func New(db *gorm.DB, ctx context.Context, opts ...DaoOption) *Dao {
	d := &Dao{
		Db:    db,
		ctx:   ctx,
		KeyDb: make(map[string]*dbEntry),
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

// Logger gets the logger
// Logger 获取日志器
func (d *Dao) Logger() *zap.Logger {
	if d.logger != nil {
		return d.logger
	}
	return zap.NewNop()
}

// Config gets the database configuration
// Config 获取数据库配置
func (d *Dao) Config() *config.DatabaseConfig {
	return d.config
}

// WriteQueueManager gets the write queue manager
// WriteQueueManager 获取写队列管理器
func (d *Dao) WriteQueueManager() *writequeue.Manager {
	return d.writeQueueMgr
}

// QueryWithOnceInit 执行带有单次初始化逻辑的数据库查询
// QueryWithOnceInit executes a database query with once-init logic.
// 参数说明:
//   - f func(*gorm.DB): 初始化函数，仅在 onceKey 首次出现时执行 (如 AutoMigrate)
//   - onceKey string: 用于确保初始化逻辑仅执行一次的唯一标识
//   - key ...string: 数据库连接标识（可变参数）。不传或为空时使用主数据库；传入时用于路由到特定租户/用户库
//
// Parameters:
//   - f func(*gorm.DB): Initialization function, executed only the first time onceKey is encountered (e.g., AutoMigrate).
//   - onceKey string: Unique identifier to ensure initialization logic runs only once.
//   - key ...string: Database connection identifier (variadic). Uses main DB if omitted/empty; uses provided key for tenant/user DB routing.
func (d *Dao) QueryWithOnceInit(f func(*gorm.DB), onceKey string, key ...string) *query.Query {
	db := d.ResolveDB(key...)
	if db == nil {
		keyName := "default"
		if len(key) > 0 {
			keyName = key[0]
		}
		panic(fmt.Sprintf("数据库 instance 为 nil (key=%s, onceKey=%s),请检查数据库配置和连接", keyName, onceKey))
	}

	// Construct library-level unique initialization Key
	// 构造库级唯一的初始化 Key
	// If a key is provided, it indicates a tenant library, and the key needs to be attached to ensure independent initialization of each tenant library
	// 如果提供了 key，说明是租户库，需附加 key 以保证每个租户库独立初始化
	actualOnceKey := onceKey
	if len(key) > 0 && key[0] != "" {
		actualOnceKey = onceKey + "@" + key[0]
	}

	if _, loaded := d.onceKeys.LoadOrStore(actualOnceKey, true); !loaded {
		f(db)
	}
	return query.Use(db)
}

// CleanupConnections cleans up idle database connections
// CleanupConnections 清理闲置数据库连接
func (d *Dao) CleanupConnections(maxIdle time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now()
	for k, v := range d.KeyDb {
		if now.Sub(v.lastUsed) > maxIdle {
			delete(d.KeyDb, k)
			if sqlDB, err := v.db.DB(); err == nil {
				sqlDB.Close()
			}
			d.Logger().Info("cleaned up idle DB connection", zap.String("key", k))
		}
	}
}

func (d *Dao) ResolveDB(key ...string) *gorm.DB {
	if len(key) == 0 || key[0] == "" {
		return d.Db
	}
	return d.GetOrCreateDB(key[0])
}

// resolveConfig gets database configuration
// key: Database identifier, tries to get user DB config if non-empty // key: 数据库标识，如果非空则尝试获取用户数据库配置
func (d *Dao) resolveConfig(key string) config.DatabaseConfig {
	var cfg config.DatabaseConfig
	// If targeted at specific Key (usually user DB) and independent UserDatabase type is configured
	// 如果是针对特定 Key (通常为用户库) 且配置了独立的 UserDatabase 类型
	if key != "" && d.userConfig != nil && d.userConfig.Type != "" {
		cfg = *d.userConfig
	} else if d.config != nil {
		// Otherwise inherits main database configuration (Fallback mode)
		// 否则继承主数据库配置 (Fallback 模式)
		cfg = *d.config
	}

	// Final fallback logic: if no type is configured globally, force default to sqlite
	// 最终回退逻辑：如果全局均未配置类型，强制默认为 sqlite
	if cfg.Type == "" {
		cfg.Type = "sqlite"
		if cfg.Path == "" {
			cfg.Path = "storage/database/db.sqlite3"
		}
	}
	return cfg
}

func (d *Dao) GetOrCreateDB(key string) *gorm.DB {
	// Use read lock to check if already exists
	// 使用读锁检查是否已存在
	d.mu.RLock()
	if entry, ok := d.KeyDb[key]; ok {
		entry.lastUsed = time.Now()
		d.mu.RUnlock()
		return entry.db
	}
	d.mu.RUnlock()

	// Get configuration
	// 获取配置
	c := d.resolveConfig(key)

	if (c.Type == "postgres") && key != "" {
		// PostgreSQL: Uniform mapping to user_<uid> Schema, ignoring specific Repo prefixes
		// PostgreSQL: 统一映射到 user_<uid> Schema，忽略具体的 Repo 前缀
		schemaName, ok := d.extractUserSchema(key)
		if !ok {
			schemaName = key // Fallback logic, use original key if resolution fails // 回退逻辑，如果无法解析则使用原始 key
		}

		if err := d.ensurePostgresSchema(schemaName); err != nil {
			d.Logger().Error("ensurePostgresSchema failed", zap.String("schema", schemaName), zap.Error(err))
			return nil
		}
		c.Schema = schemaName
		c.TablePrefix = "" // PostgreSQL 下清空前缀，改用 Schema
	} else if (c.Type == "mysql") && key != "" {
		// MySQL: Uniform mapping to user_<uid> database, implementing tenant-level DB isolation
		// MySQL: 统一映射到 user_<uid> 数据库，实现租户级库隔离
		dbName, ok := d.extractUserSchema(key)
		if !ok {
			dbName = key // Fallback logic, use original key if resolution fails // 回退逻辑，如果无法解析则使用原始 key
		}

		if err := d.ensureMysqlDatabase(dbName); err != nil {
			d.Logger().Error("ensureMysqlDatabase failed", zap.String("db", dbName), zap.Error(err))
			return nil
		}
		c.Name = dbName
		c.TablePrefix = "" // Clear prefix under MySQL, use database routing instead // MySQL 下清空前缀，改用库路由
	} else if c.Type == "sqlite" && key != "" {
		// SQLite: Maintain multi-file isolation mode (using full key as filename suffix)
		// SQLite: 维持多文件隔离模式 (使用完整的 key 作为文件名后缀)
		ext := filepath.Ext(c.Path)
		c.Path = c.Path[:len(c.Path)-len(ext)] + "_" + key + ext
	}

	dbNew, err := NewEngine(c, d.Logger())
	if err != nil {
		d.Logger().Error("GetOrCreateDB failed", zap.String("key", key), zap.Error(err))
		return nil
	}

	// Use write lock for storage
	// 使用写锁存储
	d.mu.Lock()
	defer d.mu.Unlock()
	// Double check
	// 双重检查
	if existingEntry, ok := d.KeyDb[key]; ok {
		// Close the newly created connection
		// 关闭新创建的连接
		if sqlDB, err := dbNew.DB(); err == nil {
			sqlDB.Close()
		}
		existingEntry.lastUsed = time.Now()
		return existingEntry.db
	}

	// Check cache quantity limit; if more than 100 connections, clean up the least recently used one
	// 检查缓存数量限制，如果超过 100 个连接，清理最久未使用的
	if len(d.KeyDb) >= 100 {
		var oldestKey string
		var oldestTime time.Time
		for k, v := range d.KeyDb {
			if oldestKey == "" || v.lastUsed.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.lastUsed
			}
		}
		if oldestKey != "" {
			oldEntry := d.KeyDb[oldestKey]
			delete(d.KeyDb, oldestKey)
			if sqlDB, err := oldEntry.db.DB(); err == nil {
				sqlDB.Close()
			}
			d.Logger().Info("evicted oldest DB connection", zap.String("key", oldestKey))
		}
	}

	d.KeyDb[key] = &dbEntry{
		db:       dbNew,
		lastUsed: time.Now(),
	}

	return dbNew
}

// NewEngine 创建数据库引擎（支持依赖注入）
// 函数名: NewEngine
// 函数使用说明: 根据配置创建并初始化 GORM 数据库引擎,配置连接池参数和日志模式。
// 参数说明:
//   - c DatabaseConfig: 数据库配置
//   - zapLogger *zap.Logger: 日志器（可选，为 nil 时使用默认日志）
//
// 返回值说明:
//   - *gorm.DB: 数据库连接实例
//   - error: 出错时返回错误
func NewEngine(c config.DatabaseConfig, zapLogger *zap.Logger) (*gorm.DB, error) {
	// 如果没有指定类型，则默认为 sqlite
	if c.Type == "" {
		c.Type = "sqlite"
		if c.Path == "" {
			c.Path = "storage/database/db.sqlite3"
		}
	}

	var db *gorm.DB
	var err error

	db, err = gorm.Open(getDialector(c), &gorm.Config{
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

// getDialector 获取数据库方言（支持依赖注入）
// 函数名: getDialector
// 函数使用说明: 根据数据库配置返回对应的 GORM 方言(MySQL 或 SQLite)。
// 参数说明:
//   - c DatabaseConfig: 数据库配置
//
// 返回值说明:
//   - gorm.Dialector: GORM 数据库方言
func getDialector(c config.DatabaseConfig) gorm.Dialector {
	if c.Type == "mysql" {
		host := c.Host
		if c.Port != 0 && !strings.Contains(host, ":") {
			host = fmt.Sprintf("%s:%d", host, c.Port)
		}
		return mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&parseTime=%t&loc=Local",
			c.UserName,
			c.Password,
			host,
			c.Name,
			c.Charset,
			c.ParseTime,
		))
	} else if c.Type == "postgres" {
		if c.Port == 0 {
			c.Port = 5432
		}
		if c.SSLMode == "" {
			c.SSLMode = "disable"
		}
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
			c.Host,
			c.UserName,
			c.Password,
			c.Name,
			c.Port,
			c.SSLMode,
		)
		if c.Schema != "" {
			dsn = fmt.Sprintf("%s search_path=%s", dsn, c.Schema)
		}
		return postgres.Open(dsn)
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

// WithRetry encapsulates retry logic for database operations, mainly to solve SQLite "database is locked" issues
// WithRetry 封装数据库操作的重试逻辑，主要用于解决 SQLite "database is locked" 问题
func (d *Dao) WithRetry(fn func() error) error {
	maxRetries := 5
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		// Check if it's a SQLite lock error
		// 检查是否为 SQLite 锁定错误
		errStr := err.Error()
		if strings.Contains(errStr, "database is locked") || strings.Contains(errStr, "SQLITE_BUSY") {
			// Exponential backoff or fixed delay // 指数退避或固定延迟
			time.Sleep(time.Duration(100*(i+1)) * time.Millisecond)
			continue
		}
		return err // 其他错误直接返回
	}
	return err
}

// ExecuteWrite executes write operation (serialized through write queue)
// ExecuteWrite 执行写操作（通过写队列串行化）
// Write operations will be executed serially, and write operations of the same user will be processed in FIFO order
// 写操作会被串行化执行，同一用户的写操作按 FIFO 顺序处理
// ctx: Context for timeout and cancellation control // ctx: 上下文，用于超时和取消控制
// uid: User ID, used to determine write queue // uid: 用户 ID，用于确定写队列
// fn: Write operation function, receiving user database connection // fn: 写操作函数，接收用户数据库连接
// Return value: Error of the write operation // 返回值: 写操作的错误
// Note: Write queue manager must be injected via WithWriteQueueManager // 注意: 必须通过 WithWriteQueueManager 注入写队列管理器
func (d *Dao) ExecuteWrite(ctx context.Context, uid int64, r daoDBCustomKey, fn func(*gorm.DB) error) error {
	dbKey := r.GetKey(uid)
	cfg := d.resolveConfig(dbKey)

	// Determine whether to enable write queue
	// 判断是否启用写队列
	enableQueue := (cfg.EnableWriteQueue == nil || *cfg.EnableWriteQueue)

	if enableQueue {
		if d.writeQueueMgr == nil {
			return fmt.Errorf("writeQueueMgr is nil, must inject via WithWriteQueueManager")
		}
		return d.writeQueueMgr.Execute(ctx, dbKey, func() error {
			db := d.ResolveDB(dbKey)
			if db == nil {
				return fmt.Errorf("database connection is nil (uid=%d, dbKey=%s)", uid, dbKey)
			}
			return fn(db.WithContext(ctx))
		})
	}

	// When not using write queue and concurrency control is configured, check concurrency limits
	// 不使用写队列且配置了并发控制时，检查并发限制
	if !enableQueue && cfg.MaxWriteConcurrency > 0 {
		// Determine the group identifier for configuration (used for sharing the same limiter)
		// 确定配置的分组标识（用于共享同一个限制器）
		// Simple handling here: main DB and user DB have independent concurrency limit pools
		// 这里简单处理：主库和用户库各自拥有独立的并发限制池
		groupKey := "main"
		if dbKey != "" {
			groupKey = "user"
		}

		actual, _ := d.poolSemaphores.LoadOrStore(groupKey, semaphore.NewWeighted(int64(cfg.MaxWriteConcurrency)))
		sem := actual.(*semaphore.Weighted)

		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		defer sem.Release(1)
	}

	// Execute write operation
	// 执行写操作
	db := d.ResolveDB(dbKey)
	if db == nil {
		return fmt.Errorf("database connection is nil (uid=%d)", uid)
	}
	return fn(db.WithContext(ctx))
}

// ExecuteRead executes read operation (executed directly, not through write queue)
// ExecuteRead 执行读操作（直接执行，不经过写队列）
// Read operations do not need serialization and can be executed concurrently
// 读操作不需要串行化，可以并发执行
// ctx: Context for timeout and cancellation control // ctx: 上下文，用于超时和取消控制
// uid: User ID, used to get user database connection // uid: 用户 ID，用于获取用户数据库连接
// fn: Read operation function, receiving user database connection // fn: 读操作函数，接收用户数据库连接
// Return value: Error of the read operation // 返回值: 读操作的错误
func (d *Dao) ExecuteRead(ctx context.Context, uid int64, r daoDBCustomKey, fn func(*gorm.DB) error) error {
	db := d.ResolveDB(r.GetKey(uid))
	if db == nil {
		return fmt.Errorf("database connection is nil (uid=%d)", uid)
	}
	return fn(db.WithContext(ctx))
}

// ExecuteWriteWithRetry executes write operation (serialized through write queue, with retries)
// ExecuteWriteWithRetry 执行写操作（通过写队列串行化，带重试）
// Combine write queue and retry logic to handle SQLite concurrent write issues
// 结合写队列和重试逻辑，用于处理 SQLite 并发写入问题
// ctx: Context for timeout and cancellation control // ctx: 上下文，用于超时和取消控制
// uid: User ID, used to determine write queue // uid: 用户 ID，用于确定写队列
// fn: Write operation function, receiving user database connection // fn: 写操作函数，接收用户数据库连接
// Return value: Error of the write operation // 返回值: 写操作的错误
func (d *Dao) ExecuteWriteWithRetry(ctx context.Context, uid int64, r daoDBCustomKey, fn func(*gorm.DB) error) error {
	return d.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		return d.WithRetry(func() error {
			return fn(db)
		})
	})
}

// getModelDBKey gets the corresponding database connection Key based on model name
// getModelDBKey 根据模型名称获取对应的数据库连接 Key
func (d *Dao) getModelDBKey(uid int64, modelKey string) string {
	if uid <= 0 {
		return "" // Main database // 主数据库
	}

	for _, cfg := range modelConfigs {
		if cfg.Name == modelKey {
			if cfg.IsMainDB {
				return ""
			}
			if cfg.RepoFactory != nil {
				return cfg.RepoFactory(d).GetKey(uid)
			}
		}
	}

	return ""
}

func (d *Dao) AutoMigrate(uid int64, modelKey string) error {
	// 1. If modelKey is empty, it means "full migration", route migration separately by model
	// 1. 如果 modelKey 为空，说明是“全量迁移”，按模型分别路由迁移
	if modelKey == "" {
		for _, cfg := range modelConfigs {
			if err := d.AutoMigrate(uid, cfg.Name); err != nil {
				return err
			}
		}
		return nil
	}

	dbKey := d.getModelDBKey(uid, modelKey)
	cfg := d.resolveConfig(dbKey)

	// 2. Verify the AutoMigrate flag in the configuration
	// 2. 校验配置中的 AutoMigrate 标志
	if !cfg.AutoMigrate {
		return nil
	}

	b := d.ResolveDB(dbKey)

	if b == nil {
		return fmt.Errorf("database connection is nil for model %s (uid=%d, dbKey=%s)", modelKey, uid, dbKey)
	}
	return model.AutoMigrate(b, modelKey)
}

// user gets the user query object (internal method)
// user 获取用户查询对象（内部方法）
func (d *Dao) user() *query.Query {
	return d.QueryWithOnceInit(func(g *gorm.DB) {
		model.AutoMigrate(g, "User")
	}, "user#user")
}

// GetAllUserUIDs retrieves UIDs of all users
// GetAllUserUIDs 获取所有用户的UID
// Return value description:
// 返回值说明:
//   - []int64: User UID list // 用户UID列表
//   - error: Error on failure // 出错时返回错误
func (d *Dao) GetAllUserUIDs() ([]int64, error) {
	var uids []int64
	u := d.user().User
	err := u.WithContext(d.ctx).Select(u.UID).Where(u.IsDeleted.Eq(0)).Scan(&uids)
	if err != nil {
		return nil, err
	}
	return uids, nil
}

// ensurePostgresSchema ensures the specified Schema exists in PostgreSQL
// ensurePostgresSchema 确保 PostgreSQL 中指定的 Schema 存在
func (d *Dao) ensurePostgresSchema(schemaName string) error {
	if d.userConfig == nil || d.userConfig.Type != "postgres" {
		return nil
	}

	// Construct basic connection DSN without schema
	// 构造不带 schema 的基础连接 DSN
	cfg := *d.userConfig
	cfg.Schema = ""

	// Use base connection to create a new Schema
	// 使用基础连接来创建新的 Schema
	db, err := NewEngine(cfg, d.Logger())
	if err != nil {
		return fmt.Errorf("failed to open root postgres connection: %w", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	// Execute create Schema statement
	// 执行创建 Schema 语句
	err = db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)).Error
	if err != nil {
		return fmt.Errorf("failed to create schema %s: %w", schemaName, err)
	}

	return nil
}

// extractUserSchema extracts uniform user Schema name (e.g., user_1) from connection Key (e.g., user_vault_1)
// extractUserSchema 从连接 Key (如 user_vault_1) 中提取统一的用户 Schema 名 (如 user_1)
func (d *Dao) extractUserSchema(key string) (string, bool) {
	// Find the last underscore, try to extract UID
	// 查找最后一个下划线，尝试提取 UID
	lastUnder := strings.LastIndex(key, "_")
	if lastUnder == -1 {
		return "", false
	}
	uidStr := key[lastUnder+1:]
	// If the last part is pure digits, we consider it the UID and map it to a uniform Schema: user_<uid>
	// 如果最后一部分是纯数字，我们认为它是 UID，并映射到统一的 Schema: user_<uid>
	if _, err := strconv.ParseInt(uidStr, 10, 64); err == nil {
		return "user_" + uidStr, true
	}
	return "", false
}

// ensureMysqlDatabase ensures the specified database exists in MySQL
// ensureMysqlDatabase 确保 MySQL 中指定的数据库存在
func (d *Dao) ensureMysqlDatabase(dbName string) error {
	if d.userConfig == nil || d.userConfig.Type != "mysql" {
		return nil
	}

	// Construct basic connection configuration without database name
	// 构造不带数据库名的基础连接配置
	cfg := *d.userConfig
	cfg.Name = ""

	// Use base connection to connect to MySQL service
	// 使用基础连接连接到 MySQL 服务
	db, err := NewEngine(cfg, d.Logger())
	if err != nil {
		return fmt.Errorf("failed to open root mysql connection: %w", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	// Execute create database statement
	// 执行创建数据库语句
	// Note: MySQL database names cannot contain special characters; user_<uid> is safe
	// 注意：MySQL 库名不能包含特殊字符，user_<uid> 是安全的
	err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci", dbName)).Error
	if err != nil {
		return fmt.Errorf("failed to create database %s: %w", dbName, err)
	}

	return nil
}
