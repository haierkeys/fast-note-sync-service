// Package mocks provides testify/mock implementations for domain Repository interfaces.
// Package mocks 提供 domain Repository 接口的 testify/mock 实现。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockFileRepository is a testify mock for domain.FileRepository.
// MockFileRepository 是 domain.FileRepository 的 testify mock 实现。
type MockFileRepository struct {
	mock.Mock
}

func (m *MockFileRepository) GetByID(ctx context.Context, id, uid int64) (*domain.File, error) {
	args := m.Called(ctx, id, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.File), args.Error(1)
}

func (m *MockFileRepository) GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*domain.File, error) {
	args := m.Called(ctx, pathHash, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.File), args.Error(1)
}

func (m *MockFileRepository) ListByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) ([]*domain.File, error) {
	args := m.Called(ctx, pathHash, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.File), args.Error(1)
}

func (m *MockFileRepository) GetByPath(ctx context.Context, path string, vaultID, uid int64) (*domain.File, error) {
	args := m.Called(ctx, path, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.File), args.Error(1)
}

func (m *MockFileRepository) GetByPathLike(ctx context.Context, path string, vaultID, uid int64) (*domain.File, error) {
	args := m.Called(ctx, path, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.File), args.Error(1)
}

func (m *MockFileRepository) Create(ctx context.Context, file *domain.File, uid int64) (*domain.File, error) {
	args := m.Called(ctx, file, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.File), args.Error(1)
}

func (m *MockFileRepository) Update(ctx context.Context, file *domain.File, uid int64) (*domain.File, error) {
	args := m.Called(ctx, file, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.File), args.Error(1)
}

func (m *MockFileRepository) UpdateMtime(ctx context.Context, mtime int64, id, uid int64) error {
	args := m.Called(ctx, mtime, id, uid)
	return args.Error(0)
}

func (m *MockFileRepository) UpdateActionMtime(ctx context.Context, action domain.FileAction, mtime int64, id, uid int64) error {
	args := m.Called(ctx, action, mtime, id, uid)
	return args.Error(0)
}

func (m *MockFileRepository) UpdateFID(ctx context.Context, id, fid, uid int64) error {
	args := m.Called(ctx, id, fid, uid)
	return args.Error(0)
}

func (m *MockFileRepository) Delete(ctx context.Context, id, uid int64) error {
	args := m.Called(ctx, id, uid)
	return args.Error(0)
}

func (m *MockFileRepository) DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error {
	args := m.Called(ctx, timestamp, uid)
	return args.Error(0)
}

func (m *MockFileRepository) DeletePhysicalByTimeAll(ctx context.Context, timestamp int64) error {
	args := m.Called(ctx, timestamp)
	return args.Error(0)
}

func (m *MockFileRepository) List(ctx context.Context, vaultID int64, page, pageSize int, uid int64, keyword string, isRecycle bool, sortBy string, sortOrder string) ([]*domain.File, error) {
	args := m.Called(ctx, vaultID, page, pageSize, uid, keyword, isRecycle, sortBy, sortOrder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.File), args.Error(1)
}

func (m *MockFileRepository) ListCount(ctx context.Context, vaultID, uid int64, keyword string, isRecycle bool) (int64, error) {
	args := m.Called(ctx, vaultID, uid, keyword, isRecycle)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFileRepository) ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.File, error) {
	args := m.Called(ctx, timestamp, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.File), args.Error(1)
}

func (m *MockFileRepository) ListByUpdatedTimestampPage(ctx context.Context, timestamp, vaultID, uid int64, offset, limit int) ([]*domain.File, error) {
	args := m.Called(ctx, timestamp, vaultID, uid, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.File), args.Error(1)
}

func (m *MockFileRepository) ListByMtime(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.File, error) {
	args := m.Called(ctx, timestamp, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.File), args.Error(1)
}

func (m *MockFileRepository) CountSizeSum(ctx context.Context, vaultID, uid int64) (*domain.CountSizeResult, error) {
	args := m.Called(ctx, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CountSizeResult), args.Error(1)
}

func (m *MockFileRepository) ListByFID(ctx context.Context, fid, vaultID, uid int64, page, pageSize int, sortBy, sortOrder string) ([]*domain.File, error) {
	args := m.Called(ctx, fid, vaultID, uid, page, pageSize, sortBy, sortOrder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.File), args.Error(1)
}

func (m *MockFileRepository) ListByFIDCount(ctx context.Context, fid, vaultID, uid int64) (int64, error) {
	args := m.Called(ctx, fid, vaultID, uid)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFileRepository) ListByFIDs(ctx context.Context, fids []int64, vaultID, uid int64, page, pageSize int, sortBy, sortOrder string) ([]*domain.File, error) {
	args := m.Called(ctx, fids, vaultID, uid, page, pageSize, sortBy, sortOrder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.File), args.Error(1)
}

func (m *MockFileRepository) ListByFIDsCount(ctx context.Context, fids []int64, vaultID, uid int64) (int64, error) {
	args := m.Called(ctx, fids, vaultID, uid)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFileRepository) ListByIDs(ctx context.Context, ids []int64, uid int64) ([]*domain.File, error) {
	args := m.Called(ctx, ids, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.File), args.Error(1)
}

func (m *MockFileRepository) ListByPathPrefix(ctx context.Context, pathPrefix string, vaultID, uid int64) ([]*domain.File, error) {
	args := m.Called(ctx, pathPrefix, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.File), args.Error(1)
}

func (m *MockFileRepository) DeleteByVaultID(ctx context.Context, vaultID, uid int64) error {
	args := m.Called(ctx, vaultID, uid)
	return args.Error(0)
}

func (m *MockFileRepository) RecycleClear(ctx context.Context, path, pathHash string, vaultID, uid int64) error {
	args := m.Called(ctx, path, pathHash, vaultID, uid)
	return args.Error(0)
}

// Compile-time check: MockFileRepository must implement domain.FileRepository.
// 编译时检查：MockFileRepository 必须实现 domain.FileRepository 接口。
var _ domain.FileRepository = (*MockFileRepository)(nil)
