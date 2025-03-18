package db_driver

import (
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var _ Repo = (*sqliteRepo)(nil)

type sqliteRepo struct {
	DbConn *gorm.DB
}

func (d *sqliteRepo) i() {}

func (d *sqliteRepo) GetDb() *gorm.DB {
	return d.DbConn
}

func (d *sqliteRepo) DbClose() error {
	sqlDB, err := d.DbConn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *sqliteRepo) DbNames() string {
	return "PRAGMA database_list"
}

func (d *sqliteRepo) DbType() string {
	return "sqlite"
}

func NewSqlite(Dsn string) (Repo, error) {

	db, err := gorm.Open(sqlite.Open(Dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `t_user`
		},
	})

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("[db_driver connection failed] %s", Dsn))
	}

	db.Set("gorm:table_options", "CHARSET=utf8mb4")

	return &sqliteRepo{
		DbConn: db,
	}, nil
}
