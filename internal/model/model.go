
package model

import (
	"gorm.io/gorm"
)



func AutoMigrate(db *gorm.DB, key string) error {
	switch key {

	case "File":
		return db.AutoMigrate(File{})

	case "Note":
		return db.AutoMigrate(Note{})

	case "Setting":
		return db.AutoMigrate(Setting{})

	case "User":
		return db.AutoMigrate(User{})

	case "Vault":
		return db.AutoMigrate(Vault{})
	}
	return nil
}