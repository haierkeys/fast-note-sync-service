// Package mocks provides testify/mock implementations for domain Repository interfaces.
// Package mocks 提供 domain Repository 接口的 testify/mock 实现。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockSettingRepository is a testify mock for domain.SettingRepository.
// MockSettingRepository 是 domain.SettingRepository 的 testify mock 实现。
type MockSettingRepository struct {
	mock.Mock
}

func (m *MockSettingRepository) GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*domain.Setting, error) {
	args := m.Called(ctx, pathHash, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Setting), args.Error(1)
}

func (m *MockSettingRepository) ListByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) ([]*domain.Setting, error) {
	args := m.Called(ctx, pathHash, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Setting), args.Error(1)
}

func (m *MockSettingRepository) Create(ctx context.Context, setting *domain.Setting, uid int64) (*domain.Setting, error) {
	args := m.Called(ctx, setting, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Setting), args.Error(1)
}

func (m *MockSettingRepository) Update(ctx context.Context, setting *domain.Setting, uid int64) (*domain.Setting, error) {
	args := m.Called(ctx, setting, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Setting), args.Error(1)
}

func (m *MockSettingRepository) UpdateMtime(ctx context.Context, mtime int64, id, uid int64) error {
	args := m.Called(ctx, mtime, id, uid)
	return args.Error(0)
}

func (m *MockSettingRepository) UpdateActionMtime(ctx context.Context, action domain.SettingAction, mtime int64, id, uid int64) error {
	args := m.Called(ctx, action, mtime, id, uid)
	return args.Error(0)
}

func (m *MockSettingRepository) Delete(ctx context.Context, id, uid int64) error {
	args := m.Called(ctx, id, uid)
	return args.Error(0)
}

func (m *MockSettingRepository) DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error {
	args := m.Called(ctx, timestamp, uid)
	return args.Error(0)
}

func (m *MockSettingRepository) DeletePhysicalByTimeAll(ctx context.Context, timestamp int64) error {
	args := m.Called(ctx, timestamp)
	return args.Error(0)
}

func (m *MockSettingRepository) List(ctx context.Context, vaultID int64, page, pageSize int, uid int64, keyword string) ([]*domain.Setting, error) {
	args := m.Called(ctx, vaultID, page, pageSize, uid, keyword)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Setting), args.Error(1)
}

func (m *MockSettingRepository) ListCount(ctx context.Context, vaultID, uid int64, keyword string) (int64, error) {
	args := m.Called(ctx, vaultID, uid, keyword)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSettingRepository) ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.Setting, error) {
	args := m.Called(ctx, timestamp, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Setting), args.Error(1)
}

func (m *MockSettingRepository) DeleteByVaultID(ctx context.Context, vaultID, uid int64) error {
	args := m.Called(ctx, vaultID, uid)
	return args.Error(0)
}

// Compile-time check: MockSettingRepository must implement domain.SettingRepository.
// 编译时检查：MockSettingRepository 必须实现 domain.SettingRepository 接口。
var _ domain.SettingRepository = (*MockSettingRepository)(nil)
