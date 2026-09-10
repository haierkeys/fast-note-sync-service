// Package service_test contains black-box tests for the service layer.
// Package service_test 服务层的黑盒测试
package service_test

import (
	"context"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	domainmocks "github.com/haierkeys/fast-note-sync-service/internal/domain/mocks"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// 回归测试：CleanDuplicateNotes 不得读取笔记内容文件
//
// 背景：SyncFIDTask 调用 CleanDuplicateNotes，后者用 ListByUpdatedTimestamp
// 加载全部笔记（逐条 toDomain → fillNoteContent 读 content.txt）。而
// noteRepository.Create 是"先插数据库记录、后写 content.txt"，两步骤之间
// 的窗口被并发读取命中时 fillNoteContent 报 "note content file not found"
// → 任务中断。
//
// 去重只需 path/action/timestamp 等元数据，改用 ListByUpdatedTimestampMeta
// （不读 content.txt/snapshot.txt）即可消除对内容文件的依赖。
// ---------------------------------------------------------------------------

// TestNoteService_CleanDuplicateNotes_UsesMetaListWithoutContentRead 验证去重扫描
// 走元数据路径（不触发内容文件读取）。
func TestNoteService_CleanDuplicateNotes_UsesMetaListWithoutContentRead(t *testing.T) {
	noteRepo := new(domainmocks.MockNoteRepository)

	// 修复后路径：元数据列表（不读文件）
	noteRepo.On("ListByUpdatedTimestampMeta", mock.Anything, int64(0), int64(1), int64(1)).
		Return([]*domain.Note{
			{ID: 1, Path: "a.md", PathHash: "h1", Action: domain.NoteActionCreate},
		}, nil)
	// 修复前路径：读内容文件（生产环境文件缺失即报错）
	noteRepo.On("ListByUpdatedTimestamp", mock.Anything, int64(0), int64(1), int64(1)).
		Return([]*domain.Note{}, nil).Maybe()

	svc := service.NewNoteService(nil, noteRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	err := svc.CleanDuplicateNotes(context.Background(), 1, 1)

	require.NoError(t, err, "Bug: 去重扫描不应因内容文件缺失而失败")
	noteRepo.AssertNotCalled(t, "ListByUpdatedTimestamp",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
