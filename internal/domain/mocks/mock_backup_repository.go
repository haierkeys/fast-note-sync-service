// Package mocks provides testify/mock implementations for domain Repository interfaces.
// Package mocks 提供 domain Repository 接口的 testify/mock 实现。
package mocks

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockBackupRepository is a testify mock for domain.BackupRepository.
// MockBackupRepository 是 domain.BackupRepository 的 testify mock 实现。
type MockBackupRepository struct {
	mock.Mock
}

func (m *MockBackupRepository) ListConfigs(ctx context.Context, uid int64) ([]*domain.BackupConfig, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.BackupConfig), args.Error(1)
}

func (m *MockBackupRepository) GetByID(ctx context.Context, id, uid int64) (*domain.BackupConfig, error) {
	args := m.Called(ctx, id, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BackupConfig), args.Error(1)
}

func (m *MockBackupRepository) DeleteConfig(ctx context.Context, id, uid int64) error {
	args := m.Called(ctx, id, uid)
	return args.Error(0)
}

func (m *MockBackupRepository) SaveConfig(ctx context.Context, config *domain.BackupConfig, uid int64) (*domain.BackupConfig, error) {
	args := m.Called(ctx, config, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BackupConfig), args.Error(1)
}

func (m *MockBackupRepository) ListEnabledConfigs(ctx context.Context) ([]*domain.BackupConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.BackupConfig), args.Error(1)
}

func (m *MockBackupRepository) UpdateNextRunTime(ctx context.Context, id, uid int64, nextRun time.Time) error {
	args := m.Called(ctx, id, uid, nextRun)
	return args.Error(0)
}

func (m *MockBackupRepository) CreateHistory(ctx context.Context, history *domain.BackupHistory, uid int64) (*domain.BackupHistory, error) {
	args := m.Called(ctx, history, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BackupHistory), args.Error(1)
}

func (m *MockBackupRepository) ListHistory(ctx context.Context, uid int64, configID int64, page, pageSize int) ([]*domain.BackupHistory, int64, error) {
	args := m.Called(ctx, uid, configID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.BackupHistory), args.Get(1).(int64), args.Error(2)
}

func (m *MockBackupRepository) ListOldHistory(ctx context.Context, uid int64, configID int64, cutoffTime time.Time) ([]*domain.BackupHistory, error) {
	args := m.Called(ctx, uid, configID, cutoffTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.BackupHistory), args.Error(1)
}

func (m *MockBackupRepository) DeleteOldHistory(ctx context.Context, uid int64, configID int64, cutoffTime time.Time) error {
	args := m.Called(ctx, uid, configID, cutoffTime)
	return args.Error(0)
}

// Compile-time check: MockBackupRepository must implement domain.BackupRepository.
// 编译时检查：MockBackupRepository 必须实现 domain.BackupRepository 接口。
var _ domain.BackupRepository = (*MockBackupRepository)(nil)
