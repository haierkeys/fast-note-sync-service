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

// DatabaseConfig 数据库配置（用于依赖注入）
// DatabaseConfig is now imported from internal/config

type dbEntry struct {
	db       *gorm.DB
	lastUsed time.Time
}

// Dao 数据访问对象，封装数据库操作
type Dao struct {
	Db       *gorm.DB
	KeyDb    map[string]*dbEntry
	ctx      context.Context
	onceKeys sync.Map
	mu       sync.RWMutex // 保护 KeyDb 的并发访问

	poolSemaphores sync.Map // map[string]*semaphore.Weighted 针对不同配置的并发控制

	// 注入的依赖
	config        *config.DatabaseConfig
	userConfig    *config.DatabaseConfig
	logger        *zap.Logger
	writeQueueMgr *writequeue.Manager
}

// DaoOption 用于配置 Dao 的选项函数
type DaoOption func(*Dao)

// WithConfig 设置数据库配置
func WithConfig(cfg *config.DatabaseConfig) DaoOption {
	return func(d *Dao) {
		d.config = cfg
	}
}

// WithUserDatabaseConfig 设置用户数据库配置
func WithUserDatabaseConfig(cfg *config.DatabaseConfig) DaoOption {
	return func(d *Dao) {
		d.userConfig = cfg
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

// ModelConfig 描述一个模型的数据库路由信息
type ModelConfig struct {
	Name        string
	RepoFactory func(d *Dao) daoDBCustomKey
	IsMainDB    bool
}

var modelConfigs []ModelConfig

// RegisterModel 供各 Repository 文件在 init() 中调用
func RegisterModel(cfg ModelConfig) {
	modelConfigs = append(modelConfigs, cfg)
}

// New 创建 Dao 实例（支持依赖注入）
// db: 主数据库连接
// ctx: 上下文
// opts: 可选配置项
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

// Logger 获取日志器
func (d *Dao) Logger() *zap.Logger {
	if d.logger != nil {
		return d.logger
	}
	return zap.NewNop()
}

// Config 获取数据库配置
func (d *Dao) Config() *config.DatabaseConfig {
	return d.config
}

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
		panic(fmt.Sprintf("数据库实例为 nil (key=%s, onceKey=%s),请检查数据库配置和连接", keyName, onceKey))
	}

	// 构造库级唯一的初始化 Key
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

// resolveConfig 获取数据库配置
// key: 数据库标识，如果非空则尝试获取用户数据库配置
func (d *Dao) resolveConfig(key string) config.DatabaseConfig {
	var cfg config.DatabaseConfig
	// 如果是针对特定 Key (通常为用户库) 且配置了独立的 UserDatabase 类型
	if key != "" && d.userConfig != nil && d.userConfig.Type != "" {
		cfg = *d.userConfig
	} else if d.config != nil {
		// 否则继承主数据库配置 (Fallback 模式)
		cfg = *d.config
	}

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
	// 使用读锁检查是否已存在
	d.mu.RLock()
	if entry, ok := d.KeyDb[key]; ok {
		entry.lastUsed = time.Now()
		d.mu.RUnlock()
		return entry.db
	}
	d.mu.RUnlock()

	// 获取配置
	c := d.resolveConfig(key)

	if (c.Type == "postgres") && key != "" {
		// PostgreSQL: 统一映射到 user_<uid> Schema，忽略具体的 Repo 前缀
		schemaName, ok := d.extractUserSchema(key)
		if !ok {
			schemaName = key // 回退逻辑，如果无法解析则使用原始 key
		}

		if err := d.ensurePostgresSchema(schemaName); err != nil {
			d.Logger().Error("ensurePostgresSchema failed", zap.String("schema", schemaName), zap.Error(err))
			return nil
		}
		c.Schema = schemaName
		c.TablePrefix = "" // PostgreSQL 下清空前缀，改用 Schema
	} else if (c.Type == "mysql") && key != "" {
		// MySQL: 统一映射到 user_<uid> 数据库，实现租户级库隔离
		dbName, ok := d.extractUserSchema(key)
		if !ok {
			dbName = key // 回退逻辑，如果无法解析则使用原始 key
		}

		if err := d.ensureMysqlDatabase(dbName); err != nil {
			d.Logger().Error("ensureMysqlDatabase failed", zap.String("db", dbName), zap.Error(err))
			return nil
		}
		c.Name = dbName
		c.TablePrefix = "" // MySQL 下清空前缀，改用库路由
	} else if c.Type == "sqlite" && key != "" {
		// SQLite: 维持多文件隔离模式 (使用完整的 key 作为文件名后缀)
		ext := filepath.Ext(c.Path)
		c.Path = c.Path[:len(c.Path)-len(ext)] + "_" + key + ext
	}

	dbNew, err := NewEngine(c, d.Logger())
	if err != nil {
		d.Logger().Error("GetOrCreateDB failed", zap.String("key", key), zap.Error(err))
		return nil
	}

	// 使用写锁存储
	d.mu.Lock()
	defer d.mu.Unlock()
	// 双重检查
	if existingEntry, ok := d.KeyDb[key]; ok {
		// 关闭新创建的连接
		if sqlDB, err := dbNew.DB(); err == nil {
			sqlDB.Close()
		}
		existingEntry.lastUsed = time.Now()
		return existingEntry.db
	}

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
	dbKey := r.GetKey(uid)
	cfg := d.resolveConfig(dbKey)

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

	// 不使用写队列且配置了并发控制时，检查并发限制
	if !enableQueue && cfg.MaxWriteConcurrency > 0 {
		// 确定配置的分组标识（用于共享同一个限制器）
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

	// 执行写操作
	db := d.ResolveDB(dbKey)
	if db == nil {
		return fmt.Errorf("database connection is nil (uid=%d)", uid)
	}
	return fn(db.WithContext(ctx))
}

// ExecuteRead 执行读操作（直接执行，不经过写队列）
// 读操作不需要串行化，可以并发执行
// ctx: 上下文，用于超时和取消控制
// uid: 用户 ID，用于获取用户数据库连接
// fn: 读操作函数，接收用户数据库连接
// 返回值: 读操作的错误
func (d *Dao) ExecuteRead(ctx context.Context, uid int64, r daoDBCustomKey, fn func(*gorm.DB) error) error {
	db := d.ResolveDB(r.GetKey(uid))
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

// getModelDBKey 根据模型名称获取对应的数据库连接 Key
func (d *Dao) getModelDBKey(uid int64, modelKey string) string {
	if uid <= 0 {
		return "" // 主数据库
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

// user 获取用户查询对象（内部方法）
func (d *Dao) user() *query.Query {
	return d.QueryWithOnceInit(func(g *gorm.DB) {
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

// ensurePostgresSchema 确保 PostgreSQL 中指定的 Schema 存在
func (d *Dao) ensurePostgresSchema(schemaName string) error {
	if d.userConfig == nil || d.userConfig.Type != "postgres" {
		return nil
	}

	// 构造不带 schema 的基础连接 DSN
	cfg := *d.userConfig
	cfg.Schema = ""

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

	// 执行创建 Schema 语句
	err = db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)).Error
	if err != nil {
		return fmt.Errorf("failed to create schema %s: %w", schemaName, err)
	}

	return nil
}

// extractUserSchema 从连接 Key (如 user_vault_1) 中提取统一的用户 Schema 名 (如 user_1)
func (d *Dao) extractUserSchema(key string) (string, bool) {
	// 查找最后一个下划线，尝试提取 UID
	lastUnder := strings.LastIndex(key, "_")
	if lastUnder == -1 {
		return "", false
	}
	uidStr := key[lastUnder+1:]
	// 如果最后一部分是纯数字，我们认为它是 UID，并映射到统一的 Schema: user_<uid>
	if _, err := strconv.ParseInt(uidStr, 10, 64); err == nil {
		return "user_" + uidStr, true
	}
	return "", false
}

// ensureMysqlDatabase 确保 MySQL 中指定的数据库存在
func (d *Dao) ensureMysqlDatabase(dbName string) error {
	if d.userConfig == nil || d.userConfig.Type != "mysql" {
		return nil
	}

	// 构造不带数据库名的基础连接配置
	cfg := *d.userConfig
	cfg.Name = ""

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

	// 执行创建数据库语句
	// 注意：MySQL 库名不能包含特殊字符，user_<uid> 是安全的
	err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci", dbName)).Error
	if err != nil {
		return fmt.Errorf("failed to create database %s: %w", dbName, err)
	}

	return nil
}
