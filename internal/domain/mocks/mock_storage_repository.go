// Package mocks provides testify/mock implementations for domain Repository interfaces.
// Package mocks 提供 domain Repository 接口的 testify/mock 实现。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockStorageRepository is a testify mock for domain.StorageRepository.
// MockStorageRepository 是 domain.StorageRepository 的 testify mock 实现。
type MockStorageRepository struct {
	mock.Mock
}

func (m *MockStorageRepository) GetByID(ctx context.Context, id, uid int64) (*domain.Storage, error) {
	args := m.Called(ctx, id, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Storage), args.Error(1)
}

func (m *MockStorageRepository) Create(ctx context.Context, storage *domain.Storage, uid int64) (*domain.Storage, error) {
	args := m.Called(ctx, storage, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Storage), args.Error(1)
}

func (m *MockStorageRepository) Update(ctx context.Context, storage *domain.Storage, uid int64) (*domain.Storage, error) {
	args := m.Called(ctx, storage, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Storage), args.Error(1)
}

func (m *MockStorageRepository) List(ctx context.Context, uid int64) ([]*domain.Storage, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Storage), args.Error(1)
}

func (m *MockStorageRepository) Delete(ctx context.Context, id, uid int64) error {
	args := m.Called(ctx, id, uid)
	return args.Error(0)
}

// Compile-time check: MockStorageRepository must implement domain.StorageRepository.
// 编译时检查：MockStorageRepository 必须实现 domain.StorageRepository 接口。
var _ domain.StorageRepository = (*MockStorageRepository)(nil)
