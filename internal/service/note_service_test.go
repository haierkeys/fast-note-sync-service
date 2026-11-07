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
	return service.NewNoteService(nil, noteRepo, nil, nil, nil, nil, historyRepo, vaultSvc, folderSvc, nil, nil, nil, nil)
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
// （与 share 迁移失败的处理保持一致）。
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
