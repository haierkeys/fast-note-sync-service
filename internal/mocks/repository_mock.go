package mocks

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// NoteRepositoryMock is a mock implementation of domain.NoteRepository
type NoteRepositoryMock struct {
	mock.Mock
}

func (m *NoteRepositoryMock) GetByID(ctx context.Context, id, uid int64) (*domain.Note, error) {
	args := m.Called(ctx, id, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *NoteRepositoryMock) GetByPathHashIncludeRecycle(ctx context.Context, pathHash string, vaultID, uid int64, isRecycle bool) (*domain.Note, error) {
	args := m.Called(ctx, pathHash, vaultID, uid, isRecycle)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *NoteRepositoryMock) GetAllByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*domain.Note, error) {
	args := m.Called(ctx, pathHash, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *NoteRepositoryMock) GetByPath(ctx context.Context, path string, vaultID, uid int64) (*domain.Note, error) {
	args := m.Called(ctx, path, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *NoteRepositoryMock) Create(ctx context.Context, note *domain.Note, uid int64) (*domain.Note, error) {
	args := m.Called(ctx, note, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *NoteRepositoryMock) Update(ctx context.Context, note *domain.Note, uid int64) (*domain.Note, error) {
	args := m.Called(ctx, note, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *NoteRepositoryMock) UpdateDelete(ctx context.Context, note *domain.Note, uid int64) error {
	args := m.Called(ctx, note, uid)
	return args.Error(0)
}

func (m *NoteRepositoryMock) UpdateMtime(ctx context.Context, mtime int64, id, uid int64) error {
	args := m.Called(ctx, mtime, id, uid)
	return args.Error(0)
}

func (m *NoteRepositoryMock) UpdateActionMtime(ctx context.Context, action domain.NoteAction, mtime int64, id, uid int64) error {
	args := m.Called(ctx, action, mtime, id, uid)
	return args.Error(0)
}

func (m *NoteRepositoryMock) UpdateSnapshot(ctx context.Context, snapshot, snapshotHash string, version, id, uid int64) error {
	args := m.Called(ctx, snapshot, snapshotHash, version, id, uid)
	return args.Error(0)
}

func (m *NoteRepositoryMock) Delete(ctx context.Context, id, uid int64) error {
	args := m.Called(ctx, id, uid)
	return args.Error(0)
}

func (m *NoteRepositoryMock) DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error {
	args := m.Called(ctx, timestamp, uid)
	return args.Error(0)
}

func (m *NoteRepositoryMock) DeletePhysicalByTimeAll(ctx context.Context, timestamp int64) error {
	args := m.Called(ctx, timestamp)
	return args.Error(0)
}

func (m *NoteRepositoryMock) List(ctx context.Context, vaultID int64, page, pageSize int, uid int64, keyword string, isRecycle bool, searchMode string, searchContent bool, sortBy string, sortOrder string, paths []string) ([]*domain.Note, error) {
	args := m.Called(ctx, vaultID, page, pageSize, uid, keyword, isRecycle, searchMode, searchContent, sortBy, sortOrder, paths)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Note), args.Error(1)
}

func (m *NoteRepositoryMock) ListCount(ctx context.Context, vaultID, uid int64, keyword string, isRecycle bool, searchMode string, searchContent bool, paths []string) (int64, error) {
	args := m.Called(ctx, vaultID, uid, keyword, isRecycle, searchMode, searchContent, paths)
	return args.Get(0).(int64), args.Error(1)
}

func (m *NoteRepositoryMock) ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.Note, error) {
	args := m.Called(ctx, timestamp, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Note), args.Error(1)
}

func (m *NoteRepositoryMock) ListByUpdatedTimestampPage(ctx context.Context, timestamp, vaultID, uid int64, offset, limit int) ([]*domain.Note, error) {
	args := m.Called(ctx, timestamp, vaultID, uid, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Note), args.Error(1)
}

func (m *NoteRepositoryMock) ListContentUnchanged(ctx context.Context, uid int64) ([]*domain.Note, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Note), args.Error(1)
}

func (m *NoteRepositoryMock) CountSizeSum(ctx context.Context, vaultID, uid int64) (*domain.CountSizeResult, error) {
	args := m.Called(ctx, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CountSizeResult), args.Error(1)
}

func (m *NoteRepositoryMock) ListByFID(ctx context.Context, fid, vaultID, uid int64, page, pageSize int, sortBy, sortOrder string) ([]*domain.Note, error) {
	args := m.Called(ctx, fid, vaultID, uid, page, pageSize, sortBy, sortOrder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Note), args.Error(1)
}

func (m *NoteRepositoryMock) ListByFIDCount(ctx context.Context, fid, vaultID, uid int64) (int64, error) {
	args := m.Called(ctx, fid, vaultID, uid)
	return args.Get(0).(int64), args.Error(1)
}

func (m *NoteRepositoryMock) ListByFIDs(ctx context.Context, fids []int64, vaultID, uid int64, page, pageSize int, sortBy, sortOrder string) ([]*domain.Note, error) {
	args := m.Called(ctx, fids, vaultID, uid, page, pageSize, sortBy, sortOrder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Note), args.Error(1)
}

func (m *NoteRepositoryMock) ListByFIDsCount(ctx context.Context, fids []int64, vaultID, uid int64) (int64, error) {
	args := m.Called(ctx, fids, vaultID, uid)
	return args.Get(0).(int64), args.Error(1)
}

func (m *NoteRepositoryMock) ListByIDs(ctx context.Context, ids []int64, uid int64) ([]*domain.Note, error) {
	args := m.Called(ctx, ids, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Note), args.Error(1)
}

func (m *NoteRepositoryMock) ListByPathPrefix(ctx context.Context, pathPrefix string, vaultID, uid int64) ([]*domain.Note, error) {
	args := m.Called(ctx, pathPrefix, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Note), args.Error(1)
}

func (m *NoteRepositoryMock) RecycleClear(ctx context.Context, path, pathHash string, vaultID, uid int64) error {
	args := m.Called(ctx, path, pathHash, vaultID, uid)
	return args.Error(0)
}

func (m *NoteRepositoryMock) GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*domain.Note, error) {
	args := m.Called(ctx, pathHash, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

// VaultRepositoryMock is a mock implementation of domain.VaultRepository
type VaultRepositoryMock struct {
	mock.Mock
}

func (m *VaultRepositoryMock) GetByID(ctx context.Context, id, uid int64) (*domain.Vault, error) {
	args := m.Called(ctx, id, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Vault), args.Error(1)
}

func (m *VaultRepositoryMock) GetByName(ctx context.Context, name string, uid int64) (*domain.Vault, error) {
	args := m.Called(ctx, name, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Vault), args.Error(1)
}

func (m *VaultRepositoryMock) Create(ctx context.Context, vault *domain.Vault, uid int64) (*domain.Vault, error) {
	args := m.Called(ctx, vault, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Vault), args.Error(1)
}

func (m *VaultRepositoryMock) Update(ctx context.Context, vault *domain.Vault, uid int64) error {
	args := m.Called(ctx, vault, uid)
	return args.Error(0)
}

func (m *VaultRepositoryMock) UpdateNoteCountSize(ctx context.Context, noteSize, noteCount, vaultID, uid int64) error {
	args := m.Called(ctx, noteSize, noteCount, vaultID, uid)
	return args.Error(0)
}

func (m *VaultRepositoryMock) UpdateFileCountSize(ctx context.Context, fileSize, fileCount, vaultID, uid int64) error {
	args := m.Called(ctx, fileSize, fileCount, vaultID, uid)
	return args.Error(0)
}

func (m *VaultRepositoryMock) List(ctx context.Context, uid int64) ([]*domain.Vault, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Vault), args.Error(1)
}

func (m *VaultRepositoryMock) Delete(ctx context.Context, id, uid int64) error {
	args := m.Called(ctx, id, uid)
	return args.Error(0)
}

// BackupRepositoryMock is a mock implementation of domain.BackupRepository
type BackupRepositoryMock struct {
	mock.Mock
}

func (m *BackupRepositoryMock) ListConfigs(ctx context.Context, uid int64) ([]*domain.BackupConfig, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.BackupConfig), args.Error(1)
}

func (m *BackupRepositoryMock) GetByID(ctx context.Context, id, uid int64) (*domain.BackupConfig, error) {
	args := m.Called(ctx, id, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BackupConfig), args.Error(1)
}

func (m *BackupRepositoryMock) DeleteConfig(ctx context.Context, id, uid int64) error {
	args := m.Called(ctx, id, uid)
	return args.Error(0)
}

func (m *BackupRepositoryMock) SaveConfig(ctx context.Context, config *domain.BackupConfig, uid int64) (*domain.BackupConfig, error) {
	args := m.Called(ctx, config, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BackupConfig), args.Error(1)
}

func (m *BackupRepositoryMock) ListEnabledConfigs(ctx context.Context) ([]*domain.BackupConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.BackupConfig), args.Error(1)
}

func (m *BackupRepositoryMock) UpdateNextRunTime(ctx context.Context, id, uid int64, nextRun time.Time) error {
	args := m.Called(ctx, id, uid, nextRun)
	return args.Error(0)
}

func (m *BackupRepositoryMock) CreateHistory(ctx context.Context, history *domain.BackupHistory, uid int64) (*domain.BackupHistory, error) {
	args := m.Called(ctx, history, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BackupHistory), args.Error(1)
}

func (m *BackupRepositoryMock) ListHistory(ctx context.Context, uid int64, configID int64, page, pageSize int) ([]*domain.BackupHistory, int64, error) {
	args := m.Called(ctx, uid, configID, page, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.BackupHistory), args.Get(1).(int64), args.Error(2)
}

func (m *BackupRepositoryMock) ListOldHistory(ctx context.Context, uid int64, configID int64, cutoffTime time.Time) ([]*domain.BackupHistory, error) {
	args := m.Called(ctx, uid, configID, cutoffTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.BackupHistory), args.Error(1)
}

func (m *BackupRepositoryMock) DeleteOldHistory(ctx context.Context, uid int64, configID int64, cutoffTime time.Time) error {
	args := m.Called(ctx, uid, configID, cutoffTime)
	return args.Error(0)
}

// FolderRepositoryMock is a mock implementation of domain.FolderRepository
type FolderRepositoryMock struct {
	mock.Mock
}

func (m *FolderRepositoryMock) GetByID(ctx context.Context, id, uid int64) (*domain.Folder, error) {
	args := m.Called(ctx, id, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Folder), args.Error(1)
}

func (m *FolderRepositoryMock) GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*domain.Folder, error) {
	args := m.Called(ctx, pathHash, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Folder), args.Error(1)
}

func (m *FolderRepositoryMock) GetAllByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) ([]*domain.Folder, error) {
	args := m.Called(ctx, pathHash, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Folder), args.Error(1)
}

func (m *FolderRepositoryMock) GetByFID(ctx context.Context, fid int64, vaultID, uid int64) ([]*domain.Folder, error) {
	args := m.Called(ctx, fid, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Folder), args.Error(1)
}

func (m *FolderRepositoryMock) Create(ctx context.Context, folder *domain.Folder, uid int64) (*domain.Folder, error) {
	args := m.Called(ctx, folder, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Folder), args.Error(1)
}

func (m *FolderRepositoryMock) Update(ctx context.Context, folder *domain.Folder, uid int64) (*domain.Folder, error) {
	args := m.Called(ctx, folder, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Folder), args.Error(1)
}

func (m *FolderRepositoryMock) Delete(ctx context.Context, id, uid int64) error {
	args := m.Called(ctx, id, uid)
	return args.Error(0)
}

func (m *FolderRepositoryMock) ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.Folder, error) {
	args := m.Called(ctx, timestamp, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Folder), args.Error(1)
}

func (m *FolderRepositoryMock) List(ctx context.Context, vaultID int64, uid int64) ([]*domain.Folder, error) {
	args := m.Called(ctx, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Folder), args.Error(1)
}

func (m *FolderRepositoryMock) ListAll(ctx context.Context, uid int64) ([]*domain.Folder, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Folder), args.Error(1)
}

// FileRepositoryMock is a mock implementation of domain.FileRepository
type FileRepositoryMock struct {
	mock.Mock
}

func (m *FileRepositoryMock) GetByID(ctx context.Context, id, uid int64) (*domain.File, error) {
	args := m.Called(ctx, id, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.File), args.Error(1)
}

func (m *FileRepositoryMock) GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*domain.File, error) {
	args := m.Called(ctx, pathHash, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.File), args.Error(1)
}

func (m *FileRepositoryMock) GetByPath(ctx context.Context, path string, vaultID, uid int64) (*domain.File, error) {
	args := m.Called(ctx, path, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.File), args.Error(1)
}

func (m *FileRepositoryMock) GetByPathLike(ctx context.Context, path string, vaultID, uid int64) (*domain.File, error) {
	args := m.Called(ctx, path, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.File), args.Error(1)
}

func (m *FileRepositoryMock) Create(ctx context.Context, file *domain.File, uid int64) (*domain.File, error) {
	args := m.Called(ctx, file, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.File), args.Error(1)
}

func (m *FileRepositoryMock) Update(ctx context.Context, file *domain.File, uid int64) (*domain.File, error) {
	args := m.Called(ctx, file, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.File), args.Error(1)
}

func (m *FileRepositoryMock) UpdateMtime(ctx context.Context, mtime int64, id, uid int64) error {
	args := m.Called(ctx, mtime, id, uid)
	return args.Error(0)
}

func (m *FileRepositoryMock) UpdateActionMtime(ctx context.Context, action domain.FileAction, mtime int64, id, uid int64) error {
	args := m.Called(ctx, action, mtime, id, uid)
	return args.Error(0)
}

func (m *FileRepositoryMock) Delete(ctx context.Context, id, uid int64) error {
	args := m.Called(ctx, id, uid)
	return args.Error(0)
}

func (m *FileRepositoryMock) DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error {
	args := m.Called(ctx, timestamp, uid)
	return args.Error(0)
}

func (m *FileRepositoryMock) DeletePhysicalByTimeAll(ctx context.Context, timestamp int64) error {
	args := m.Called(ctx, timestamp)
	return args.Error(0)
}

func (m *FileRepositoryMock) List(ctx context.Context, vaultID int64, page, pageSize int, uid int64, keyword string, isRecycle bool, sortBy string, sortOrder string) ([]*domain.File, error) {
	args := m.Called(ctx, vaultID, page, pageSize, uid, keyword, isRecycle, sortBy, sortOrder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.File), args.Error(1)
}

func (m *FileRepositoryMock) ListCount(ctx context.Context, vaultID, uid int64, keyword string, isRecycle bool) (int64, error) {
	args := m.Called(ctx, vaultID, uid, keyword, isRecycle)
	return args.Get(0).(int64), args.Error(1)
}

func (m *FileRepositoryMock) ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.File, error) {
	args := m.Called(ctx, timestamp, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.File), args.Error(1)
}

func (m *FileRepositoryMock) ListByUpdatedTimestampPage(ctx context.Context, timestamp, vaultID, uid int64, offset, limit int) ([]*domain.File, error) {
	args := m.Called(ctx, timestamp, vaultID, uid, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.File), args.Error(1)
}

func (m *FileRepositoryMock) ListByMtime(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.File, error) {
	args := m.Called(ctx, timestamp, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.File), args.Error(1)
}

func (m *FileRepositoryMock) CountSizeSum(ctx context.Context, vaultID, uid int64) (*domain.CountSizeResult, error) {
	args := m.Called(ctx, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CountSizeResult), args.Error(1)
}

func (m *FileRepositoryMock) ListByFID(ctx context.Context, fid, vaultID, uid int64, page, pageSize int, sortBy, sortOrder string) ([]*domain.File, error) {
	args := m.Called(ctx, fid, vaultID, uid, page, pageSize, sortBy, sortOrder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.File), args.Error(1)
}

func (m *FileRepositoryMock) ListByFIDCount(ctx context.Context, fid, vaultID, uid int64) (int64, error) {
	args := m.Called(ctx, fid, vaultID, uid)
	return args.Get(0).(int64), args.Error(1)
}

func (m *FileRepositoryMock) ListByFIDs(ctx context.Context, fids []int64, vaultID, uid int64, page, pageSize int, sortBy, sortOrder string) ([]*domain.File, error) {
	args := m.Called(ctx, fids, vaultID, uid, page, pageSize, sortBy, sortOrder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.File), args.Error(1)
}

func (m *FileRepositoryMock) ListByFIDsCount(ctx context.Context, fids []int64, vaultID, uid int64) (int64, error) {
	args := m.Called(ctx, fids, vaultID, uid)
	return args.Get(0).(int64), args.Error(1)
}

func (m *FileRepositoryMock) ListByIDs(ctx context.Context, ids []int64, uid int64) ([]*domain.File, error) {
	args := m.Called(ctx, ids, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.File), args.Error(1)
}

func (m *FileRepositoryMock) ListByPathPrefix(ctx context.Context, pathPrefix string, vaultID, uid int64) ([]*domain.File, error) {
	args := m.Called(ctx, pathPrefix, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.File), args.Error(1)
}

// UserRepositoryMock is a mock implementation of domain.UserRepository
type UserRepositoryMock struct {
	mock.Mock
}

func (m *UserRepositoryMock) GetByUID(ctx context.Context, uid int64) (*domain.User, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepositoryMock) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepositoryMock) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepositoryMock) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepositoryMock) UpdatePassword(ctx context.Context, password string, uid int64) error {
	args := m.Called(ctx, password, uid)
	return args.Error(0)
}

func (m *UserRepositoryMock) GetAllUIDs(ctx context.Context) ([]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}
