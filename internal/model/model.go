
package model

import (
	"sync"

	"gorm.io/gorm"
)

var once sync.Once

func AutoMigrate(db *gorm.DB, key string) {
	switch key {

	case "Note":
		once.Do(func() {
			db.AutoMigrate(Note{})
		})

	case "User":
		once.Do(func() {
			db.AutoMigrate(User{})
		})
	}
}