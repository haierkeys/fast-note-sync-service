package service

import (
	"context"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	domainmocks "github.com/haierkeys/fast-note-sync-service/internal/domain/mocks"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// 回归测试：服务端「墓碑闸门」——陈旧重报不得复活已删资源
// Capture tests: stale uploads must not revive server-deleted resources
//
// 背景：客户端本地同步状态（folderSnapshot / hashManager）失效后会周期性
// 全量重报整个 vault；服务端此前对 action=delete 的记录无条件复活（笔记、
// 附件、目录三处），形成"删了又被顶回来"的复活拉锯。
// 修复：以墓碑记录的 updated_timestamp 为删除时刻，上传的 mtime 与 ctime
// 均早于「删除时刻 - 容差」时判定为陈旧重报 → 拒绝复活。
// 修复前，本文件中期望"拒绝"的用例必然 FAIL（复活发生）；修复后 PASS。
// ---------------------------------------------------------------------------

const testDeletedAt = int64(1_700_000_000_000)

// stubVaultService 仅覆盖测试所需的 VaultService 子集（其余方法不会被调用）。
type stubVaultService struct {
	VaultService
}

func (stubVaultService) MustGetID(ctx context.Context, uid int64, name string) (int64, error) {
	return 1, nil
}

func (stubVaultService) UpdateNoteStats(ctx context.Context, noteSize, noteCount, vaultID, uid int64) error {
	return nil
}

// stubFolderService 仅覆盖 SyncResourceFID（成功路径会异步调用它）。
type stubFolderService struct {
	FolderService
}

func (stubFolderService) SyncResourceFID(ctx context.Context, uid int64, vaultID int64, noteIDs []int64, fileIDs []int64) error {
	return nil
}

// ---------------------------------------------------------------------------
// Section A：isStaleRevive 判定表
// ---------------------------------------------------------------------------

func TestIsStaleRevive_Table(t *testing.T) {
	tests := []struct {
		name      string
		mtime     int64
		ctime     int64
		deletedAt int64
		want      bool
	}{
		{"both timestamps long before deletion -> stale", testDeletedAt - 10_000_000, testDeletedAt - 10_000_000, testDeletedAt, true},
		{"mtime after deletion (genuine edit) -> fresh", testDeletedAt + 5_000, testDeletedAt - 10_000_000, testDeletedAt, false},
		{"ctime after deletion (recycle restore) -> fresh", testDeletedAt - 10_000_000, testDeletedAt + 5_000, testDeletedAt, false},
		{"mtime at tolerance edge counts as fresh", testDeletedAt - staleReviveToleranceMs, testDeletedAt - 10_000_000, testDeletedAt, false},
		{"mtime one ms past tolerance -> stale", testDeletedAt - staleReviveToleranceMs - 1, testDeletedAt - 10_000_000, testDeletedAt, true},
		{"both unknown (zero) -> stale", 0, 0, testDeletedAt, true},
		{"ctime unknown, mtime old -> stale", testDeletedAt - 10_000_000, 0, testDeletedAt, true},
		{"ctime unknown, mtime fresh -> fresh", testDeletedAt + 1, 0, testDeletedAt, false},
		{"no tombstone timestamp -> never stale", 0, 0, 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, isStaleRevive(tc.mtime, tc.ctime, tc.deletedAt),
				"Bug: isStaleRevive(mtime=%d, ctime=%d, deletedAt=%d) 判定错误", tc.mtime, tc.ctime, tc.deletedAt)
		})
	}
}

// ---------------------------------------------------------------------------
// Section B：笔记侧墓碑闸门
// ---------------------------------------------------------------------------

func deletedNoteFixture() *domain.Note {
	return &domain.Note{
		ID:               7,
		VaultID:          1,
		Action:           domain.NoteActionDelete,
		Path:             "skills/old.md",
		PathHash:         "h1",
		Content:          "old content",
		ContentHash:      "oldhash",
		Size:             11,
		Ctime:            100,
		Mtime:            100,
		UpdatedTimestamp: testDeletedAt,
	}
}

// TestNoteService_ModifyOrCreate_StaleReviveRejected：双时间戳均早于删除时刻的
// 上传必须被拒绝且不写库（修复前该上传会把 delete 记录无条件翻回 create）。
func TestNoteService_ModifyOrCreate_StaleReviveRejected(t *testing.T) {
	noteRepo := new(domainmocks.MockNoteRepository)
	noteRepo.On("GetAllByPathHash", mock.Anything, "h1", int64(1), int64(1)).Return(deletedNoteFixture(), nil)

	svc := NewNoteService(nil, noteRepo, nil, nil, nil, nil, stubVaultService{}, stubFolderService{}, nil, nil, nil, nil)

	_, got, err := svc.ModifyOrCreate(context.Background(), 1, &dto.NoteModifyOrCreateRequest{
		Vault:       "vault",
		Path:        "skills/old.md",
		PathHash:    "h1",
		Content:     "stale re-reported content",
		ContentHash: "stalehash",
		Mtime:       200,
		Ctime:       200,
	}, true)

	require.ErrorIs(t, err, code.ErrorStaleReviveRejected,
		"Bug: mtime/ctime 均早于删除时刻的上传必须被墓碑闸门拒绝（修复前会被无条件复活）")
	require.Nil(t, got)
	noteRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
	noteRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
}

// TestNoteService_ModifyOrCreate_FreshMtimeRevivesDeletedNote：mtime 晚于删除时刻
// （用户在删除后真实编辑/重建）= 新意图，必须照常复活。
func TestNoteService_ModifyOrCreate_FreshMtimeRevivesDeletedNote(t *testing.T) {
	noteRepo := new(domainmocks.MockNoteRepository)
	noteRepo.On("GetAllByPathHash", mock.Anything, "h1", int64(1), int64(1)).Return(deletedNoteFixture(), nil)

	updated := deletedNoteFixture()
	updated.Action = domain.NoteActionCreate
	updated.Mtime = testDeletedAt + 60_000
	updated.UpdatedTimestamp = testDeletedAt + 60_000
	noteRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Note"), int64(1)).Return(updated, nil)
	noteRepo.On("CountSizeSum", mock.Anything, int64(1), int64(1)).Return(&domain.CountSizeResult{Count: 1, Size: 11}, nil).Maybe()

	svc := NewNoteService(nil, noteRepo, nil, nil, nil, nil, stubVaultService{}, stubFolderService{}, nil, nil, nil, nil)

	isNew, got, err := svc.ModifyOrCreate(context.Background(), 1, &dto.NoteModifyOrCreateRequest{
		Vault:       "vault",
		Path:        "skills/old.md",
		PathHash:    "h1",
		Content:     "genuinely new content",
		ContentHash: "newhash",
		Mtime:       testDeletedAt + 60_000,
		Ctime:       100,
	}, true)

	require.NoError(t, err)
	require.False(t, isNew)
	require.NotNil(t, got)
	noteRepo.AssertCalled(t, "Update", mock.Anything, mock.AnythingOfType("*domain.Note"), int64(1))
}

// TestNoteService_ModifyOrCreate_FreshCtimeRevivesDeletedNote：回收站恢复场景，
// mtime 保留旧值但 ctime（rename 会刷新）晚于删除时刻 → 属于真实恢复，放行。
func TestNoteService_ModifyOrCreate_FreshCtimeRevivesDeletedNote(t *testing.T) {
	noteRepo := new(domainmocks.MockNoteRepository)
	noteRepo.On("GetAllByPathHash", mock.Anything, "h1", int64(1), int64(1)).Return(deletedNoteFixture(), nil)

	updated := deletedNoteFixture()
	updated.Action = domain.NoteActionCreate
	updated.UpdatedTimestamp = testDeletedAt + 60_000
	noteRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Note"), int64(1)).Return(updated, nil)
	noteRepo.On("CountSizeSum", mock.Anything, int64(1), int64(1)).Return(&domain.CountSizeResult{Count: 1, Size: 11}, nil).Maybe()

	svc := NewNoteService(nil, noteRepo, nil, nil, nil, nil, stubVaultService{}, stubFolderService{}, nil, nil, nil, nil)

	_, got, err := svc.ModifyOrCreate(context.Background(), 1, &dto.NoteModifyOrCreateRequest{
		Vault:       "vault",
		Path:        "skills/old.md",
		PathHash:    "h1",
		Content:     "restored content",
		ContentHash: "restoredhash",
		Mtime:       100, // 保留旧值
		Ctime:       testDeletedAt + 60_000,
	}, true)

	require.NoError(t, err)
	require.NotNil(t, got)
	noteRepo.AssertCalled(t, "Update", mock.Anything, mock.AnythingOfType("*domain.Note"), int64(1))
}

// ---------------------------------------------------------------------------
// Section C：附件侧墓碑闸门
// ---------------------------------------------------------------------------

func deletedFileFixture() *domain.File {
	return &domain.File{
		ID:               9,
		VaultID:          1,
		Action:           domain.FileActionDelete,
		Path:             "skills/img.png",
		PathHash:         "fh1",
		ContentHash:      "oldfh",
		Size:             5,
		Ctime:            100,
		Mtime:            100,
		UpdatedTimestamp: testDeletedAt,
	}
}

func TestFileService_UpdateOrCreate_StaleReviveRejected(t *testing.T) {
	fileRepo := new(domainmocks.MockFileRepository)
	fileRepo.On("GetByPathHash", mock.Anything, "fh1", int64(1), int64(1)).Return(deletedFileFixture(), nil)

	svc := NewFileService(nil, fileRepo, nil, stubVaultService{}, stubFolderService{}, nil, nil, nil, nil)

	_, got, err := svc.UpdateOrCreate(context.Background(), 1, &dto.FileUpdateRequest{
		Vault:       "vault",
		Path:        "skills/img.png",
		PathHash:    "fh1",
		ContentHash: "stalehash",
		SavePath:    "/tmp/test-not-written",
		Size:        5,
		Ctime:       200,
		Mtime:       200,
	}, true)

	require.ErrorIs(t, err, code.ErrorStaleReviveRejected,
		"Bug: 附件陈旧重报同样必须被墓碑闸门拒绝")
	require.Nil(t, got)
	fileRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
	fileRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
}

func TestFileService_UpdateOrCreate_FreshEditRevivesDeletedFile(t *testing.T) {
	fileRepo := new(domainmocks.MockFileRepository)
	fileRepo.On("GetByPathHash", mock.Anything, "fh1", int64(1), int64(1)).Return(deletedFileFixture(), nil)

	updated := deletedFileFixture()
	updated.Action = domain.FileActionCreate
	updated.Mtime = testDeletedAt + 60_000
	updated.UpdatedTimestamp = testDeletedAt + 60_000
	fileRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.File"), int64(1)).Return(updated, nil)
	fileRepo.On("CountSizeSum", mock.Anything, int64(1), int64(1)).Return(&domain.CountSizeResult{Count: 1, Size: 5}, nil).Maybe()

	svc := NewFileService(nil, fileRepo, nil, stubVaultService{}, stubFolderService{}, nil, nil, nil, nil)

	_, got, err := svc.UpdateOrCreate(context.Background(), 1, &dto.FileUpdateRequest{
		Vault:       "vault",
		Path:        "skills/img.png",
		PathHash:    "fh1",
		ContentHash: "newhash",
		SavePath:    "/tmp/test-not-written",
		Size:        5,
		Ctime:       100,
		Mtime:       testDeletedAt + 60_000,
	}, true)

	require.NoError(t, err)
	require.NotNil(t, got)
	fileRepo.AssertCalled(t, "Update", mock.Anything, mock.AnythingOfType("*domain.File"), int64(1))
}

// ---------------------------------------------------------------------------
// Section D：目录侧墓碑判定（doFolderSync 用它拦截旧上报复活）
// ---------------------------------------------------------------------------

func TestFolderService_IsDeleted(t *testing.T) {
	ctx := context.Background()

	t.Run("tombstone record reports deleted", func(t *testing.T) {
		repo := new(domainmocks.MockFolderRepository)
		repo.On("GetByPathHash", mock.Anything, "fh1", int64(1), int64(1)).
			Return(&domain.Folder{ID: 3, Action: domain.FolderActionDelete}, nil)

		svc := NewFolderService(repo, nil, nil, stubVaultService{}, nil, nil, nil, nil)

		deleted, err := svc.IsDeleted(ctx, 1, "vault", "fh1")
		require.NoError(t, err)
		require.True(t, deleted, "Bug: 已删除目录（墓碑）必须报告 deleted=true，供 doFolderSync 拦截陈旧上报")
	})

	t.Run("active record reports alive", func(t *testing.T) {
		repo := new(domainmocks.MockFolderRepository)
		repo.On("GetByPathHash", mock.Anything, "fh1", int64(1), int64(1)).
			Return(&domain.Folder{ID: 3, Action: domain.FolderActionCreate}, nil)

		svc := NewFolderService(repo, nil, nil, stubVaultService{}, nil, nil, nil, nil)

		deleted, err := svc.IsDeleted(ctx, 1, "vault", "fh1")
		require.NoError(t, err)
		require.False(t, deleted)
	})

	t.Run("missing record reports alive", func(t *testing.T) {
		repo := new(domainmocks.MockFolderRepository)
		repo.On("GetByPathHash", mock.Anything, "fh1", int64(1), int64(1)).
			Return(nil, gorm.ErrRecordNotFound)

		svc := NewFolderService(repo, nil, nil, stubVaultService{}, nil, nil, nil, nil)

		deleted, err := svc.IsDeleted(ctx, 1, "vault", "fh1")
		require.NoError(t, err)
		require.False(t, deleted)
	})
}
