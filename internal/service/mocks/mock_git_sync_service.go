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

// MockGitSyncService is a testify/mock implementation of service.GitSyncService.
// MockGitSyncService 是 service.GitSyncService 的 testify/mock 实现。
type MockGitSyncService struct {
	mock.Mock
}

// Ensure MockGitSyncService implements service.GitSyncService at compile time.
// 编译期确保 MockGitSyncService 实现了 service.GitSyncService 接口。
var _ service.GitSyncService = (*MockGitSyncService)(nil)

func (m *MockGitSyncService) GetConfigs(ctx context.Context, uid int64) ([]*dto.GitSyncConfigDTO, error) {
	args := m.Called(ctx, uid)
	if v := args.Get(0); v != nil {
		return v.([]*dto.GitSyncConfigDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGitSyncService) GetConfig(ctx context.Context, uid int64, vaultID int64) (*dto.GitSyncConfigDTO, error) {
	args := m.Called(ctx, uid, vaultID)
	if v := args.Get(0); v != nil {
		return v.(*dto.GitSyncConfigDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGitSyncService) UpdateConfig(ctx context.Context, uid int64, params *dto.GitSyncConfigRequest) (*dto.GitSyncConfigDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.GitSyncConfigDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockGitSyncService) DeleteConfig(ctx context.Context, uid int64, id int64) error {
	args := m.Called(ctx, uid, id)
	return args.Error(0)
}

func (m *MockGitSyncService) Validate(ctx context.Context, params *dto.GitSyncValidateRequest) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockGitSyncService) ExecuteSync(ctx context.Context, uid int64, id int64) error {
	args := m.Called(ctx, uid, id)
	return args.Error(0)
}

func (m *MockGitSyncService) CleanWorkspace(ctx context.Context, uid int64, configID int64) error {
	args := m.Called(ctx, uid, configID)
	return args.Error(0)
}

func (m *MockGitSyncService) ListHistory(ctx context.Context, uid int64, configID int64, pager *pkgapp.Pager) ([]*dto.GitSyncHistoryDTO, int64, error) {
	args := m.Called(ctx, uid, configID, pager)
	if v := args.Get(0); v != nil {
		return v.([]*dto.GitSyncHistoryDTO), args.Get(1).(int64), args.Error(2)
	}
	return nil, args.Get(1).(int64), args.Error(2)
}

func (m *MockGitSyncService) NotifyUpdated(uid int64, vaultID int64) {
	m.Called(uid, vaultID)
}

func (m *MockGitSyncService) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
