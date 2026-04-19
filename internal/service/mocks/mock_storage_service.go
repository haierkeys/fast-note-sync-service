// Package mocks provides testify/mock implementations for service interfaces.
// Package mocks 提供服务接口的 testify/mock 实现。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/stretchr/testify/mock"
)

// MockStorageService is a testify/mock implementation of service.StorageService.
// MockStorageService 是 service.StorageService 的 testify/mock 实现。
type MockStorageService struct {
	mock.Mock
}

// Ensure MockStorageService implements service.StorageService at compile time.
// 编译期确保 MockStorageService 实现了 service.StorageService 接口。
var _ service.StorageService = (*MockStorageService)(nil)

func (m *MockStorageService) CreateOrUpdate(ctx context.Context, uid int64, id int64, storageDTO *dto.StoragePostRequest) (*dto.StorageDTO, error) {
	args := m.Called(ctx, uid, id, storageDTO)
	if v := args.Get(0); v != nil {
		return v.(*dto.StorageDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStorageService) Get(ctx context.Context, uid int64, id int64) (*dto.StorageDTO, error) {
	args := m.Called(ctx, uid, id)
	if v := args.Get(0); v != nil {
		return v.(*dto.StorageDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStorageService) List(ctx context.Context, uid int64) ([]*dto.StorageDTO, error) {
	args := m.Called(ctx, uid)
	if v := args.Get(0); v != nil {
		return v.([]*dto.StorageDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStorageService) Delete(ctx context.Context, uid int64, id int64) error {
	args := m.Called(ctx, uid, id)
	return args.Error(0)
}

func (m *MockStorageService) GetEnabledTypes() ([]string, error) {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]string), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStorageService) Validate(ctx context.Context, req *dto.StoragePostRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}
