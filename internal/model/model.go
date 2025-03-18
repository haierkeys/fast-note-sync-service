package model

import (
	"fmt"
	"os"
	"time"

	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/fileurl"

	"github.com/haierkeys/gormTracing"
	"gorm.io/driver/mysql"

	//"gorm.io/driver/sqlite"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Predicate string

var (
	// Eq =
	Eq = Predicate("=")
	// Neq <>
	Neq = Predicate("<>")
	// Gt >
	Gt = Predicate(">")
	// Egt >=
	Egt = Predicate(">=")
	// Lt <
	Lt = Predicate("<")
	// Elt <=
	Elt  = Predicate("<=")
	Like = Predicate("LIKE")
)

func NewDBEngine(c global.Database) (*gorm.DB, error) {

	var db *gorm.DB
	var err error
	var isEnable bool

	if c.Type == "mysql" {
		db, err = gorm.Open(
			mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&parseTime=%t&loc=Local",
				c.UserName,
				c.Password,
				c.Host,
				c.Name,
				c.Charset,
				c.ParseTime,
			)),
			&gorm.Config{
				Logger: logger.Default.LogMode(logger.Silent),
				NamingStrategy: schema.NamingStrategy{
					TablePrefix:   c.TablePrefix, // 表名前缀，`User` 的表名应该是 `t_users`
					SingularTable: true,          // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `t_user`
				},
			},
		)
		if err != nil {
			return nil, err
		}
		isEnable = true
	} else if c.Type == "sqlite" {

		if !fileurl.IsExist(c.Path) {
			fileurl.CreatePath(c.Path, os.ModePerm)
		}

		db, err = gorm.Open(sqlite.Open(c.Path), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   c.TablePrefix, // 表名前缀，`User` 的表名应该是 `t_users`
				SingularTable: true,          // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `t_user`
			},
		})
		if err != nil {
			return nil, err
		}
		isEnable = true
	}

	if isEnable {

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
		sqlDB.SetConnMaxLifetime(time.Hour)

		_ = db.Use(&gormTracing.OpentracingPlugin{})

		return db, nil
	} else {
		return nil, err
	}
}
