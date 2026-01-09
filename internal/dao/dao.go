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

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"

	"github.com/glebarez/sqlite"
	"github.com/haierkeys/gormTracing"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"go.uber.org/zap"
)

type Dao struct {
	Db       *gorm.DB
	KeyDb    map[string]*gorm.DB
	ctx      context.Context
	onceKeys sync.Map
}

func New(db *gorm.DB, ctx context.Context) *Dao {
	return &Dao{Db: db, ctx: ctx, KeyDb: make(map[string]*gorm.DB)}
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

func (d *Dao) UseDb(key string) *gorm.DB {

	if db, ok := d.KeyDb[key]; ok {
		return db
	}

	c := global.Config.Database

	if c.Type == "mysql" {
		c.Name = c.Name + "_" + key
	} else if c.Type == "sqlite" {
		ext := filepath.Ext(c.Path)
		c.Path = c.Path[:len(c.Path)-len(ext)] + "_" + key + ext
	}

	dbNew, err := NewDBEngine(c)
	if err != nil {
		if global.Logger != nil {
			global.Logger.Error("UseDb failed", zap.String("key", key), zap.Error(err))
		}
		return nil
	}

	d.KeyDb[key] = dbNew
	return dbNew

}

// NewDBEngine 创建数据库引擎
// 函数名: NewDBEngine
// 函数使用说明: 根据配置创建并初始化 GORM 数据库引擎,配置连接池参数和日志模式。
// 参数说明:
//   - c global.Database: 数据库配置
//
// 返回值说明:
//   - *gorm.DB: 数据库连接实例
//   - error: 出错时返回错误
func NewDBEngine(c global.Database) (*gorm.DB, error) {

	var db *gorm.DB
	var err error

	db, err = gorm.Open(useDia(c), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   c.TablePrefix, // 表名前缀，`User` 的表名应该是 `t_users`
			SingularTable: true,          // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `t_user`
		},
	})
	if err != nil {
		return nil, err
	}
	if global.Config.Server.RunMode == "debug" {
		db.Config.Logger = logger.Default.LogMode(logger.Info)
	} else {
		db.Config.Logger = logger.Default.LogMode(logger.Silent)
	}

	// 获取通用数据库对象 sql.DB ，然后使用其提供的功能
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// SetMaxIdleConns 用于设置连接池中空闲连接的最大数量。
	sqlDB.SetMaxIdleConns(c.MaxIdleConns)

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	//sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxOpenConns(c.MaxOpenConns)

	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Minute * 10)

	_ = db.Use(&gormTracing.OpentracingPlugin{})

	return db, nil

}

// useDia 获取数据库方言
// 函数名: useDia
// 函数使用说明: 根据数据库配置返回对应的 GORM 方言(MySQL 或 SQLite)。
// 参数说明:
//   - c global.Database: 数据库配置
//
// 返回值说明:
//   - gorm.Dialector: GORM 数据库方言
func useDia(c global.Database) gorm.Dialector {
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

func (d *Dao) AutoMigrate(uid int64, modelKey string) error {
	var b *gorm.DB
	if uid > 0 {
		key := "user_" + strconv.FormatInt(uid, 10)
		b = d.UseKey(key)
	} else {
		b = d.Db
	}
	if b == nil {
		return fmt.Errorf("database connection is nil (uid=%d)", uid)
	}
	return model.AutoMigrate(b, modelKey)
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
