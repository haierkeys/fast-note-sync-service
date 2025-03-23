package dao

import (
	"context"
	"fmt"
	"os"
	"reflect"
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
	Q     *query.Query

	User query.Query
}

func (d *Dao) Use(key ...string) *query.Query {
	if len(key) > 0 {
		d.Q = query.Use(d.KeyDb[key[0]])
	} else {
		d.Q = query.Use(d.Db)
	}
	return d.Q
}

func (d *Dao) Use2(modelName string) query.IUserDo {
	Q := d.Use()
	qValue := reflect.ValueOf(Q)
	modelValue := qValue.FieldByName(modelName)
	if !modelValue.IsValid() {
		fmt.Errorf("model %s not found in query.Q", modelName)
		return nil
	}
	// 找到 WithContext 方法
	method := modelValue.MethodByName("WithContext")
	if !method.IsValid() {
		fmt.Errorf("model %s does not have a WithContext method", modelName)
		return nil
	}
	// 调用 WithContext 方法
	results := method.Call([]reflect.Value{reflect.ValueOf(d.ctx)})
	if len(results) == 0 {
		fmt.Errorf("WithContext call returned no results for model %s", modelName)
		return nil
	}
	// 类型断言将结果转换为 query.IUserDo
	userDo, ok := results[0].Interface().(query.IUserDo)
	if !ok {
		fmt.Errorf("result cannot be converted to query.IUserDo")
		return nil
	}
	return userDo

}

func (d *Dao) DB() *gorm.DB {
	return d.Db
}

func (d *Dao) UseKey(key string) *gorm.DB {
	if db, ok := d.KeyDb[key]; ok {
		return db
	}

	db := d.Db.Session(&gorm.Session{})

	db.Dialector = changeDb(global.Config.Database, key)
	d.KeyDb[key] = db
	return db

}

func New(db *gorm.DB) *Dao {

	dao := &Dao{Db: db}
	dao.User = query.Use(dao.Db).User
	x, _ := u.WithContext(d.ctx).Where(u.UID.Eq(10)).First()

	query.SetDefault(db)
	return &Dao{Db: db, Q: query.Q}
}

// switchDB 重新初始化 GORM 连接以切换数据库
func changeDb(c global.Database, key string) gorm.Dialector {

	if c.Type == "mysql" {
		c.Name = c.Name + "_" + key
	} else if c.Type == "sqlite" {
		c.Path = c.Path + "_" + key
	} else {
		return nil
	}
	return userDialector(c)
}

func NewDBEngine(c global.Database) (*gorm.DB, error) {

	var db *gorm.DB
	var err error

	db, err = gorm.Open(userDialector(c), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
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

func userDialector(c global.Database) gorm.Dialector {
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

		if !fileurl.IsExist(c.Path) {
			fileurl.CreatePath(c.Path, os.ModePerm)
		}
		return sqlite.Open(c.Path)
	}
	return nil

}
