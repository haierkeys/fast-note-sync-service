package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/stretchr/testify/mock"
)

// NoteServiceMock is a mock implementation of service.NoteService
type NoteServiceMock struct {
	mock.Mock
}

func (m *NoteServiceMock) Get(ctx context.Context, uid int64, params *dto.NoteGetRequest) (*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.NoteDTO), args.Error(1)
}

func (m *NoteServiceMock) UpdateCheck(ctx context.Context, uid int64, params *dto.NoteUpdateCheckRequest) (string, *dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if args.Get(1) == nil {
		return args.String(0), nil, args.Error(2)
	}
	return args.String(0), args.Get(1).(*dto.NoteDTO), args.Error(2)
}

func (m *NoteServiceMock) ModifyOrCreate(ctx context.Context, uid int64, params *dto.NoteModifyOrCreateRequest, mtimeCheck bool) (bool, *dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params, mtimeCheck)
	if args.Get(1) == nil {
		return args.Bool(0), nil, args.Error(2)
	}
	return args.Bool(0), args.Get(1).(*dto.NoteDTO), args.Error(2)
}

func (m *NoteServiceMock) Delete(ctx context.Context, uid int64, params *dto.NoteDeleteRequest) (*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.NoteDTO), args.Error(1)
}

func (m *NoteServiceMock) Restore(ctx context.Context, uid int64, params *dto.NoteRestoreRequest) (*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.NoteDTO), args.Error(1)
}

func (m *NoteServiceMock) Rename(ctx context.Context, uid int64, params *dto.NoteRenameRequest) (*dto.NoteDTO, *dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	var d1, d2 *dto.NoteDTO
	if args.Get(0) != nil {
		d1 = args.Get(0).(*dto.NoteDTO)
	}
	if args.Get(1) != nil {
		d2 = args.Get(1).(*dto.NoteDTO)
	}
	return d1, d2, args.Error(2)
}

func (m *NoteServiceMock) List(ctx context.Context, uid int64, params *dto.NoteListRequest, pager *pkgapp.Pager) ([]*dto.NoteNoContentDTO, int, error) {
	args := m.Called(ctx, uid, params, pager)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*dto.NoteNoContentDTO), args.Int(1), args.Error(2)
}

func (m *NoteServiceMock) ListByLastTime(ctx context.Context, uid int64, params *dto.NoteSyncRequest) ([]*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.NoteDTO), args.Error(1)
}

func (m *NoteServiceMock) Sync(ctx context.Context, uid int64, params *dto.NoteSyncRequest) ([]*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.NoteDTO), args.Error(1)
}

func (m *NoteServiceMock) CountSizeSum(ctx context.Context, vaultID int64, uid int64) error {
	args := m.Called(ctx, vaultID, uid)
	return args.Error(0)
}

func (m *NoteServiceMock) Cleanup(ctx context.Context, uid int64) error {
	args := m.Called(ctx, uid)
	return args.Error(0)
}

func (m *NoteServiceMock) CleanupByTime(ctx context.Context, cutoffTime int64) error {
	args := m.Called(ctx, cutoffTime)
	return args.Error(0)
}

func (m *NoteServiceMock) ListNeedSnapshot(ctx context.Context, uid int64) ([]*dto.NoteDTO, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.NoteDTO), args.Error(1)
}

func (m *NoteServiceMock) Migrate(ctx context.Context, oldNoteID, newNoteID int64, uid int64) error {
	args := m.Called(ctx, oldNoteID, newNoteID, uid)
	return args.Error(0)
}

func (m *NoteServiceMock) MigratePush(oldNoteID, newNoteID int64, uid int64) {
	m.Called(oldNoteID, newNoteID, uid)
}

func (m *NoteServiceMock) WithClient(name, version string) service.NoteService {
	args := m.Called(name, version)
	return args.Get(0).(service.NoteService)
}

func (m *NoteServiceMock) PatchFrontmatter(ctx context.Context, uid int64, params *dto.NotePatchFrontmatterRequest) (*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.NoteDTO), args.Error(1)
}

func (m *NoteServiceMock) AppendContent(ctx context.Context, uid int64, params *dto.NoteAppendRequest) (*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.NoteDTO), args.Error(1)
}

func (m *NoteServiceMock) PrependContent(ctx context.Context, uid int64, params *dto.NotePrependRequest) (*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.NoteDTO), args.Error(1)
}

func (m *NoteServiceMock) ReplaceContent(ctx context.Context, uid int64, params *dto.NoteReplaceRequest) (*dto.NoteReplaceResponse, error) {
	args := m.Called(ctx, uid, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.NoteReplaceResponse), args.Error(1)
}

func (m *NoteServiceMock) Move(ctx context.Context, uid int64, params *dto.NoteMoveRequest) (*dto.NoteDTO, error) {
	args := m.Called(ctx, uid, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.NoteDTO), args.Error(1)
}

func (m *NoteServiceMock) UpdateNoteLinks(ctx context.Context, noteID int64, content string, vaultID, uid int64) {
	m.Called(ctx, noteID, content, vaultID, uid)
}

func (m *NoteServiceMock) RecycleClear(ctx context.Context, uid int64, params *dto.NoteRecycleClearRequest) error {
	args := m.Called(ctx, uid, params)
	return args.Error(0)
}

// VaultServiceMock is a mock implementation of service.VaultService
type VaultServiceMock struct {
	mock.Mock
}

func (m *VaultServiceMock) Get(ctx context.Context, uid int64, params *dto.VaultGetRequest) (*dto.VaultDTO, error) {
	args := m.Called(ctx, uid, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VaultDTO), args.Error(1)
}

func (m *VaultServiceMock) CreateOrUpdate(ctx context.Context, uid int64, params *dto.VaultCreateRequest) (*dto.VaultDTO, error) {
	args := m.Called(ctx, uid, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VaultDTO), args.Error(1)
}

func (m *VaultServiceMock) List(ctx context.Context, uid int64) ([]*dto.VaultDTO, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.VaultDTO), args.Error(1)
}

func (m *VaultServiceMock) Delete(ctx context.Context, uid int64, params *dto.VaultDeleteRequest) error {
	args := m.Called(ctx, uid, params)
	return args.Error(0)
}

func (m *VaultServiceMock) MustGetID(ctx context.Context, uid int64, name string) (int64, error) {
	args := m.Called(ctx, uid, name)
	return args.Get(0).(int64), args.Error(1)
}

func (m *VaultServiceMock) UpdateNoteStats(ctx context.Context, noteSize, noteCount, vaultID, uid int64) error {
	args := m.Called(ctx, noteSize, noteCount, vaultID, uid)
	return args.Error(0)
}

func (m *VaultServiceMock) UpdateFileStats(ctx context.Context, fileSize, fileCount, vaultID, uid int64) error {
	args := m.Called(ctx, fileSize, fileCount, vaultID, uid)
	return args.Error(0)
}

// StorageServiceMock is a mock implementation of service.StorageService
type StorageServiceMock struct {
	mock.Mock
}

func (m *StorageServiceMock) Get(ctx context.Context, uid int64, id int64) (*dto.StorageDTO, error) {
	args := m.Called(ctx, uid, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.StorageDTO), args.Error(1)
}

func (m *StorageServiceMock) CreateOrUpdate(ctx context.Context, uid int64, req *dto.StorageRequest) (*dto.StorageDTO, error) {
	args := m.Called(ctx, uid, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.StorageDTO), args.Error(1)
}

func (m *StorageServiceMock) List(ctx context.Context, uid int64) ([]*dto.StorageDTO, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.StorageDTO), args.Error(1)
}

func (m *StorageServiceMock) Delete(ctx context.Context, uid int64, id int64) error {
	args := m.Called(ctx, uid, id)
	return args.Error(0)
}

func (m *StorageServiceMock) EnabledTypes(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *StorageServiceMock) Validate(ctx context.Context, uid int64, req *dto.StorageRequest) error {
	args := m.Called(ctx, uid, req)
	return args.Error(0)
}
