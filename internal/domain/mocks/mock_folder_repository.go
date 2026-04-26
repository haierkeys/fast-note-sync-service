// Package mocks provides testify/mock implementations for domain Repository interfaces.
// Package mocks 提供 domain Repository 接口的 testify/mock 实现。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockFolderRepository is a testify mock for domain.FolderRepository.
// MockFolderRepository 是 domain.FolderRepository 的 testify mock 实现。
type MockFolderRepository struct {
	mock.Mock
}

func (m *MockFolderRepository) GetByID(ctx context.Context, id, uid int64) (*domain.Folder, error) {
	args := m.Called(ctx, id, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Folder), args.Error(1)
}

func (m *MockFolderRepository) GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*domain.Folder, error) {
	args := m.Called(ctx, pathHash, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Folder), args.Error(1)
}

func (m *MockFolderRepository) GetAllByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) ([]*domain.Folder, error) {
	args := m.Called(ctx, pathHash, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Folder), args.Error(1)
}

func (m *MockFolderRepository) GetByFID(ctx context.Context, fid int64, vaultID, uid int64) ([]*domain.Folder, error) {
	args := m.Called(ctx, fid, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Folder), args.Error(1)
}

func (m *MockFolderRepository) Create(ctx context.Context, folder *domain.Folder, uid int64) (*domain.Folder, error) {
	args := m.Called(ctx, folder, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Folder), args.Error(1)
}

func (m *MockFolderRepository) Update(ctx context.Context, folder *domain.Folder, uid int64) (*domain.Folder, error) {
	args := m.Called(ctx, folder, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Folder), args.Error(1)
}

func (m *MockFolderRepository) Delete(ctx context.Context, id, uid int64) error {
	args := m.Called(ctx, id, uid)
	return args.Error(0)
}

func (m *MockFolderRepository) ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.Folder, error) {
	args := m.Called(ctx, timestamp, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Folder), args.Error(1)
}

func (m *MockFolderRepository) List(ctx context.Context, vaultID int64, uid int64) ([]*domain.Folder, error) {
	args := m.Called(ctx, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Folder), args.Error(1)
}

func (m *MockFolderRepository) ListAll(ctx context.Context, uid int64) ([]*domain.Folder, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Folder), args.Error(1)
}

func (m *MockFolderRepository) DeleteByVaultID(ctx context.Context, vaultID, uid int64) error {
	args := m.Called(ctx, vaultID, uid)
	return args.Error(0)
}

// Compile-time check: MockFolderRepository must implement domain.FolderRepository.
// 编译时检查：MockFolderRepository 必须实现 domain.FolderRepository 接口。
var _ domain.FolderRepository = (*MockFolderRepository)(nil)
