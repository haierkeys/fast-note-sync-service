// Package mocks provides testify/mock implementations for service interfaces.
// Package mocks 提供服务接口的 testify/mock 实现。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/stretchr/testify/mock"
)

// MockShareService is a testify/mock implementation of service.ShareService.
// MockShareService 是 service.ShareService 的 testify/mock 实现。
type MockShareService struct {
	mock.Mock
}

// Ensure MockShareService implements service.ShareService at compile time.
// 编译期确保 MockShareService 实现了 service.ShareService 接口。
var _ service.ShareService = (*MockShareService)(nil)

func (m *MockShareService) ShareGenerate(ctx context.Context, uid int64, vaultName string, path string, pathHash string, password string) (*dto.ShareCreateResponse, error) {
	args := m.Called(ctx, uid, vaultName, path, pathHash, password)
	if v := args.Get(0); v != nil {
		return v.(*dto.ShareCreateResponse), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockShareService) VerifyShare(ctx context.Context, token string, rid string, rtp string, password string) (*pkgapp.ShareEntity, error) {
	args := m.Called(ctx, token, rid, rtp, password)
	if v := args.Get(0); v != nil {
		return v.(*pkgapp.ShareEntity), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockShareService) GetSharedNote(ctx context.Context, shareToken string, noteID int64, password string) (*dto.NoteDTO, error) {
	args := m.Called(ctx, shareToken, noteID, password)
	if v := args.Get(0); v != nil {
		return v.(*dto.NoteDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockShareService) GetSharedFile(ctx context.Context, shareToken string, fileID int64, password string) ([]byte, string, int64, string, string, error) {
	args := m.Called(ctx, shareToken, fileID, password)
	var content []byte
	if v := args.Get(0); v != nil {
		content = v.([]byte)
	}
	return content, args.String(1), args.Get(2).(int64), args.String(3), args.String(4), args.Error(5)
}

func (m *MockShareService) GetSharedFileInfo(ctx context.Context, shareToken string, fileID int64, password string) (string, string, int64, string, string, error) {
	args := m.Called(ctx, shareToken, fileID, password)
	return args.String(0), args.String(1), args.Get(2).(int64), args.String(3), args.String(4), args.Error(5)
}

func (m *MockShareService) RecordView(uid int64, id int64) {
	m.Called(uid, id)
}

func (m *MockShareService) StopShare(ctx context.Context, uid int64, id int64) error {
	args := m.Called(ctx, uid, id)
	return args.Error(0)
}

func (m *MockShareService) UpdateSharePassword(ctx context.Context, uid int64, vaultName string, path string, pathHash string, password string) error {
	args := m.Called(ctx, uid, vaultName, path, pathHash, password)
	return args.Error(0)
}

func (m *MockShareService) CreateShortLink(ctx context.Context, uid int64, vaultName string, path string, pathHash string, baseURL string, longURL string, isForce bool) (string, error) {
	args := m.Called(ctx, uid, vaultName, path, pathHash, baseURL, longURL, isForce)
	return args.String(0), args.Error(1)
}

func (m *MockShareService) ListShares(ctx context.Context, uid int64, sortBy string, sortOrder string, pager *pkgapp.Pager) ([]*dto.ShareListItem, int, error) {
	args := m.Called(ctx, uid, sortBy, sortOrder, pager)
	if v := args.Get(0); v != nil {
		return v.([]*dto.ShareListItem), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockShareService) GetShareByPath(ctx context.Context, uid int64, vaultName string, pathHash string) (*domain.UserShare, error) {
	args := m.Called(ctx, uid, vaultName, pathHash)
	if v := args.Get(0); v != nil {
		return v.(*domain.UserShare), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockShareService) StopShareByPath(ctx context.Context, uid int64, vaultName string, pathHash string) error {
	args := m.Called(ctx, uid, vaultName, pathHash)
	return args.Error(0)
}

func (m *MockShareService) GetActiveNotePathsByVault(ctx context.Context, uid int64, vaultName string) ([]string, error) {
	args := m.Called(ctx, uid, vaultName)
	if v := args.Get(0); v != nil {
		return v.([]string), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockShareService) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
