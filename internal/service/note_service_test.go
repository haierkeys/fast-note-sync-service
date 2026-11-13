// Package service_test contains black-box tests for the service layer.
// Package service_test 服务层的黑盒测试
package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	domainmocks "github.com/haierkeys/fast-note-sync-service/internal/domain/mocks"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	servicemocks "github.com/haierkeys/fast-note-sync-service/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// 回归测试：重命名（Rename）必须迁移 note_history 表的历史归属
// Regression: Rename must migrate note_history ownership from old note ID to new note ID
//
// 背景：noteService.Rename 会新建笔记记录（新 ID）并软删旧记录（Rename=1），
// 所有挂在旧 ID 上的关联数据（snapshot/version/share）都已迁移，唯独 note_history
// 表被遗漏，导致重命名后历史版本丢失。本次修复在 noteService.Migrate 中补充
// historyRepo.Migrate 调用。以下测试防止该迁移再次被遗漏。
// ---------------------------------------------------------------------------

func oldNoteFixture() *domain.Note {
	return &domain.Note{
		ID:                      137,
		VaultID:                 1,
		Action:                  domain.NoteActionDelete,
		Rename:                  1,
		FID:                     10,
		Path:                    "旧笔记.md",
		PathHash:                "oldpathhash",
		Content:                 "当前内容",
		ContentHash:             "currenthash",
		ContentLastSnapshot:     "上次快照",
		ContentLastSnapshotHash: "snapshothash",
		Version:                 5,
		ClientName:              "obsidian",
		ClientType:              "desktop",
		ClientVersion:           "2.4.0",
		Size:                    1024,
		Ctime:                   1000,
		Mtime:                   2000,
		UpdatedTimestamp:        3000,
	}
}

// newNoteSvc 构造 NoteService（vaultSvc/folderSvc 可为 nil，Migrate 级测试不需要它们）
func newNoteSvc(t *testing.T, noteRepo *domainmocks.MockNoteRepository, historyRepo *domainmocks.MockNoteHistoryRepository, vaultSvc service.VaultService, folderSvc service.FolderService) service.NoteService {
	t.Helper()
	return service.NewNoteService(nil, noteRepo, nil, nil, nil, historyRepo, vaultSvc, folderSvc, nil, nil, nil, nil)
}

// TestNoteService_Migrate_MigratesNoteHistory 验证 Migrate 会把旧笔记 ID 的历史迁移到新笔记 ID。
// 本次 bug 的直接回归测试：修复前 Migrate 从不调用 historyRepo.Migrate。
func TestNoteService_Migrate_MigratesNoteHistory(t *testing.T) {
	noteRepo := new(domainmocks.MockNoteRepository)
	historyRepo := new(domainmocks.MockNoteHistoryRepository)

	oldNote := oldNoteFixture()

	noteRepo.On("GetByID", mock.Anything, int64(137), int64(1)).Return(oldNote, nil)
	noteRepo.On("UpdateSnapshot", mock.Anything, oldNote.ContentLastSnapshot, oldNote.ContentLastSnapshotHash, oldNote.Version, int64(2048), int64(1)).Return(nil)
	noteRepo.On("UpdateDelete", mock.Anything, mock.AnythingOfType("*domain.Note"), int64(1)).Return(nil)
	// CountSizeSum 有 10 秒防抖 timer，测试窗口内不会触发，标记 Maybe
	noteRepo.On("CountSizeSum", mock.Anything, int64(1), int64(1)).Return(&domain.CountSizeResult{Count: 1, Size: 1024}, nil).Maybe()
	// 关键断言：历史迁移必须发生
	historyRepo.On("Migrate", mock.Anything, int64(137), int64(2048), int64(1)).Return(nil)

	svc := newNoteSvc(t, noteRepo, historyRepo, nil, nil)

	err := svc.Migrate(context.Background(), 137, 2048, 1)
	require.NoError(t, err)

	historyRepo.AssertCalled(t, "Migrate", mock.Anything, int64(137), int64(2048), int64(1))
	noteRepo.AssertCalled(t, "UpdateSnapshot", mock.Anything, oldNote.ContentLastSnapshot, oldNote.ContentLastSnapshotHash, oldNote.Version, int64(2048), int64(1))
	noteRepo.AssertCalled(t, "UpdateDelete", mock.Anything, mock.AnythingOfType("*domain.Note"), int64(1))
	historyRepo.AssertExpectations(t)
	noteRepo.AssertExpectations(t)
}

// TestNoteService_Migrate_HistoryErrorWarnNotFail 验证历史迁移失败只告警、不阻断整个迁移流程
//（与 share 迁移失败的处理保持一致）。
func TestNoteService_Migrate_HistoryErrorWarnNotFail(t *testing.T) {
	noteRepo := new(domainmocks.MockNoteRepository)
	historyRepo := new(domainmocks.MockNoteHistoryRepository)

	oldNote := oldNoteFixture()

	noteRepo.On("GetByID", mock.Anything, int64(137), int64(1)).Return(oldNote, nil)
	noteRepo.On("UpdateSnapshot", mock.Anything, oldNote.ContentLastSnapshot, oldNote.ContentLastSnapshotHash, oldNote.Version, int64(2048), int64(1)).Return(nil)
	noteRepo.On("UpdateDelete", mock.Anything, mock.AnythingOfType("*domain.Note"), int64(1)).Return(nil)
	noteRepo.On("CountSizeSum", mock.Anything, int64(1), int64(1)).Return(&domain.CountSizeResult{Count: 1, Size: 1024}, nil)
	historyRepo.On("Migrate", mock.Anything, int64(137), int64(2048), int64(1)).Return(assert.AnError)

	svc := newNoteSvc(t, noteRepo, historyRepo, nil, nil)

	err := svc.Migrate(context.Background(), 137, 2048, 1)
	require.NoError(t, err)
	historyRepo.AssertCalled(t, "Migrate", mock.Anything, int64(137), int64(2048), int64(1))
}

// TestNoteService_Rename_MigratesNoteHistory 端到端回归：Rename 完成后，
// 后台 Migrate 必须把旧笔记 ID 的历史迁移到新笔记 ID（异步，用 Eventually 等待）。
func TestNoteService_Rename_MigratesNoteHistory(t *testing.T) {
	noteRepo := new(domainmocks.MockNoteRepository)
	historyRepo := new(domainmocks.MockNoteHistoryRepository)
	vaultSvc := new(servicemocks.MockVaultService)
	folderSvc := new(servicemocks.MockFolderService)

	oldNote := oldNoteFixture()
	newNote := &domain.Note{
		ID:          2048,
		VaultID:     1,
		Action:      domain.NoteActionCreate,
		Path:        "新笔记.md",
		PathHash:    "newpathhash",
		FID:         10,
		Content:     oldNote.Content,
		ContentHash: oldNote.ContentHash,
		Version:     oldNote.Version,
		Ctime:       oldNote.Ctime,
		Mtime:       oldNote.Mtime,
	}

	// Rename 主流程依赖
	vaultSvc.On("MustGetID", mock.Anything, int64(1), "obsidian-vault").Return(int64(1), nil)
	noteRepo.On("GetAllByPathHash", mock.Anything, "newpathhash", int64(1), int64(1)).Return(nil, nil)
	noteRepo.On("GetByPathHash", mock.Anything, "oldpathhash", int64(1), int64(1)).Return(oldNote, nil)
	noteRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Note"), int64(1)).Return(oldNote, nil)
	folderSvc.On("EnsurePathFID", mock.Anything, int64(1), int64(1), "").Return(int64(10), nil)
	noteRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Note"), int64(1)).Return(newNote, nil)
	folderSvc.On("SyncResourceFID", mock.Anything, int64(1), int64(1), mock.Anything, mock.Anything).Return(nil)
	folderSvc.On("CleanupEmptyAncestors", mock.Anything, int64(1), int64(1), "旧笔记.md").Return(nil)

	// Migrate goroutine 依赖
	noteRepo.On("GetByID", mock.Anything, int64(137), int64(1)).Return(oldNote, nil)
	noteRepo.On("UpdateSnapshot", mock.Anything, oldNote.ContentLastSnapshot, oldNote.ContentLastSnapshotHash, oldNote.Version, int64(2048), int64(1)).Return(nil)
	noteRepo.On("UpdateDelete", mock.Anything, mock.AnythingOfType("*domain.Note"), int64(1)).Return(nil)
	noteRepo.On("CountSizeSum", mock.Anything, int64(1), int64(1)).Return(&domain.CountSizeResult{Count: 1, Size: 1024}, nil).Maybe()
	historyRepo.On("Migrate", mock.Anything, int64(137), int64(2048), int64(1)).Return(nil)

	svc := newNoteSvc(t, noteRepo, historyRepo, vaultSvc, folderSvc)

	_, _, err := svc.Rename(context.Background(), 1, &dto.NoteRenameRequest{
		Vault:       "obsidian-vault",
		Path:        "新笔记.md",
		PathHash:    "newpathhash",
		OldPath:     "旧笔记.md",
		OldPathHash: "oldpathhash",
	})
	require.NoError(t, err)

	// Rename 内部 go s.Migrate(...) 异步执行，等待 history 迁移被触发
	deadline := time.Now().Add(2 * time.Second)
	migrated := false
	for time.Now().Before(deadline) {
		if len(historyRepo.Calls) > 0 {
			migrated = true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	require.True(t, migrated, "Rename 后应触发 note_history 迁移（oldNoteID 137 -> newNoteID 2048）")
	historyRepo.AssertCalled(t, "Migrate", mock.Anything, int64(137), int64(2048), int64(1))

	// 确认 Rename 主流程 stub 全部命中
	noteRepo.AssertExpectations(t)
	vaultSvc.AssertExpectations(t)
	folderSvc.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// 回归测试：Rename 必须把 Size（以及复用记录时的 Ctime）带到新笔记记录上
// Regression: Rename must carry Size (and Ctime when reusing a record) over to the new note record
//
// 背景：修复前，新建分支创建的记录 Size 为 0；复用已删除记录的分支保留了旧记录的 Size 和 Ctime。
// 内容本身完好，但 note_list、GET /api/notes 以及同步消息都报告 0 字节（或过期的大小）。
// Before the fix the create branch stored Size 0, and the reuse branch kept the stale Size/Ctime
// of the deleted record. Content was intact, but note_list, GET /api/notes and sync messages
// reported 0 bytes (or an outdated size).
// ---------------------------------------------------------------------------

// renameMigrateMocks 注册 Rename 内部异步 goroutine（Migrate、SyncResourceFID）可能用到的 stub
// renameMigrateMocks registers stubs for the async goroutines started by Rename
func renameMigrateMocks(noteRepo *domainmocks.MockNoteRepository, historyRepo *domainmocks.MockNoteHistoryRepository, folderSvc *servicemocks.MockFolderService) {
	noteRepo.On("GetByID", mock.Anything, mock.Anything, int64(1)).Return(oldNoteFixture(), nil).Maybe()
	noteRepo.On("UpdateSnapshot", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, int64(1)).Return(nil).Maybe()
	noteRepo.On("UpdateDelete", mock.Anything, mock.AnythingOfType("*domain.Note"), int64(1)).Return(nil).Maybe()
	noteRepo.On("CountSizeSum", mock.Anything, int64(1), int64(1)).Return(&domain.CountSizeResult{Count: 1, Size: 1024}, nil).Maybe()
	historyRepo.On("Migrate", mock.Anything, mock.Anything, mock.Anything, int64(1)).Return(nil).Maybe()
	folderSvc.On("SyncResourceFID", mock.Anything, int64(1), int64(1), mock.Anything, mock.Anything).Return(nil).Maybe()
}

// TestNoteService_Rename_NewRecordKeepsSize 新建分支：新记录必须带上旧笔记的 Size
// Create branch: the new record must carry the old note's Size
func TestNoteService_Rename_NewRecordKeepsSize(t *testing.T) {
	noteRepo := new(domainmocks.MockNoteRepository)
	historyRepo := new(domainmocks.MockNoteHistoryRepository)
	vaultSvc := new(servicemocks.MockVaultService)
	folderSvc := new(servicemocks.MockFolderService)

	oldNote := oldNoteFixture()
	var created *domain.Note

	vaultSvc.On("MustGetID", mock.Anything, int64(1), "obsidian-vault").Return(int64(1), nil)
	noteRepo.On("GetAllByPathHash", mock.Anything, "newpathhash", int64(1), int64(1)).Return(nil, nil)
	noteRepo.On("GetByPathHash", mock.Anything, "oldpathhash", int64(1), int64(1)).Return(oldNote, nil)
	noteRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Note"), int64(1)).Return(oldNote, nil)
	folderSvc.On("EnsurePathFID", mock.Anything, int64(1), int64(1), "目录").Return(int64(11), nil)
	noteRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Note"), int64(1)).
		Run(func(args mock.Arguments) {
			n := *args.Get(1).(*domain.Note)
			created = &n
		}).
		Return(&domain.Note{ID: 2048, VaultID: 1, Path: "目录/新笔记.md", PathHash: "newpathhash"}, nil)
	folderSvc.On("CleanupEmptyAncestors", mock.Anything, int64(1), int64(1), "旧笔记.md").Return(nil)
	renameMigrateMocks(noteRepo, historyRepo, folderSvc)

	svc := newNoteSvc(t, noteRepo, historyRepo, vaultSvc, folderSvc)

	_, _, err := svc.Rename(context.Background(), 1, &dto.NoteRenameRequest{
		Vault:       "obsidian-vault",
		Path:        "目录/新笔记.md",
		PathHash:    "newpathhash",
		OldPath:     "旧笔记.md",
		OldPathHash: "oldpathhash",
	})
	require.NoError(t, err)
	require.NotNil(t, created, "Rename 应创建新记录")
	assert.Equal(t, int64(1024), created.Size, "新记录必须保留旧笔记的 Size")
	assert.Equal(t, int64(1000), created.Ctime)
	assert.Equal(t, "当前内容", created.Content)
}

// TestNoteService_Rename_ReusedRecordTakesSizeAndCtime 复用分支：复用已删除记录时必须覆盖 Size 和 Ctime
// Reuse branch: when a deleted record is reused, Size and Ctime must be overwritten
func TestNoteService_Rename_ReusedRecordTakesSizeAndCtime(t *testing.T) {
	noteRepo := new(domainmocks.MockNoteRepository)
	historyRepo := new(domainmocks.MockNoteHistoryRepository)
	vaultSvc := new(servicemocks.MockVaultService)
	folderSvc := new(servicemocks.MockFolderService)

	oldNote := oldNoteFixture()
	deletedAtTarget := &domain.Note{
		ID:       4096,
		VaultID:  1,
		Action:   domain.NoteActionDelete,
		Rename:   1,
		Path:     "新笔记.md",
		PathHash: "newpathhash",
		Content:  "很久以前的内容",
		Size:     7,
		Ctime:    9,
	}
	var updated []domain.Note

	vaultSvc.On("MustGetID", mock.Anything, int64(1), "obsidian-vault").Return(int64(1), nil)
	noteRepo.On("GetAllByPathHash", mock.Anything, "newpathhash", int64(1), int64(1)).Return(deletedAtTarget, nil)
	noteRepo.On("GetByPathHash", mock.Anything, "oldpathhash", int64(1), int64(1)).Return(oldNote, nil)
	noteRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Note"), int64(1)).
		Run(func(args mock.Arguments) {
			updated = append(updated, *args.Get(1).(*domain.Note))
		}).
		Return(&domain.Note{ID: 4096, VaultID: 1, Path: "新笔记.md", PathHash: "newpathhash"}, nil)
	folderSvc.On("EnsurePathFID", mock.Anything, int64(1), int64(1), "").Return(int64(10), nil)
	folderSvc.On("CleanupEmptyAncestors", mock.Anything, int64(1), int64(1), "旧笔记.md").Return(nil)
	renameMigrateMocks(noteRepo, historyRepo, folderSvc)

	svc := newNoteSvc(t, noteRepo, historyRepo, vaultSvc, folderSvc)

	_, _, err := svc.Rename(context.Background(), 1, &dto.NoteRenameRequest{
		Vault:       "obsidian-vault",
		Path:        "新笔记.md",
		PathHash:    "newpathhash",
		OldPath:     "旧笔记.md",
		OldPathHash: "oldpathhash",
	})
	require.NoError(t, err)
	noteRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)

	var reused *domain.Note
	for i := range updated {
		if updated[i].ID == 4096 {
			reused = &updated[i]
		}
	}
	require.NotNil(t, reused, "Rename 应复用目标路径上已删除的记录")
	assert.Equal(t, domain.NoteActionCreate, reused.Action)
	assert.Equal(t, int64(1024), reused.Size, "复用记录必须取旧笔记的 Size，而不是已删除记录的旧值")
	assert.Equal(t, int64(1000), reused.Ctime, "复用记录必须取旧笔记的 Ctime")
	assert.Equal(t, "当前内容", reused.Content)
}
