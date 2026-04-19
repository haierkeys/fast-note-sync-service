// Package mocks provides testify/mock implementations for service interfaces.
// Package mocks 提供服务接口的 testify/mock 实现。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/stretchr/testify/mock"
)

// MockSettingService is a testify/mock implementation of service.SettingService.
// MockSettingService 是 service.SettingService 的 testify/mock 实现。
type MockSettingService struct {
	mock.Mock
}

// Ensure MockSettingService implements service.SettingService at compile time.
// 编译期确保 MockSettingService 实现了 service.SettingService 接口。
var _ service.SettingService = (*MockSettingService)(nil)

func (m *MockSettingService) UpdateCheck(ctx context.Context, uid int64, params *dto.SettingUpdateCheckRequest) (string, *dto.SettingDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(1); v != nil {
		return args.String(0), v.(*dto.SettingDTO), args.Error(2)
	}
	return args.String(0), nil, args.Error(2)
}

func (m *MockSettingService) ModifyCheck(ctx context.Context, uid int64, params *dto.SettingUpdateCheckRequest) (string, *dto.SettingDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(1); v != nil {
		return args.String(0), v.(*dto.SettingDTO), args.Error(2)
	}
	return args.String(0), nil, args.Error(2)
}

func (m *MockSettingService) ModifyOrCreate(ctx context.Context, uid int64, params *dto.SettingModifyOrCreateRequest, mtimeCheck bool) (bool, *dto.SettingDTO, error) {
	args := m.Called(ctx, uid, params, mtimeCheck)
	if v := args.Get(1); v != nil {
		return args.Bool(0), v.(*dto.SettingDTO), args.Error(2)
	}
	return args.Bool(0), nil, args.Error(2)
}

func (m *MockSettingService) Modify(ctx context.Context, uid int64, params *dto.SettingModifyOrCreateRequest) (bool, *dto.SettingDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(1); v != nil {
		return args.Bool(0), v.(*dto.SettingDTO), args.Error(2)
	}
	return args.Bool(0), nil, args.Error(2)
}

func (m *MockSettingService) Delete(ctx context.Context, uid int64, params *dto.SettingDeleteRequest) (*dto.SettingDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.SettingDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSettingService) Get(ctx context.Context, uid int64, params *dto.SettingGetRequest) (*dto.SettingDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.SettingDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSettingService) ListByLastTime(ctx context.Context, uid int64, params *dto.SettingSyncRequest) ([]*dto.SettingDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.([]*dto.SettingDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSettingService) CleanDuplicateSettings(ctx context.Context, uid int64, vaultID int64) error {
	args := m.Called(ctx, uid, vaultID)
	return args.Error(0)
}

func (m *MockSettingService) Sync(ctx context.Context, uid int64, params *dto.SettingSyncRequest) ([]*dto.SettingDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.([]*dto.SettingDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSettingService) List(ctx context.Context, uid int64, params *dto.SettingListRequest, pager *pkgapp.Pager) ([]*dto.SettingDTO, int64, error) {
	args := m.Called(ctx, uid, params, pager)
	if v := args.Get(0); v != nil {
		return v.([]*dto.SettingDTO), args.Get(1).(int64), args.Error(2)
	}
	return nil, args.Get(1).(int64), args.Error(2)
}

func (m *MockSettingService) Rename(ctx context.Context, uid int64, params *dto.SettingRenameRequest) (*dto.SettingDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.SettingDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSettingService) Cleanup(ctx context.Context, uid int64) error {
	args := m.Called(ctx, uid)
	return args.Error(0)
}

func (m *MockSettingService) CleanupByTime(ctx context.Context, cutoffTime int64) error {
	args := m.Called(ctx, cutoffTime)
	return args.Error(0)
}

func (m *MockSettingService) ClearByVault(ctx context.Context, uid int64, vaultName string) error {
	args := m.Called(ctx, uid, vaultName)
	return args.Error(0)
}
