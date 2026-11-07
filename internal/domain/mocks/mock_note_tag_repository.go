// Package mocks provides testify/mock implementations for domain Repository interfaces.
// Package mocks 提供 domain Repository 接口的 testify/mock 实现。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockNoteTagRepository is a testify mock for domain.NoteTagRepository.
// MockNoteTagRepository 是 domain.NoteTagRepository 的 testify mock 实现。
type MockNoteTagRepository struct {
	mock.Mock
}

func (m *MockNoteTagRepository) ReplaceForNote(ctx context.Context, noteID int64, tags []string, vaultID, uid int64, source string) error {
	args := m.Called(ctx, noteID, tags, vaultID, uid, source)
	return args.Error(0)
}

func (m *MockNoteTagRepository) DeleteByNoteID(ctx context.Context, noteID, uid int64) error {
	args := m.Called(ctx, noteID, uid)
	return args.Error(0)
}

func (m *MockNoteTagRepository) DeleteByVaultID(ctx context.Context, vaultID, uid int64) error {
	args := m.Called(ctx, vaultID, uid)
	return args.Error(0)
}

func (m *MockNoteTagRepository) ListByVault(ctx context.Context, vaultID, uid int64, prefix, scope string, limit int) ([]*domain.TagCount, error) {
	args := m.Called(ctx, vaultID, uid, prefix, scope, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.TagCount), args.Error(1)
}

func (m *MockNoteTagRepository) GetTagsByNoteID(ctx context.Context, noteID, uid int64) ([]string, error) {
	args := m.Called(ctx, noteID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockNoteTagRepository) RebuildForVault(ctx context.Context, vaultID, uid int64) (int, error) {
	args := m.Called(ctx, vaultID, uid)
	return args.Int(0), args.Error(1)
}

// Ensure MockNoteTagRepository implements domain.NoteTagRepository
// 确保 MockNoteTagRepository 实现了 domain.NoteTagRepository
var _ domain.NoteTagRepository = (*MockNoteTagRepository)(nil)
