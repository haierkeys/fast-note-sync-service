package model

import (
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB, key string) error {
	if db == nil {
		return nil
	}
	switch key {

	case "File":
		return db.AutoMigrate(File{})

	case "Note":
		return db.AutoMigrate(Note{})

	case "NoteHistory":
		return db.AutoMigrate(NoteHistory{})

	case "Setting":
		return db.AutoMigrate(Setting{})

	case "User":
		return db.AutoMigrate(User{})

	case "Vault":
		return db.AutoMigrate(Vault{})
	case "":
		return db.AutoMigrate(File{}, Note{}, NoteHistory{}, Setting{}, User{}, Vault{})
	}
	return nil
}
