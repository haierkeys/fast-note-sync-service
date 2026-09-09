// Package service_test contains black-box tests for the service layer.
// Package service_test 服务层的黑盒测试
package service_test

import (
	"context"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	domainmocks "github.com/haierkeys/fast-note-sync-service/internal/domain/mocks"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	servicemocks "github.com/haierkeys/fast-note-sync-service/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// 回归测试：删除操作必须幂等
//
// 背景：客户端在快速创建-删除竞争下会重复发送删除请求；服务端对"已删除"
// （note: GetByPathHashIncludeRecycle 过滤 action='delete'）与"不存在"
// （file: 直接透传 gorm.ErrRecordNotFound）的记录返回错误，而客户端仅在
// 收到删除确认后才清理本地哈希，收到错误不清理 → 本地状态残留 → 每次同步
// 重发 → 无限循环。
//
// 修复：Delete 对"记录不存在"与"已是删除状态"都返回成功（幂等），
// 并用请求参数构造返回 DTO（确认消息的 path/pathHash 必须可被客户端匹配）。
// ---------------------------------------------------------------------------

// TestNoteService_Delete_AlreadyDeleted_IsIdempotent 验证重复删除已删除的笔记返回成功。
// 修复前：GetByPathHashIncludeRecycle(isRecycle=false) 过滤 delete → ErrRecordNotFound → 430。
func TestNoteService_Delete_AlreadyDeleted_IsIdempotent(t *testing.T) {
	noteRepo := new(domainmocks.MockNoteRepository)
	vaultSvc := new(servicemocks.MockVaultService)
	vaultSvc.On("MustGetID", mock.Anything, int64(1), "vault").Return(int64(1), nil)

	deletedNote := &domain.Note{
		ID:               42,
		VaultID:          1,
		Action:           domain.NoteActionDelete,
		Path:             "a.md",
		PathHash:         "h1",
		UpdatedTimestamp: 1000,
	}
	// 修复前代码路径：查询过滤 delete → 找不到
	noteRepo.On("GetByPathHashIncludeRecycle", mock.Anything, "h1", int64(1), int64(1), false).
		Return(nil, gorm.ErrRecordNotFound).Maybe()
	// 修复后代码路径：不过滤 action → 找到已删除记录
	noteRepo.On("GetAllByPathHash", mock.Anything, "h1", int64(1), int64(1)).
		Return(deletedNote, nil).Maybe()

	svc := service.NewNoteService(nil, noteRepo, nil, nil, nil, nil, vaultSvc, nil, nil, nil, nil, nil)
	got, err := svc.Delete(context.Background(), 1, &dto.NoteDeleteRequest{
		Vault: "vault", Path: "a.md", PathHash: "h1",
	})

	require.NoError(t, err, "Bug: 重复删除已删除的笔记应幂等成功，但返回了错误（客户端会无限重试）")
	require.NotNil(t, got, "Bug: 幂等删除应返回非空 DTO（客户端 Ack 匹配需要 path）")
	assert.Equal(t, "a.md", got.Path)
	assert.Equal(t, "h1", got.PathHash)
	// 已删除记录不应被重复写库
	noteRepo.AssertNotCalled(t, "UpdateDelete", mock.Anything, mock.Anything, mock.Anything)
}

// TestNoteService_Delete_RecordNotFound_IsIdempotent 验证删除不存在的笔记返回成功。
// 场景：客户端"创建后秒删"，删除请求先于上传完成送达 → 服务端从未有该记录。
func TestNoteService_Delete_RecordNotFound_IsIdempotent(t *testing.T) {
	noteRepo := new(domainmocks.MockNoteRepository)
	vaultSvc := new(servicemocks.MockVaultService)
	vaultSvc.On("MustGetID", mock.Anything, int64(1), "vault").Return(int64(1), nil)

	noteRepo.On("GetByPathHashIncludeRecycle", mock.Anything, "h1", int64(1), int64(1), false).
		Return(nil, gorm.ErrRecordNotFound).Maybe()
	noteRepo.On("GetAllByPathHash", mock.Anything, "h1", int64(1), int64(1)).
		Return(nil, gorm.ErrRecordNotFound).Maybe()

	svc := service.NewNoteService(nil, noteRepo, nil, nil, nil, nil, vaultSvc, nil, nil, nil, nil, nil)
	got, err := svc.Delete(context.Background(), 1, &dto.NoteDeleteRequest{
		Vault: "vault", Path: "a.md", PathHash: "h1",
	})

	require.NoError(t, err, "Bug: 删除不存在的笔记应幂等成功（时序竞争下删除先于上传），但返回了错误")
	require.NotNil(t, got, "Bug: 幂等删除应返回用请求参数构造的 DTO")
	assert.Equal(t, "a.md", got.Path, "幂等返回的 path 必须来自请求参数（客户端按 path 匹配 pending）")
	assert.Equal(t, "h1", got.PathHash)
}

// TestNoteService_Delete_ActiveNote_MarksDeleted 基线：活跃笔记仍走正常删除流程。
func TestNoteService_Delete_ActiveNote_MarksDeleted(t *testing.T) {
	noteRepo := new(domainmocks.MockNoteRepository)
	shareRepo := new(domainmocks.MockUserShareRepository)
	vaultSvc := new(servicemocks.MockVaultService)
	vaultSvc.On("MustGetID", mock.Anything, int64(1), "vault").Return(int64(1), nil)

	activeNote := &domain.Note{
		ID:       42,
		VaultID:  1,
		Action:   domain.NoteActionCreate,
		Path:     "a.md",
		PathHash: "h1",
	}
	noteRepo.On("GetAllByPathHash", mock.Anything, "h1", int64(1), int64(1)).Return(activeNote, nil)
	noteRepo.On("GetByPathHashIncludeRecycle", mock.Anything, "h1", int64(1), int64(1), false).
		Return(activeNote, nil).Maybe()
	noteRepo.On("UpdateDelete", mock.Anything, mock.AnythingOfType("*domain.Note"), int64(1)).Return(nil)
	shareRepo.On("UpdateStatusByRes", mock.Anything, int64(1), "note", int64(42), domain.UserShareStatusRevoked).Return(nil)

	svc := service.NewNoteService(nil, noteRepo, nil, nil, shareRepo, nil, vaultSvc, nil, nil, nil, nil, nil)
	got, err := svc.Delete(context.Background(), 1, &dto.NoteDeleteRequest{
		Vault: "vault", Path: "a.md", PathHash: "h1",
	})

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, string(domain.NoteActionDelete), got.Action, "活跃笔记删除后 action 应为 delete")
	noteRepo.AssertCalled(t, "UpdateDelete", mock.Anything, mock.AnythingOfType("*domain.Note"), int64(1))
}

// TestFileService_Delete_RecordNotFound_IsIdempotent 验证删除不存在的文件返回成功。
// 修复前：fileService.Delete 直接透传 gorm.ErrRecordNotFound → 客户端收到 "record not found"。
func TestFileService_Delete_RecordNotFound_IsIdempotent(t *testing.T) {
	fileRepo := new(domainmocks.MockFileRepository)
	vaultSvc := new(servicemocks.MockVaultService)
	vaultSvc.On("MustGetID", mock.Anything, int64(1), "vault").Return(int64(1), nil)

	fileRepo.On("GetByPathHash", mock.Anything, "h2", int64(1), int64(1)).
		Return(nil, gorm.ErrRecordNotFound)

	svc := service.NewFileService(nil, fileRepo, nil, vaultSvc, nil, nil, nil, nil, nil)
	got, err := svc.Delete(context.Background(), 1, &dto.FileDeleteRequest{
		Vault: "vault", Path: "b.png", PathHash: "h2",
	})

	require.NoError(t, err, "Bug: 删除不存在的文件应幂等成功，但返回了错误（客户端会无限重试）")
	require.NotNil(t, got, "Bug: 幂等删除应返回用请求参数构造的 DTO")
	assert.Equal(t, "b.png", got.Path)
	assert.Equal(t, "h2", got.PathHash)
}

// TestFileService_Delete_AlreadyDeleted_IsIdempotent 验证重复删除已删除的文件返回成功。
func TestFileService_Delete_AlreadyDeleted_IsIdempotent(t *testing.T) {
	fileRepo := new(domainmocks.MockFileRepository)
	vaultSvc := new(servicemocks.MockVaultService)
	vaultSvc.On("MustGetID", mock.Anything, int64(1), "vault").Return(int64(1), nil)

	deletedFile := &domain.File{
		ID:       7,
		VaultID:  1,
		Action:   domain.FileActionDelete,
		Path:     "b.png",
		PathHash: "h2",
	}
	fileRepo.On("GetByPathHash", mock.Anything, "h2", int64(1), int64(1)).Return(deletedFile, nil)
	// 修复前路径会重复写库（Update 被调用）；定义为 Maybe 使失败表现为断言失败而非 mock panic
	fileRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.File"), int64(1)).
		Return(deletedFile, nil).Maybe()

	svc := service.NewFileService(nil, fileRepo, nil, vaultSvc, nil, nil, nil, nil, nil)
	got, err := svc.Delete(context.Background(), 1, &dto.FileDeleteRequest{
		Vault: "vault", Path: "b.png", PathHash: "h2",
	})

	require.NoError(t, err, "Bug: 重复删除已删除的文件应幂等成功")
	require.NotNil(t, got)
	assert.Equal(t, "b.png", got.Path)
	fileRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}
