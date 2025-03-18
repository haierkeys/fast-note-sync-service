package db_driver

import (
	"gorm.io/gorm"
)

type Repo interface {
	i()
	GetDb() *gorm.DB
	DbClose() error
	DbType() string
}
