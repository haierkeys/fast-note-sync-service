package app

import (
	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
)

// Repositories encapsulates all repository instances
type Repositories struct {
	NoteRepo        domain.NoteRepository
	VaultRepo       domain.VaultRepository
	UserRepo        domain.UserRepository
	FileRepo        domain.FileRepository
	SettingRepo     domain.SettingRepository
	NoteHistoryRepo domain.NoteHistoryRepository
	NoteLinkRepo    domain.NoteLinkRepository
	ShareRepo       domain.UserShareRepository
	FolderRepo      domain.FolderRepository
	StorageRepo     domain.StorageRepository
	BackupRepo      domain.BackupRepository
	GitSyncRepo     domain.GitSyncRepository
	SyncLogRepo     domain.SyncLogRepository
	NoteFTSRepo     domain.NoteFTSRepository
	AuthTokenRepo   domain.AuthTokenRepository
	AuthTokenLogRepo domain.AuthTokenLogRepository
}

// initRepositories initializes all repositories
func initRepositories(d *dao.Dao) *Repositories {
	return &Repositories{
		NoteRepo:        dao.NewNoteRepository(d),
		VaultRepo:       dao.NewVaultRepository(d),
		UserRepo:        dao.NewUserRepository(d),
		FileRepo:        dao.NewFileRepository(d),
		SettingRepo:     dao.NewSettingRepository(d),
		NoteHistoryRepo: dao.NewNoteHistoryRepository(d),
		NoteLinkRepo:    dao.NewNoteLinkRepository(d),
		ShareRepo:       dao.NewUserShareRepository(d),
		FolderRepo:      dao.NewFolderRepository(d),
		StorageRepo:     dao.NewStorageRepository(d),
		BackupRepo:      dao.NewBackupRepository(d),
		GitSyncRepo:     dao.NewGitSyncRepository(d),
		SyncLogRepo:     dao.NewSyncLogRepository(d),
		NoteFTSRepo:     dao.NewNoteFTSRepository(d),
		AuthTokenRepo:   dao.NewAuthTokenRepository(d),
		AuthTokenLogRepo: dao.NewAuthTokenLogRepository(d),
	}
}
