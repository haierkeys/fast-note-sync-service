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

// MockNoteHistoryService is a testify/mock implementation of service.NoteHistoryService.
// MockNoteHistoryService 是 service.NoteHistoryService 的 testify/mock 实现。
type MockNoteHistoryService struct {
	mock.Mock
}

// Ensure MockNoteHistoryService implements service.NoteHistoryService at compile time.
// 编译期确保 MockNoteHistoryService 实现了 service.NoteHistoryService 接口。
var _ service.NoteHistoryService = (*MockNoteHistoryService)(nil)

func (m *MockNoteHistoryService) Get(ctx context.Context, uid int64, id int64) (*dto.NoteHistoryDTO, error) {
	args := m.Called(ctx, uid, id)
	if v := args.Get(0); v != nil {
		return v.(*dto.NoteHistoryDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNoteHistoryService) GetByNoteIDAndHash(ctx context.Context, uid int64, noteID int64, contentHash string) (*dto.NoteHistoryDTO, error) {
	args := m.Called(ctx, uid, noteID, contentHash)
	if v := args.Get(0); v != nil {
		return v.(*dto.NoteHistoryDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNoteHistoryService) List(ctx context.Context, uid int64, params *dto.NoteHistoryListRequest, pager *app.Pager) ([]*dto.NoteHistoryNoContentDTO, int64, error) {
	args := m.Called(ctx, uid, params, pager)
	if v := args.Get(0); v != nil {
		return v.([]*dto.NoteHistoryNoContentDTO), args.Get(1).(int64), args.Error(2)
	}
	return nil, args.Get(1).(int64), args.Error(2)
}

func (m *MockNoteHistoryService) RestoreFromHistory(ctx context.Context, uid int64, historyID int64) (*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, historyID)
	if v := args.Get(0); v != nil {
		return v.(*dto.NoteDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNoteHistoryService) ProcessDelay(ctx context.Context, noteID int64, uid int64) error {
	args := m.Called(ctx, noteID, uid)
	return args.Error(0)
}

func (m *MockNoteHistoryService) Migrate(ctx context.Context, oldNoteID, newNoteID int64, uid int64) error {
	args := m.Called(ctx, oldNoteID, newNoteID, uid)
	return args.Error(0)
}

func (m *MockNoteHistoryService) CleanupByTime(ctx context.Context, cutoffTime int64, keepVersions int) error {
	args := m.Called(ctx, cutoffTime, keepVersions)
	return args.Error(0)
}
