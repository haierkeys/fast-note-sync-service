// Package mocks provides testify/mock implementations for service interfaces.
// Package mocks 提供服务接口的 testify/mock 实现。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/stretchr/testify/mock"
)

// MockBackupService is a testify/mock implementation of service.BackupService.
// MockBackupService 是 service.BackupService 的 testify/mock 实现。
type MockBackupService struct {
	mock.Mock
}

// Ensure MockBackupService implements service.BackupService at compile time.
// 编译期确保 MockBackupService 实现了 service.BackupService 接口。
var _ service.BackupService = (*MockBackupService)(nil)

func (m *MockBackupService) GetConfigs(ctx context.Context, uid int64) ([]*dto.BackupConfigDTO, error) {
	args := m.Called(ctx, uid)
	if v := args.Get(0); v != nil {
		return v.([]*dto.BackupConfigDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockBackupService) DeleteConfig(ctx context.Context, uid int64, configID int64) error {
	args := m.Called(ctx, uid, configID)
	return args.Error(0)
}

func (m *MockBackupService) UpdateConfig(ctx context.Context, uid int64, req *dto.BackupConfigRequest) (*dto.BackupConfigDTO, error) {
	args := m.Called(ctx, uid, req)
	if v := args.Get(0); v != nil {
		return v.(*dto.BackupConfigDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockBackupService) ListHistory(ctx context.Context, uid int64, configID int64, pager *app.Pager) ([]*dto.BackupHistoryDTO, int64, error) {
	args := m.Called(ctx, uid, configID, pager)
	if v := args.Get(0); v != nil {
		return v.([]*dto.BackupHistoryDTO), args.Get(1).(int64), args.Error(2)
	}
	return nil, args.Get(1).(int64), args.Error(2)
}

func (m *MockBackupService) ExecuteUserBackup(ctx context.Context, uid int64, configID int64) error {
	args := m.Called(ctx, uid, configID)
	return args.Error(0)
}

func (m *MockBackupService) ExecuteTaskBackups(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockBackupService) NotifyUpdated(uid int64) {
	m.Called(uid)
}

func (m *MockBackupService) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
