// Package mocks provides testify/mock implementations for service interfaces.
// Package mocks 提供服务接口的 testify/mock 实现。
package mocks

import (
	"context"
	"io"

	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/stretchr/testify/mock"
)

// MockFileService is a testify/mock implementation of service.FileService.
// MockFileService 是 service.FileService 的 testify/mock 实现。
type MockFileService struct {
	mock.Mock
}

// Ensure MockFileService implements service.FileService at compile time.
// 编译期确保 MockFileService 实现了 service.FileService 接口。
var _ service.FileService = (*MockFileService)(nil)

func (m *MockFileService) Get(ctx context.Context, uid int64, params *dto.FileGetRequest) (*dto.FileDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.FileDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockFileService) UpdateCheck(ctx context.Context, uid int64, params *dto.FileUpdateCheckRequest) (string, *dto.FileDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(1); v != nil {
		return args.String(0), v.(*dto.FileDTO), args.Error(2)
	}
	return args.String(0), nil, args.Error(2)
}

func (m *MockFileService) UploadCheck(ctx context.Context, uid int64, params *dto.FileUpdateCheckRequest) (string, *dto.FileDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(1); v != nil {
		return args.String(0), v.(*dto.FileDTO), args.Error(2)
	}
	return args.String(0), nil, args.Error(2)
}

func (m *MockFileService) UpdateOrCreate(ctx context.Context, uid int64, params *dto.FileUpdateRequest, mtimeCheck bool) (bool, *dto.FileDTO, error) {
	args := m.Called(ctx, uid, params, mtimeCheck)
	if v := args.Get(1); v != nil {
		return args.Bool(0), v.(*dto.FileDTO), args.Error(2)
	}
	return args.Bool(0), nil, args.Error(2)
}

func (m *MockFileService) UploadComplete(ctx context.Context, uid int64, params *dto.FileUpdateRequest) (bool, *dto.FileDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(1); v != nil {
		return args.Bool(0), v.(*dto.FileDTO), args.Error(2)
	}
	return args.Bool(0), nil, args.Error(2)
}

func (m *MockFileService) Delete(ctx context.Context, uid int64, params *dto.FileDeleteRequest) (*dto.FileDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.FileDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockFileService) List(ctx context.Context, uid int64, params *dto.FileListRequest, pager *app.Pager) ([]*dto.FileDTO, int, error) {
	args := m.Called(ctx, uid, params, pager)
	if v := args.Get(0); v != nil {
		return v.([]*dto.FileDTO), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockFileService) ListByLastTime(ctx context.Context, uid int64, params *dto.FileSyncRequest) ([]*dto.FileDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.([]*dto.FileDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockFileService) CountSizeSum(ctx context.Context, vaultID int64, uid int64) error {
	args := m.Called(ctx, vaultID, uid)
	return args.Error(0)
}

func (m *MockFileService) Cleanup(ctx context.Context, uid int64) error {
	args := m.Called(ctx, uid)
	return args.Error(0)
}

func (m *MockFileService) CleanupByTime(ctx context.Context, cutoffTime int64) error {
	args := m.Called(ctx, cutoffTime)
	return args.Error(0)
}

func (m *MockFileService) ResolveEmbedLinks(ctx context.Context, uid int64, vaultName string, notePath string, content string) (map[string]string, error) {
	args := m.Called(ctx, uid, vaultName, notePath, content)
	if v := args.Get(0); v != nil {
		return v.(map[string]string), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockFileService) GetContent(ctx context.Context, uid int64, params *dto.FileGetRequest) (io.ReadCloser, string, int64, string, error) {
	args := m.Called(ctx, uid, params)
	var rc io.ReadCloser
	if v := args.Get(0); v != nil {
		rc = v.(io.ReadCloser)
	}
	return rc, args.String(1), args.Get(2).(int64), args.String(3), args.Error(4)
}

func (m *MockFileService) GetContentInfo(ctx context.Context, uid int64, params *dto.FileGetRequest) (string, string, int64, string, string, error) {
	args := m.Called(ctx, uid, params)
	return args.String(0), args.String(1), args.Get(2).(int64), args.String(3), args.String(4), args.Error(5)
}

func (m *MockFileService) Restore(ctx context.Context, uid int64, params *dto.FileRestoreRequest) (*dto.FileDTO, error) {
	args := m.Called(ctx, uid, params)
	if v := args.Get(0); v != nil {
		return v.(*dto.FileDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockFileService) Rename(ctx context.Context, uid int64, params *dto.FileRenameRequest) (*dto.FileDTO, *dto.FileDTO, error) {
	args := m.Called(ctx, uid, params)
	var old, newF *dto.FileDTO
	if v := args.Get(0); v != nil {
		old = v.(*dto.FileDTO)
	}
	if v := args.Get(1); v != nil {
		newF = v.(*dto.FileDTO)
	}
	return old, newF, args.Error(2)
}

func (m *MockFileService) WithClient(clientType, name, version string) service.FileService {
	args := m.Called(clientType, name, version)
	return args.Get(0).(service.FileService)
}

func (m *MockFileService) RecycleClear(ctx context.Context, uid int64, params *dto.FileRecycleClearRequest) error {
	args := m.Called(ctx, uid, params)
	return args.Error(0)
}

func (m *MockFileService) CleanDuplicateFiles(ctx context.Context, uid int64, vaultID int64) error {
	args := m.Called(ctx, uid, vaultID)
	return args.Error(0)
}
