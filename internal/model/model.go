
package model

import (
	"gorm.io/gorm"
)



func AutoMigrate(db *gorm.DB, key string) error {
	switch key {

	case "Note":
		return db.AutoMigrate(Note{})

	case "User":
		return db.AutoMigrate(User{})

	case "Vault":
		return db.AutoMigrate(Vault{})
	}
	return nil
}