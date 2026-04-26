// Package mocks provides testify/mock implementations for domain Repository interfaces.
// Package mocks 提供 domain Repository 接口的 testify/mock 实现。
package mocks

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockNoteRepository is a testify mock for domain.NoteRepository.
// MockNoteRepository 是 domain.NoteRepository 的 testify mock 实现。
type MockNoteRepository struct {
	mock.Mock
}

func (m *MockNoteRepository) GetByID(ctx context.Context, id, uid int64) (*domain.Note, error) {
	args := m.Called(ctx, id, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*domain.Note, error) {
	args := m.Called(ctx, pathHash, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) GetByPathHashIncludeRecycle(ctx context.Context, pathHash string, vaultID, uid int64, isRecycle bool) (*domain.Note, error) {
	args := m.Called(ctx, pathHash, vaultID, uid, isRecycle)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) GetAllByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*domain.Note, error) {
	args := m.Called(ctx, pathHash, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) ListByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) ([]*domain.Note, error) {
	args := m.Called(ctx, pathHash, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) GetByPath(ctx context.Context, path string, vaultID, uid int64) (*domain.Note, error) {
	args := m.Called(ctx, path, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) Create(ctx context.Context, note *domain.Note, uid int64) (*domain.Note, error) {
	args := m.Called(ctx, note, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) Update(ctx context.Context, note *domain.Note, uid int64) (*domain.Note, error) {
	args := m.Called(ctx, note, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) UpdateDelete(ctx context.Context, note *domain.Note, uid int64) error {
	args := m.Called(ctx, note, uid)
	return args.Error(0)
}

func (m *MockNoteRepository) UpdateMtime(ctx context.Context, mtime int64, id, uid int64) error {
	args := m.Called(ctx, mtime, id, uid)
	return args.Error(0)
}

func (m *MockNoteRepository) UpdateActionMtime(ctx context.Context, action domain.NoteAction, mtime int64, id, uid int64) error {
	args := m.Called(ctx, action, mtime, id, uid)
	return args.Error(0)
}

func (m *MockNoteRepository) UpdateFID(ctx context.Context, id, fid, uid int64) error {
	args := m.Called(ctx, id, fid, uid)
	return args.Error(0)
}

func (m *MockNoteRepository) UpdateSnapshot(ctx context.Context, snapshot, snapshotHash string, version, id, uid int64) error {
	args := m.Called(ctx, snapshot, snapshotHash, version, id, uid)
	return args.Error(0)
}

func (m *MockNoteRepository) Delete(ctx context.Context, id, uid int64) error {
	args := m.Called(ctx, id, uid)
	return args.Error(0)
}

func (m *MockNoteRepository) DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error {
	args := m.Called(ctx, timestamp, uid)
	return args.Error(0)
}

func (m *MockNoteRepository) DeletePhysicalByTimeAll(ctx context.Context, timestamp int64) error {
	args := m.Called(ctx, timestamp)
	return args.Error(0)
}

func (m *MockNoteRepository) List(ctx context.Context, vaultID int64, page, pageSize int, uid int64, keyword string, isRecycle bool, searchMode string, searchContent bool, sortBy string, sortOrder string, paths []string) ([]*domain.Note, error) {
	args := m.Called(ctx, vaultID, page, pageSize, uid, keyword, isRecycle, searchMode, searchContent, sortBy, sortOrder, paths)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) ListCount(ctx context.Context, vaultID, uid int64, keyword string, isRecycle bool, searchMode string, searchContent bool, paths []string) (int64, error) {
	args := m.Called(ctx, vaultID, uid, keyword, isRecycle, searchMode, searchContent, paths)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockNoteRepository) ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.Note, error) {
	args := m.Called(ctx, timestamp, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) ListByUpdatedTimestampPage(ctx context.Context, timestamp, vaultID, uid int64, offset, limit int) ([]*domain.Note, error) {
	args := m.Called(ctx, timestamp, vaultID, uid, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) ListContentUnchanged(ctx context.Context, uid int64) ([]*domain.Note, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) CountSizeSum(ctx context.Context, vaultID, uid int64) (*domain.CountSizeResult, error) {
	args := m.Called(ctx, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CountSizeResult), args.Error(1)
}

func (m *MockNoteRepository) ListByFID(ctx context.Context, fid, vaultID, uid int64, page, pageSize int, sortBy, sortOrder string) ([]*domain.Note, error) {
	args := m.Called(ctx, fid, vaultID, uid, page, pageSize, sortBy, sortOrder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) ListByFIDCount(ctx context.Context, fid, vaultID, uid int64) (int64, error) {
	args := m.Called(ctx, fid, vaultID, uid)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockNoteRepository) ListByFIDs(ctx context.Context, fids []int64, vaultID, uid int64, page, pageSize int, sortBy, sortOrder string) ([]*domain.Note, error) {
	args := m.Called(ctx, fids, vaultID, uid, page, pageSize, sortBy, sortOrder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) ListByFIDsCount(ctx context.Context, fids []int64, vaultID, uid int64) (int64, error) {
	args := m.Called(ctx, fids, vaultID, uid)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockNoteRepository) ListByIDs(ctx context.Context, ids []int64, uid int64) ([]*domain.Note, error) {
	args := m.Called(ctx, ids, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) ListByPathPrefix(ctx context.Context, pathPrefix string, vaultID, uid int64) ([]*domain.Note, error) {
	args := m.Called(ctx, pathPrefix, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Note), args.Error(1)
}

func (m *MockNoteRepository) RecycleClear(ctx context.Context, path, pathHash string, vaultID, uid int64) error {
	args := m.Called(path, pathHash, vaultID, uid)
	return args.Error(0)
}

func (m *MockNoteRepository) DeleteByVaultID(ctx context.Context, vaultID, uid int64) error {
	args := m.Called(ctx, vaultID, uid)
	return args.Error(0)
}

// Compile-time check: MockNoteRepository must implement domain.NoteRepository.
// 编译时检查：MockNoteRepository 必须实现 domain.NoteRepository 接口。
var _ domain.NoteRepository = (*MockNoteRepository)(nil)

// suppress unused import warning for time
var _ = time.Now
