
package model

import (
	"gorm.io/gorm"
)



func AutoMigrate(db *gorm.DB, key string) {
	switch key {

	case "Note":
		db.AutoMigrate(Note{})

	case "User":
		db.AutoMigrate(User{})

	case "Vault":
		db.AutoMigrate(Vault{})
	}
}