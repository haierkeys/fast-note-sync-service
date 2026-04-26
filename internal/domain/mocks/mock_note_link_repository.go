// Package mocks provides testify/mock implementations for domain Repository interfaces.
// Package mocks 提供 domain Repository 接口的 testify/mock 实现。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockNoteLinkRepository is a testify mock for domain.NoteLinkRepository.
// MockNoteLinkRepository 是 domain.NoteLinkRepository 的 testify mock 实现。
type MockNoteLinkRepository struct {
	mock.Mock
}

func (m *MockNoteLinkRepository) CreateBatch(ctx context.Context, links []*domain.NoteLink, uid int64) error {
	args := m.Called(ctx, links, uid)
	return args.Error(0)
}

func (m *MockNoteLinkRepository) DeleteBySourceNoteID(ctx context.Context, sourceNoteID, uid int64) error {
	args := m.Called(ctx, sourceNoteID, uid)
	return args.Error(0)
}

func (m *MockNoteLinkRepository) GetBacklinks(ctx context.Context, targetPathHash string, vaultID, uid int64) ([]*domain.NoteLink, error) {
	args := m.Called(ctx, targetPathHash, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.NoteLink), args.Error(1)
}

func (m *MockNoteLinkRepository) GetBacklinksByHashes(ctx context.Context, targetPathHashes []string, vaultID, uid int64) ([]*domain.NoteLink, error) {
	args := m.Called(ctx, targetPathHashes, vaultID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.NoteLink), args.Error(1)
}

func (m *MockNoteLinkRepository) GetOutlinks(ctx context.Context, sourceNoteID, uid int64) ([]*domain.NoteLink, error) {
	args := m.Called(ctx, sourceNoteID, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.NoteLink), args.Error(1)
}

func (m *MockNoteLinkRepository) DeleteByVaultID(ctx context.Context, vaultID, uid int64) error {
	args := m.Called(ctx, vaultID, uid)
	return args.Error(0)
}

// Compile-time check: MockNoteLinkRepository must implement domain.NoteLinkRepository.
// 编译时检查：MockNoteLinkRepository 必须实现 domain.NoteLinkRepository 接口。
var _ domain.NoteLinkRepository = (*MockNoteLinkRepository)(nil)
