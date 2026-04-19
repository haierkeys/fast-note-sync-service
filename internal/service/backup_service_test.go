// Package service implements the business logic layer.
// Package service 实现业务逻辑层。
package service

import (
	"context"
	"testing"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	domainmocks "github.com/haierkeys/fast-note-sync-service/internal/domain/mocks"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// backupStorageStub is a minimal StorageService stub for backup tests.
// backupStorageStub 是用于备份测试的最小 StorageService stub，避免循环导入。
// Note: service/mocks imports service, so it cannot be imported back in service package tests.
// 注意: service/mocks 包导入了 service，因此在 service 包测试中不能反向导入。
type backupStorageStub struct {
	storages map[int64]*dto.StorageDTO
}

func (s *backupStorageStub) Get(_ context.Context, _ int64, id int64) (*dto.StorageDTO, error) {
	if v, ok := s.storages[id]; ok {
		return v, nil
	}
	return nil, nil
}
func (s *backupStorageStub) CreateOrUpdate(_ context.Context, _ int64, _ int64, _ *dto.StoragePostRequest) (*dto.StorageDTO, error) {
	return nil, nil
}
func (s *backupStorageStub) List(_ context.Context, _ int64) ([]*dto.StorageDTO, error) {
	return nil, nil
}
func (s *backupStorageStub) Delete(_ context.Context, _ int64, _ int64) error { return nil }
func (s *backupStorageStub) GetEnabledTypes() ([]string, error)               { return nil, nil }
func (s *backupStorageStub) Validate(_ context.Context, _ *dto.StoragePostRequest) error {
	return nil
}

// --- BackupService constructor helper ---

// newBackupSvc builds a backupService with mock dependencies for testing.
// newBackupSvc 使用 mock 依赖构建用于测试的 backupService。
func newBackupSvc(
	backupRepo *domainmocks.MockBackupRepository,
	vaultRepo *domainmocks.MockVaultRepository,
	storageSvc *backupStorageStub,
) *backupService {
	return &backupService{
		backupRepo:     backupRepo,
		noteRepo:       new(domainmocks.MockNoteRepository),
		folderRepo:     new(domainmocks.MockFolderRepository),
		fileRepo:       new(domainmocks.MockFileRepository),
		vaultRepo:      vaultRepo,
		storageService: storageSvc,
		logger:         zap.NewNop(),
		syncTimers:     make(map[int64]*time.Timer),
		runningTasks:   make(map[int64]context.CancelFunc),
	}
}

// --- GetConfigs ---

// TestBackupService_GetConfigs_Success verifies that GetConfigs returns mapped DTOs.
// TestBackupService_GetConfigs_Success 验证 GetConfigs 正确返回映射后的 DTO 列表。
func TestBackupService_GetConfigs_Success(t *testing.T) {
	backupRepo := new(domainmocks.MockBackupRepository)
	vaultRepo := new(domainmocks.MockVaultRepository)
	storageSvc := &backupStorageStub{}

	configs := []*domain.BackupConfig{
		{ID: 1, UID: 1, Type: "full", IsEnabled: true},
		{ID: 2, UID: 1, Type: "incremental", IsEnabled: false},
	}
	backupRepo.On("ListConfigs", mock.Anything, int64(1)).Return(configs, nil)

	// GetConfigs internally calls VaultRepo.GetByID for each config's VaultID
	// GetConfigs 内部会对每个 config 的 VaultID 调用 VaultRepo.GetByID
	vaultRepo.On("GetByID", mock.Anything, int64(0), int64(1)).Return(nil, nil).Maybe()

	svc := newBackupSvc(backupRepo, vaultRepo, storageSvc)
	result, err := svc.GetConfigs(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, int64(2), result[1].ID)
	backupRepo.AssertExpectations(t)
}

// TestBackupService_GetConfigs_Empty verifies that GetConfigs returns empty slice when no configs exist.
// TestBackupService_GetConfigs_Empty 验证没有备份配置时返回空列表。
func TestBackupService_GetConfigs_Empty(t *testing.T) {
	backupRepo := new(domainmocks.MockBackupRepository)
	vaultRepo := new(domainmocks.MockVaultRepository)
	storageSvc := &backupStorageStub{}

	backupRepo.On("ListConfigs", mock.Anything, int64(1)).Return([]*domain.BackupConfig{}, nil)

	svc := newBackupSvc(backupRepo, vaultRepo, storageSvc)
	result, err := svc.GetConfigs(context.Background(), 1)

	assert.NoError(t, err)
	assert.Empty(t, result)
	backupRepo.AssertExpectations(t)
}

// --- UpdateConfig ---

// TestBackupService_UpdateConfig_Success verifies config is saved with resolved vault ID.
// TestBackupService_UpdateConfig_Success 验证配置以解析后的 VaultID 保存。
func TestBackupService_UpdateConfig_Success(t *testing.T) {
	backupRepo := new(domainmocks.MockBackupRepository)
	vaultRepo := new(domainmocks.MockVaultRepository)
	storageSvc := &backupStorageStub{storages: map[int64]*dto.StorageDTO{
		200: {ID: 200, IsEnabled: true},
	}}

	vault := &domain.Vault{ID: 100, Name: "myvault"}
	vaultRepo.On("GetByName", mock.Anything, "myvault", int64(1)).Return(vault, nil)

	// storageSvc.Get is handled by backupStorageStub directly.
	// StorageService.Get 由 backupStorageStub 直接处理，无需 mock.On 配置。

	savedConfig := &domain.BackupConfig{ID: 1, VaultID: 100, Type: "full", IsEnabled: true}
	backupRepo.On("SaveConfig", mock.Anything, mock.MatchedBy(func(c *domain.BackupConfig) bool {
		return c.VaultID == 100 && c.Type == "full"
	}), int64(1)).Return(savedConfig, nil)

	// GetByID is called after save from configToDTO; the uid passed is config.UID (0 in this test fixture).
	// configToDTO 调用 GetByID 时传入的 uid 是 config.UID（本测试 fixture 为 0）。
	vaultRepo.On("GetByID", mock.Anything, int64(100), int64(0)).Return(vault, nil)

	svc := newBackupSvc(backupRepo, vaultRepo, storageSvc)
	req := &dto.BackupConfigRequest{
		Vault:      "myvault",
		StorageIds: "[200]",
		Type:       "full",
		IsEnabled:  true,
	}
	result, err := svc.UpdateConfig(context.Background(), 1, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "myvault", result.Vault)
	backupRepo.AssertExpectations(t)
	vaultRepo.AssertExpectations(t)
}

// --- DeleteConfig ---

// TestBackupService_DeleteConfig_Success verifies existing config is deleted.
// TestBackupService_DeleteConfig_Success 验证成功删除已存在的备份配置。
func TestBackupService_DeleteConfig_Success(t *testing.T) {
	backupRepo := new(domainmocks.MockBackupRepository)
	vaultRepo := new(domainmocks.MockVaultRepository)
	storageSvc := &backupStorageStub{}

	backupRepo.On("GetByID", mock.Anything, int64(1), int64(1)).Return(
		&domain.BackupConfig{ID: 1, UID: 1}, nil,
	)
	backupRepo.On("DeleteConfig", mock.Anything, int64(1), int64(1)).Return(nil)

	svc := newBackupSvc(backupRepo, vaultRepo, storageSvc)
	err := svc.DeleteConfig(context.Background(), 1, 1)

	assert.NoError(t, err)
	backupRepo.AssertExpectations(t)
}

// TestBackupService_DeleteConfig_NotFound verifies error when config does not exist.
// TestBackupService_DeleteConfig_NotFound 验证删除不存在的配置时返回错误。
func TestBackupService_DeleteConfig_NotFound(t *testing.T) {
	backupRepo := new(domainmocks.MockBackupRepository)
	vaultRepo := new(domainmocks.MockVaultRepository)
	storageSvc := &backupStorageStub{}

	backupRepo.On("GetByID", mock.Anything, int64(999), int64(1)).Return(nil, nil)

	svc := newBackupSvc(backupRepo, vaultRepo, storageSvc)
	err := svc.DeleteConfig(context.Background(), 1, 999)

	assert.Error(t, err)
	backupRepo.AssertExpectations(t)
}
