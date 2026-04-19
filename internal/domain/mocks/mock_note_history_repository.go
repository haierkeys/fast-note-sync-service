// Package mocks provides testify/mock implementations for domain Repository interfaces.
// Package mocks 提供 domain Repository 接口的 testify/mock 实现。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockNoteHistoryRepository is a testify mock for domain.NoteHistoryRepository.
// MockNoteHistoryRepository 是 domain.NoteHistoryRepository 的 testify mock 实现。
type MockNoteHistoryRepository struct {
	mock.Mock
}

func (m *MockNoteHistoryRepository) GetByID(ctx context.Context, id, uid int64) (*domain.NoteHistory, error) {
	args := m.Called(ctx, id, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.NoteHistory), args.Error(1)
}

func (m *MockNoteHistoryRepository) GetByNoteIDAndHash(ctx context.Context, noteID int64, contentHash string, uid int64) (*domain.NoteHistory, error) {
	args := m.Called(ctx, noteID, contentHash, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.NoteHistory), args.Error(1)
}

func (m *MockNoteHistoryRepository) Create(ctx context.Context, history *domain.NoteHistory, uid int64) (*domain.NoteHistory, error) {
	args := m.Called(ctx, history, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.NoteHistory), args.Error(1)
}

func (m *MockNoteHistoryRepository) ListByNoteID(ctx context.Context, noteID int64, page, pageSize int, uid int64) ([]*domain.NoteHistory, int64, error) {
	args := m.Called(ctx, noteID, page, pageSize, uid)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*domain.NoteHistory), args.Get(1).(int64), args.Error(2)
}

func (m *MockNoteHistoryRepository) GetLatestVersion(ctx context.Context, noteID, uid int64) (int64, error) {
	args := m.Called(ctx, noteID, uid)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockNoteHistoryRepository) Migrate(ctx context.Context, oldNoteID, newNoteID, uid int64) error {
	args := m.Called(ctx, oldNoteID, newNoteID, uid)
	return args.Error(0)
}

func (m *MockNoteHistoryRepository) GetNoteIDsWithOldHistory(ctx context.Context, cutoffTime int64, uid int64) ([]int64, error) {
	args := m.Called(ctx, cutoffTime, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockNoteHistoryRepository) DeleteOldVersions(ctx context.Context, noteID int64, cutoffTime int64, keepVersions int, uid int64) error {
	args := m.Called(ctx, noteID, cutoffTime, keepVersions, uid)
	return args.Error(0)
}

func (m *MockNoteHistoryRepository) Delete(ctx context.Context, id, uid int64) error {
	args := m.Called(ctx, id, uid)
	return args.Error(0)
}

// Compile-time check: MockNoteHistoryRepository must implement domain.NoteHistoryRepository.
// 编译时检查：MockNoteHistoryRepository 必须实现 domain.NoteHistoryRepository 接口。
var _ domain.NoteHistoryRepository = (*MockNoteHistoryRepository)(nil)
