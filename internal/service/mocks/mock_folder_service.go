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

// MockFolderService is a testify/mock implementation of service.FolderService.
// MockFolderService 是 service.FolderService 的 testify/mock 实现。
type MockFolderService struct {
	mock.Mock
}

// Ensure MockFolderService implements service.FolderService at compile time.
// 编译期确保 MockFolderService 实现了 service.FolderService 接口。
var _ service.FolderService = (*MockFolderService)(nil)

func (m *MockFolderService) Get(ctx context.Context, uid int64, params *dto.FolderGetRequest) (*dto.FolderDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.FolderDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockFolderService) List(ctx context.Context, uid int64, params *dto.FolderListRequest) ([]*dto.FolderDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.([]*dto.FolderDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockFolderService) ListByUpdatedTimestamp(ctx context.Context, uid int64, vault string, lastTime int64) ([]*dto.FolderDTO, error) {
	args := m.Called(ctx, uid, vault, lastTime)
	if v := args.Get(0); v != nil {
		return v.([]*dto.FolderDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockFolderService) UpdateOrCreate(ctx context.Context, uid int64, params *dto.FolderCreateRequest) (*dto.FolderDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.FolderDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockFolderService) Delete(ctx context.Context, uid int64, params *dto.FolderDeleteRequest) (*dto.FolderDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.FolderDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockFolderService) Rename(ctx context.Context, uid int64, params *dto.FolderRenameRequest) (*dto.FolderDTO, *dto.FolderDTO, error) {
	args := m.Called(ctx, uid, params)
	var old, newF *dto.FolderDTO
	if v := args.Get(0); v != nil {
		old = v.(*dto.FolderDTO)
	}
	if v := args.Get(1); v != nil {
		newF = v.(*dto.FolderDTO)
	}
	return old, newF, args.Error(2)
}

func (m *MockFolderService) ListNotes(ctx context.Context, uid int64, params *dto.FolderContentRequest, pager *app.Pager) ([]*dto.NoteNoContentDTO, int, error) {
	args := m.Called(ctx, uid, params, pager)
	if v := args.Get(0); v != nil {
		return v.([]*dto.NoteNoContentDTO), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockFolderService) ListFiles(ctx context.Context, uid int64, params *dto.FolderContentRequest, pager *app.Pager) ([]*dto.FileDTO, int, error) {
	args := m.Called(ctx, uid, params, pager)
	if v := args.Get(0); v != nil {
		return v.([]*dto.FileDTO), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockFolderService) EnsurePathFID(ctx context.Context, uid int64, vaultID int64, path string) (int64, error) {
	args := m.Called(ctx, uid, vaultID, path)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFolderService) SyncResourceFID(ctx context.Context, uid int64, vaultID int64, noteIDs []int64, fileIDs []int64) error {
	args := m.Called(ctx, uid, vaultID, noteIDs, fileIDs)
	return args.Error(0)
}

func (m *MockFolderService) GetTree(ctx context.Context, uid int64, params *dto.FolderTreeRequest) (*dto.FolderTreeResponse, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.FolderTreeResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockFolderService) CleanDuplicateFolders(ctx context.Context, uid int64, vaultID int64) error {
	args := m.Called(ctx, uid, vaultID)
	return args.Error(0)
}

func (m *MockFolderService) WithClient(clientType, name, version string) service.FolderService {
	args := m.Called(clientType, name, version)
	return args.Get(0).(service.FolderService)
}
