package db_driver

import (
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var _ Repo = (*mysqlRepo)(nil)

type mysqlRepo struct {
	DbConn *gorm.DB
}

func (d *mysqlRepo) i() {}

func (d *mysqlRepo) GetDb() *gorm.DB {
	return d.DbConn
}

func (d *mysqlRepo) DbClose() error {
	sqlDB, err := d.DbConn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *mysqlRepo) DbType() string {
	return "mysql"
}

func NewMysql(Dsn string) (Repo, error) {

	db, err := gorm.Open(mysql.Open(Dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		// Logger: logger.Default.LogMode(logger.Info), // 日志配置
	})

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("[db_driver connection failed] %s", Dsn))
	}

	db.Set("gorm:table_options", "CHARSET=utf8mb4")

	return &mysqlRepo{
		DbConn: db,
	}, nil
}
