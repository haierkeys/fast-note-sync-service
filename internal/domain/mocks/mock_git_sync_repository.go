// Package mocks provides testify/mock implementations for domain Repository interfaces.
// Package mocks 提供 domain Repository 接口的 testify/mock 实现。
package mocks

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockGitSyncRepository is a testify mock for domain.GitSyncRepository.
// MockGitSyncRepository 是 domain.GitSyncRepository 的 testify mock 实现。
type MockGitSyncRepository struct {
	mock.Mock
}

func (m *MockGitSyncRepository) GetByID(ctx context.Context, id, uid int64) (*domain.GitSyncConfig, error) {
	args := m.Called(ctx, id, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GitSyncConfig), args.Error(1)
}

func (m *MockGitSyncRepository) GetByVaultID(ctx context.Context, vaultID, uid int64) (*domain.GitSyncConfig, error) {
	args := m.Called(ctx, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GitSyncConfig), args.Error(1)
}

func (m *MockGitSyncRepository) Save(ctx context.Context, config *domain.GitSyncConfig, uid int64) (*domain.GitSyncConfig, error) {
	args := m.Called(ctx, config, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GitSyncConfig), args.Error(1)
}

func (m *MockGitSyncRepository) Delete(ctx context.Context, id, uid int64) error {
	args := m.Called(ctx, id, uid)
	return args.Error(0)
}

func (m *MockGitSyncRepository) List(ctx context.Context, uid int64) ([]*domain.GitSyncConfig, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.GitSyncConfig), args.Error(1)
}

func (m *MockGitSyncRepository) ListByVaultID(ctx context.Context, vaultID, uid int64) ([]*domain.GitSyncConfig, error) {
	args := m.Called(ctx, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.GitSyncConfig), args.Error(1)
}

func (m *MockGitSyncRepository) ListEnabled(ctx context.Context) ([]*domain.GitSyncConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.GitSyncConfig), args.Error(1)
}

func (m *MockGitSyncRepository) CreateHistory(ctx context.Context, history *domain.GitSyncHistory, uid int64) (*domain.GitSyncHistory, error) {
	args := m.Called(ctx, history, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GitSyncHistory), args.Error(1)
}

func (m *MockGitSyncRepository) ListHistory(ctx context.Context, uid int64, configID int64, page, pageSize int) ([]*domain.GitSyncHistory, int64, error) {
	args := m.Called(ctx, uid, configID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.GitSyncHistory), args.Get(1).(int64), args.Error(2)
}

func (m *MockGitSyncRepository) DeleteHistory(ctx context.Context, uid int64, configID int64) error {
	args := m.Called(ctx, uid, configID)
	return args.Error(0)
}

func (m *MockGitSyncRepository) DeleteOldHistory(ctx context.Context, uid int64, configID int64, cutoffTime time.Time) error {
	args := m.Called(ctx, uid, configID, cutoffTime)
	return args.Error(0)
}

// Compile-time check: MockGitSyncRepository must implement domain.GitSyncRepository.
// 编译时检查：MockGitSyncRepository 必须实现 domain.GitSyncRepository 接口。
var _ domain.GitSyncRepository = (*MockGitSyncRepository)(nil)
