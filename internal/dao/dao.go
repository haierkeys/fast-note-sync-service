package dao

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"

	"github.com/glebarez/sqlite"
	"github.com/haierkeys/gormTracing"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Dao struct {
	Db    *gorm.DB
	KeyDb map[string]*gorm.DB
	ctx   context.Context
}

// New 创建 Dao 实例
// 函数名: New
// 函数使用说明: 创建并初始化一个新的 Dao 实例,用于数据库操作。
// 参数说明:
//   - db *gorm.DB: GORM 数据库连接实例
//   - ctx context.Context: 上下文对象
//
// 返回值说明:
//   - *Dao: Dao 实例
func New(db *gorm.DB, ctx context.Context) *Dao {
	return &Dao{Db: db, ctx: ctx, KeyDb: make(map[string]*gorm.DB)}
}

// Use 获取查询对象
// 函数名: Use
// 函数使用说明: 获取指定数据库的查询对象,并执行初始化函数。
// 参数说明:
//   - f func(*gorm.DB): 数据库初始化函数
//   - key ...string: 数据库键名(可选)
//
// 返回值说明:
//   - *query.Query: 查询对象
func (d *Dao) Use(f func(*gorm.DB), key ...string) *query.Query {
	db := d.UseKey(key...)
	f(db)
	return query.Use(db)
}

// UseKey 获取数据库连接
// 函数名: UseKey
// 函数使用说明: 根据键名获取对应的数据库连接,如果未指定键名则返回默认数据库连接。
// 参数说明:
//   - key ...string: 数据库键名(可选)
//
// 返回值说明:
//   - *gorm.DB: 数据库连接实例
func (d *Dao) UseKey(key ...string) *gorm.DB {
	var db *gorm.DB
	if len(key) > 0 {
		db = d.UseDb(key...)
	} else {
		db = d.Db
	}
	return db
}

// UseDb 获取或创建指定键名的数据库连接
// 函数名: UseDb
// 函数使用说明: 根据键名获取数据库连接,如果不存在则创建新的数据库连接并缓存。
// 参数说明:
//   - k ...string: 数据库键名
//
// 返回值说明:
//   - *gorm.DB: 数据库连接实例
func (d *Dao) UseDb(k ...string) *gorm.DB {

	key := k[0]
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

	dbNew, _ := NewDBEngine(c)

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
		return sqlite.Open(c.Path)
	}
	return nil

}
