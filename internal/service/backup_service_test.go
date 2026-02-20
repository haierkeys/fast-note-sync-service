package service

import (
	"context"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"go.uber.org/zap"
)

// --- Mocks ---

type backupMockBackupRepo struct {
	domain.BackupRepository
	configs []*domain.BackupConfig
	saved   *domain.BackupConfig
}

func (m *backupMockBackupRepo) ListConfigs(ctx context.Context, uid int64) ([]*domain.BackupConfig, error) {
	return m.configs, nil
}

func (m *backupMockBackupRepo) SaveConfig(ctx context.Context, config *domain.BackupConfig, uid int64) (*domain.BackupConfig, error) {
	m.saved = config
	return config, nil
}

func (m *backupMockBackupRepo) GetByID(ctx context.Context, id, uid int64) (*domain.BackupConfig, error) {
	for _, c := range m.configs {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, nil
}

func (m *backupMockBackupRepo) DeleteConfig(ctx context.Context, id, uid int64) error {
	for i, c := range m.configs {
		if c.ID == id {
			m.configs = append(m.configs[:i], m.configs[i+1:]...)
			return nil
		}
	}
	return nil
}

type backupMockNoteRepo struct {
	domain.NoteRepository
}

type backupMockFolderRepo struct {
	domain.FolderRepository
}

type backupMockFileRepo struct {
	domain.FileRepository
}

type backupMockVaultRepo struct {
	domain.VaultRepository
	vaults map[int64]*domain.Vault
}

func (m *backupMockVaultRepo) GetByID(ctx context.Context, id, uid int64) (*domain.Vault, error) {
	if v, ok := m.vaults[id]; ok {
		return v, nil
	}
	return nil, nil
}

func (m *backupMockVaultRepo) GetByName(ctx context.Context, name string, uid int64) (*domain.Vault, error) {
	for _, v := range m.vaults {
		if v.Name == name {
			return v, nil
		}
	}
	return nil, nil
}

type backupMockStorageService struct {
	StorageService
	storages map[int64]*dto.StorageDTO
}

func (m *backupMockStorageService) Get(ctx context.Context, uid int64, id int64) (*dto.StorageDTO, error) {
	if s, ok := m.storages[id]; ok {
		return s, nil
	}
	return nil, nil
}

// --- Tests ---

func TestNewBackupService(t *testing.T) {
	s := NewBackupService(
		&backupMockBackupRepo{},
		&backupMockNoteRepo{},
		&backupMockFolderRepo{},
		&backupMockFileRepo{},
		&backupMockVaultRepo{},
		&backupMockStorageService{},
		&config.StorageConfig{},
		zap.NewNop(),
	)

	if s == nil {
		t.Fatal("NewBackupService returned nil")
	}
}

func TestBackupService_GetConfigs(t *testing.T) {
	mockRepo := &backupMockBackupRepo{
		configs: []*domain.BackupConfig{
			{ID: 1, UID: 1, Type: "full", IsEnabled: true},
			{ID: 2, UID: 1, Type: "incremental", IsEnabled: false},
		},
	}

	s := &backupService{
		backupRepo: mockRepo,
		vaultRepo:  &backupMockVaultRepo{vaults: make(map[int64]*domain.Vault)},
	}

	configs, err := s.GetConfigs(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetConfigs failed: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("expected 2 configs, got %d", len(configs))
	}

	if configs[0].ID != 1 || configs[1].ID != 2 {
		t.Errorf("config IDs mismatch")
	}
}

func TestBackupService_UpdateConfig(t *testing.T) {
	vaults := make(map[int64]*domain.Vault)
	vaults[100] = &domain.Vault{ID: 100, Name: "myvault"}

	storages := make(map[int64]*dto.StorageDTO)
	storages[200] = &dto.StorageDTO{ID: 200, IsEnabled: true}

	mockBackup := &backupMockBackupRepo{}
	mockVault := &backupMockVaultRepo{vaults: vaults}
	mockStorage := &backupMockStorageService{storages: storages}

	s := &backupService{
		backupRepo:     mockBackup,
		vaultRepo:      mockVault,
		storageService: mockStorage,
		logger:         zap.NewNop(),
	}

	req := &dto.BackupConfigRequest{
		Vault:      "myvault",
		StorageIds: "[200]",
		Type:       "full",
		IsEnabled:  true,
	}

	res, err := s.UpdateConfig(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("UpdateConfig failed: %v", err)
	}

	if res.Vault != "myvault" {
		t.Errorf("expected vault myvault, got %s", res.Vault)
	}

	if mockBackup.saved == nil {
		t.Fatal("config was not saved to repository")
	}

	if mockBackup.saved.VaultID != 100 {
		t.Errorf("expected vault ID 100, got %d", mockBackup.saved.VaultID)
	}
}

func TestBackupService_DeleteConfig(t *testing.T) {
	mockRepo := &backupMockBackupRepo{
		configs: []*domain.BackupConfig{
			{ID: 1, UID: 1},
		},
	}

	s := &backupService{
		backupRepo: mockRepo,
	}

	// 1. Delete existing
	err := s.DeleteConfig(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("DeleteConfig failed: %v", err)
	}
	if len(mockRepo.configs) != 0 {
		t.Errorf("expected 0 configs after delete, got %d", len(mockRepo.configs))
	}

	// 2. Delete non-existing
	err = s.DeleteConfig(context.Background(), 1, 999)
	if err == nil {
		t.Error("expected error when deleting non-existing config, got nil")
	}
}
