package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockSyncLogRepository is a testify mock for domain.SyncLogRepository.
type MockSyncLogRepository struct {
	mock.Mock
}

func (m *MockSyncLogRepository) Create(ctx context.Context, log *domain.SyncLog, uid int64) error {
	args := m.Called(ctx, log, uid)
	return args.Error(0)
}

func (m *MockSyncLogRepository) List(ctx context.Context, uid int64, logType, action string, page, pageSize int) ([]*domain.SyncLog, int64, error) {
	args := m.Called(ctx, uid, logType, action, page, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.SyncLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockSyncLogRepository) CleanupByTime(ctx context.Context, timestamp int64, uid int64) error {
	args := m.Called(ctx, timestamp, uid)
	return args.Error(0)
}

func (m *MockSyncLogRepository) CleanupByTimeAll(ctx context.Context, timestamp int64) error {
	args := m.Called(ctx, timestamp)
	return args.Error(0)
}

func (m *MockSyncLogRepository) DeleteByVaultID(ctx context.Context, vaultID, uid int64) error {
	args := m.Called(ctx, vaultID, uid)
	return args.Error(0)
}

var _ domain.SyncLogRepository = (*MockSyncLogRepository)(nil)
