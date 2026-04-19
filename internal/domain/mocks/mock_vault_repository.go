// Package mocks provides testify/mock implementations for domain Repository interfaces.
// Package mocks 提供 domain Repository 接口的 testify/mock 实现。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockVaultRepository is a testify mock for domain.VaultRepository.
// MockVaultRepository 是 domain.VaultRepository 的 testify mock 实现。
type MockVaultRepository struct {
	mock.Mock
}

// GetByID retrieves vault by ID and UID.
// GetByID 根据 ID 和 UID 获取 Vault。
func (m *MockVaultRepository) GetByID(ctx context.Context, id, uid int64) (*domain.Vault, error) {
	args := m.Called(ctx, id, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Vault), args.Error(1)
}

// GetByName retrieves vault by name and UID.
// GetByName 根据名称和 UID 获取 Vault。
func (m *MockVaultRepository) GetByName(ctx context.Context, name string, uid int64) (*domain.Vault, error) {
	args := m.Called(ctx, name, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Vault), args.Error(1)
}

// Create creates a new vault.
// Create 创建新的 Vault。
func (m *MockVaultRepository) Create(ctx context.Context, vault *domain.Vault, uid int64) (*domain.Vault, error) {
	args := m.Called(ctx, vault, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Vault), args.Error(1)
}

// Update updates an existing vault.
// Update 更新现有的 Vault。
func (m *MockVaultRepository) Update(ctx context.Context, vault *domain.Vault, uid int64) error {
	args := m.Called(ctx, vault, uid)
	return args.Error(0)
}

// UpdateNoteCountSize updates note count and size for a vault.
// UpdateNoteCountSize 更新 Vault 的笔记数量和大小。
func (m *MockVaultRepository) UpdateNoteCountSize(ctx context.Context, noteSize, noteCount, vaultID, uid int64) error {
	args := m.Called(ctx, noteSize, noteCount, vaultID, uid)
	return args.Error(0)
}

// UpdateFileCountSize updates file count and size for a vault.
// UpdateFileCountSize 更新 Vault 的文件数量和大小。
func (m *MockVaultRepository) UpdateFileCountSize(ctx context.Context, fileSize, fileCount, vaultID, uid int64) error {
	args := m.Called(ctx, fileSize, fileCount, vaultID, uid)
	return args.Error(0)
}

// List retrieves all vaults for a user.
// List 获取用户的所有 Vault 列表。
func (m *MockVaultRepository) List(ctx context.Context, uid int64) ([]*domain.Vault, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Vault), args.Error(1)
}

// Delete soft-deletes a vault.
// Delete 软删除 Vault。
func (m *MockVaultRepository) Delete(ctx context.Context, id, uid int64) error {
	args := m.Called(ctx, id, uid)
	return args.Error(0)
}

// Compile-time check: MockVaultRepository must implement domain.VaultRepository.
// 编译时检查：MockVaultRepository 必须实现 domain.VaultRepository 接口。
var _ domain.VaultRepository = (*MockVaultRepository)(nil)
