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

// MockNoteService is a testify/mock implementation of service.NoteService.
// MockNoteService 是 service.NoteService 的 testify/mock 实现。
type MockNoteService struct {
	mock.Mock
}

// Ensure MockNoteService implements service.NoteService at compile time.
// 编译期确保 MockNoteService 实现了 service.NoteService 接口。
var _ service.NoteService = (*MockNoteService)(nil)

func (m *MockNoteService) Get(ctx context.Context, uid int64, params *dto.NoteGetRequest) (*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.NoteDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNoteService) UpdateCheck(ctx context.Context, uid int64, params *dto.NoteUpdateCheckRequest) (string, *dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(1); v != nil {
		return args.String(0), v.(*dto.NoteDTO), args.Error(2)
	}
	return args.String(0), nil, args.Error(2)
}

func (m *MockNoteService) ModifyOrCreate(ctx context.Context, uid int64, params *dto.NoteModifyOrCreateRequest, mtimeCheck bool) (bool, *dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params, mtimeCheck)
	if v := args.Get(1); v != nil {
		return args.Bool(0), v.(*dto.NoteDTO), args.Error(2)
	}
	return args.Bool(0), nil, args.Error(2)
}

func (m *MockNoteService) Delete(ctx context.Context, uid int64, params *dto.NoteDeleteRequest) (*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.NoteDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNoteService) Restore(ctx context.Context, uid int64, params *dto.NoteRestoreRequest) (*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.NoteDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNoteService) Rename(ctx context.Context, uid int64, params *dto.NoteRenameRequest) (*dto.NoteDTO, *dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	var old, newN *dto.NoteDTO
	if v := args.Get(0); v != nil {
		old = v.(*dto.NoteDTO)
	}
	if v := args.Get(1); v != nil {
		newN = v.(*dto.NoteDTO)
	}
	return old, newN, args.Error(2)
}

func (m *MockNoteService) List(ctx context.Context, uid int64, params *dto.NoteListRequest, pager *app.Pager) ([]*dto.NoteNoContentDTO, int, error) {
	args := m.Called(ctx, uid, params, pager)
	if v := args.Get(0); v != nil {
		return v.([]*dto.NoteNoContentDTO), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockNoteService) ListByLastTime(ctx context.Context, uid int64, params *dto.NoteSyncRequest) ([]*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.([]*dto.NoteDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNoteService) Sync(ctx context.Context, uid int64, params *dto.NoteSyncRequest) ([]*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.([]*dto.NoteDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNoteService) CountSizeSum(ctx context.Context, vaultID int64, uid int64) error {
	args := m.Called(ctx, vaultID, uid)
	return args.Error(0)
}

func (m *MockNoteService) Cleanup(ctx context.Context, uid int64) error {
	args := m.Called(ctx, uid)
	return args.Error(0)
}

func (m *MockNoteService) CleanupByTime(ctx context.Context, cutoffTime int64) error {
	args := m.Called(ctx, cutoffTime)
	return args.Error(0)
}

func (m *MockNoteService) ListNeedSnapshot(ctx context.Context, uid int64) ([]*dto.NoteDTO, error) {
	args := m.Called(ctx, uid)
	if v := args.Get(0); v != nil {
		return v.([]*dto.NoteDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNoteService) Migrate(ctx context.Context, oldNoteID, newNoteID int64, uid int64) error {
	args := m.Called(ctx, oldNoteID, newNoteID, uid)
	return args.Error(0)
}

func (m *MockNoteService) MigratePush(oldNoteID, newNoteID int64, uid int64) {
	m.Called(oldNoteID, newNoteID, uid)
}

func (m *MockNoteService) WithClient(clientType, name, version string) service.NoteService {
	args := m.Called(clientType, name, version)
	return args.Get(0).(service.NoteService)
}

func (m *MockNoteService) PatchFrontmatter(ctx context.Context, uid int64, params *dto.NotePatchFrontmatterRequest) (*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.NoteDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNoteService) AppendContent(ctx context.Context, uid int64, params *dto.NoteAppendRequest) (*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.NoteDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNoteService) PrependContent(ctx context.Context, uid int64, params *dto.NotePrependRequest) (*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.NoteDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNoteService) ReplaceContent(ctx context.Context, uid int64, params *dto.NoteReplaceRequest) (*dto.NoteReplaceResponse, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.NoteReplaceResponse), args.Error(1)
	}
	return nil, args.Error(1)
}


func (m *MockNoteService) UpdateNoteLinks(ctx context.Context, noteID int64, content string, vaultID, uid int64) {
	m.Called(ctx, noteID, content, vaultID, uid)
}

func (m *MockNoteService) RecycleClear(ctx context.Context, uid int64, params *dto.NoteRecycleClearRequest) error {
	args := m.Called(ctx, uid, params)
	return args.Error(0)
}

func (m *MockNoteService) CleanDuplicateNotes(ctx context.Context, uid int64, vaultID int64) error {
	args := m.Called(ctx, uid, vaultID)
	return args.Error(0)
}
