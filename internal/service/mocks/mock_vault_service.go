// Package mocks provides testify/mock implementations for service layer interfaces.
// Package mocks 提供 service 层接口的 testify/mock 实现，供路由层测试使用。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/stretchr/testify/mock"
)

// MockVaultService is a testify mock for service.VaultService.
// MockVaultService 是 service.VaultService 的 testify mock 实现。
type MockVaultService struct {
	mock.Mock
}

// GetByName retrieves vault by name.
// GetByName 根据名称获取 Vault。
func (m *MockVaultService) GetByName(ctx context.Context, uid int64, name string) (*domain.Vault, error) {
	args := m.Called(ctx, uid, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Vault), args.Error(1)
}

// GetOrCreate retrieves or creates vault.
// GetOrCreate 获取或创建 Vault。
func (m *MockVaultService) GetOrCreate(ctx context.Context, uid int64, name string) (*domain.Vault, error) {
	args := m.Called(ctx, uid, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Vault), args.Error(1)
}

// MustGetID retrieves vault ID, returns error if not found.
// MustGetID 获取 Vault ID，不存在则返回错误。
func (m *MockVaultService) MustGetID(ctx context.Context, uid int64, name string) (int64, error) {
	args := m.Called(ctx, uid, name)
	return args.Get(0).(int64), args.Error(1)
}

// Create creates a new vault.
// Create 创建新 Vault。
func (m *MockVaultService) Create(ctx context.Context, uid int64, name string) (*dto.VaultDTO, error) {
	args := m.Called(ctx, uid, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VaultDTO), args.Error(1)
}

// Update updates an existing vault.
// Update 更新现有 Vault。
func (m *MockVaultService) Update(ctx context.Context, uid int64, id int64, name string) (*dto.VaultDTO, error) {
	args := m.Called(ctx, uid, id, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VaultDTO), args.Error(1)
}

// Get retrieves vault by ID.
// Get 根据 ID 获取 Vault。
func (m *MockVaultService) Get(ctx context.Context, uid int64, id int64) (*dto.VaultDTO, error) {
	args := m.Called(ctx, uid, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VaultDTO), args.Error(1)
}

// List retrieves all vaults for a user.
// List 获取用户的 Vault 列表。
func (m *MockVaultService) List(ctx context.Context, uid int64) ([]*dto.VaultDTO, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.VaultDTO), args.Error(1)
}

// Delete deletes a vault.
// Delete 删除 Vault。
func (m *MockVaultService) Delete(ctx context.Context, uid int64, id int64) error {
	args := m.Called(ctx, uid, id)
	return args.Error(0)
}

// UpdateNoteStats updates note statistics for a vault.
// UpdateNoteStats 更新 Vault 的笔记统计信息。
func (m *MockVaultService) UpdateNoteStats(ctx context.Context, noteSize, noteCount, vaultID, uid int64) error {
	args := m.Called(ctx, noteSize, noteCount, vaultID, uid)
	return args.Error(0)
}

// UpdateFileStats updates file statistics for a vault.
// UpdateFileStats 更新 Vault 的文件统计信息。
func (m *MockVaultService) UpdateFileStats(ctx context.Context, fileSize, fileCount, vaultID, uid int64) error {
	args := m.Called(ctx, fileSize, fileCount, vaultID, uid)
	return args.Error(0)
}

// Compile-time check: MockVaultService must implement service.VaultService.
// 编译时检查：MockVaultService 必须实现 service.VaultService 接口。
var _ service.VaultService = (*MockVaultService)(nil)
