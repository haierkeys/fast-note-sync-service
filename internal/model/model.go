package model

import (
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB, key string) error {
	if db == nil {
		return nil
	}
	switch key {

	case "BackupConfig":
		return db.AutoMigrate(BackupConfig{})

	case "BackupHistory":
		return db.AutoMigrate(BackupHistory{})

	case "File":
		return db.AutoMigrate(File{})

	case "Folder":
		return db.AutoMigrate(Folder{})

	case "GitSyncConfig":
		return db.AutoMigrate(GitSyncConfig{})

	case "GitSyncHistory":
		return db.AutoMigrate(GitSyncHistory{})

	case "Note":
		return db.AutoMigrate(Note{})

	case "NoteHistory":
		return db.AutoMigrate(NoteHistory{})

	case "NoteLink":
		return db.AutoMigrate(NoteLink{})

	case "Setting":
		return db.AutoMigrate(Setting{})

	case "Storage":
		return db.AutoMigrate(Storage{})

	case "User":
		return db.AutoMigrate(User{})

	case "UserShare":
		return db.AutoMigrate(UserShare{})

	case "Vault":
		return db.AutoMigrate(Vault{})
	}
	return nil
}
