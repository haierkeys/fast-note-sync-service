package dao

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/haierkeys/obsidian-better-sync-service/internal/query"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/fileurl"

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

func New(db *gorm.DB, ctx context.Context) *Dao {
	return &Dao{Db: db, ctx: ctx, KeyDb: make(map[string]*gorm.DB)}
}

func (d *Dao) Use(f func(*gorm.DB), key ...string) *query.Query {

	var db *gorm.DB
	if len(key) > 0 {
		db = d.UseDb(key...)
	} else {
		db = d.Db
	}
	f(db)
	return query.Use(db)
}

func (d *Dao) UseDb(k ...string) *gorm.DB {

	key := k[0]
	if db, ok := d.KeyDb[key]; ok {
		return db
	}

	c := global.Config.Database

	if c.Type == "mysql" {
		c.Name = c.Name + "_" + key
	} else if c.Type == "sqlite" {
		c.Path = c.Path + "_" + key
	}

	db := d.Db.Session(&gorm.Session{})

	db.Dialector = d.switchDB(global.Config.Database, key)
	d.KeyDb[key] = db
	return db

}

// switchDB 重新初始化 GORM 连接以切换数据库
func (d *Dao) switchDB(c global.Database, key string) gorm.Dialector {

	if c.Type == "mysql" {
		c.Name = c.Name + "_" + key
	} else if c.Type == "sqlite" {
		c.Path = c.Path + "_" + key
	} else {
		return nil
	}
	return useDia(c)
}

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

func NewDBEngineTest(c global.Database) (*gorm.DB, error) {

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
