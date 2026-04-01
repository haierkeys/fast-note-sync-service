package service

import (
	"context"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/mocks"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestBackupService_GetConfigs(t *testing.T) {
	mockRepo := new(mocks.BackupRepositoryMock)
	mockVaultRepo := new(mocks.VaultRepositoryMock)
	
	s := &backupService{
		backupRepo: mockRepo,
		vaultRepo:  mockVaultRepo,
	}

	ctx := context.Background()
	uid := int64(1)

	t.Run("Success", func(t *testing.T) {
		configs := []*domain.BackupConfig{
			{ID: 1, UID: uid, Type: "full", IsEnabled: true},
			{ID: 2, UID: uid, Type: "incremental", IsEnabled: false},
		}

		mockRepo.On("ListConfigs", ctx, uid).Return(configs, nil)

		res, err := s.GetConfigs(ctx, uid)

		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, int64(1), res[0].ID)
		assert.Equal(t, int64(2), res[1].ID)
		mockRepo.AssertExpectations(t)
	})
}

func TestBackupService_UpdateConfig(t *testing.T) {
	mockRepo := new(mocks.BackupRepositoryMock)
	mockVaultRepo := new(mocks.VaultRepositoryMock)
	mockStorageSvc := new(mocks.StorageServiceMock)
	
	s := &backupService{
		backupRepo:     mockRepo,
		vaultRepo:      mockVaultRepo,
		storageService: mockStorageSvc,
		logger:         zap.NewNop(),
	}

	ctx := context.Background()
	uid := int64(1)

	t.Run("CreateSuccess", func(t *testing.T) {
		req := &dto.BackupConfigRequest{
			Vault:      "myvault",
			StorageIds: "[200]",
			Type:       "full",
			IsEnabled:  true,
		}

		mockVaultRepo.On("GetByName", ctx, "myvault", uid).Return(&domain.Vault{ID: 100, Name: "myvault"}, nil)
		mockStorageSvc.On("Get", ctx, uid, int64(200)).Return(&dto.StorageDTO{ID: 200, IsEnabled: true}, nil)
		
		expectedConfig := &domain.BackupConfig{
			UID:              uid,
			VaultID:          100,
			Type:             "full",
			StorageIds:       "[200]",
			IsEnabled:        true,
		}
		
		mockRepo.On("SaveConfig", ctx, mock.MatchedBy(func(cfg *domain.BackupConfig) bool {
			return cfg.UID == uid && cfg.VaultID == 100 && cfg.Type == "full"
		}), uid).Return(expectedConfig, nil)

		res, err := s.UpdateConfig(ctx, uid, req)

		assert.NoError(t, err)
		assert.Equal(t, "myvault", res.Vault)
		mockVaultRepo.AssertExpectations(t)
		mockStorageSvc.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func TestBackupService_DeleteConfig(t *testing.T) {
	mockRepo := new(mocks.BackupRepositoryMock)
	
	s := &backupService{
		backupRepo: mockRepo,
	}

	ctx := context.Background()
	uid := int64(1)
	configID := int64(10)

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, configID, uid).Return(&domain.BackupConfig{ID: configID, UID: uid}, nil)
		mockRepo.On("DeleteConfig", ctx, configID, uid).Return(nil)

		err := s.DeleteConfig(ctx, uid, configID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, configID, uid).Return(nil, nil)

		err := s.DeleteConfig(ctx, uid, configID)

		assert.Error(t, err)
		assert.Equal(t, code.ErrorBackupConfigNotFound, err)
	})
}

func TestNewBackupService(t *testing.T) {
	mockBackupRepo := new(mocks.BackupRepositoryMock)
	mockNoteRepo := new(mocks.NoteRepositoryMock)
	mockFolderRepo := new(mocks.FolderRepositoryMock)
	mockFileRepo := new(mocks.FileRepositoryMock)
	mockVaultRepo := new(mocks.VaultRepositoryMock)
	mockStorageSvc := new(mocks.StorageServiceMock)

	s := NewBackupService(
		mockBackupRepo,
		mockNoteRepo,
		mockFolderRepo,
		mockFileRepo,
		mockVaultRepo,
		mockStorageSvc,
		&config.StorageConfig{},
		zap.NewNop(),
	)

	assert.NotNil(t, s)
}
