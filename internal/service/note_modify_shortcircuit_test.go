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
)

// ---------------------------------------------------------------------------
// 回归测试：ModifyOrCreate 的幂等短路必须返回有效 DTO
//
// 背景：客户端重连后重复上传"内容与 mtime 均一致"的笔记时，
// ModifyOrCreate 的幂等短路分支返回 (dto=nil, err=nil)，调用方
// ws_note.go:419 直接访问 note.UpdatedTimestamp → 空指针 panic。
//
// 修复：短路分支与相邻的"仅更新 mtime"分支保持一致，返回 s.domainToDTO(note)。
// ---------------------------------------------------------------------------

// TestNoteService_ModifyOrCreate_IdempotentShortCircuit_ReturnsDTO 验证内容+mtime 一致时
// 短路分支返回非空 DTO（修复前返回 nil，调用方访问字段即 panic）。
func TestNoteService_ModifyOrCreate_IdempotentShortCircuit_ReturnsDTO(t *testing.T) {
	noteRepo := new(domainmocks.MockNoteRepository)
	vaultSvc := new(servicemocks.MockVaultService)
	vaultSvc.On("MustGetID", mock.Anything, int64(1), "vault").Return(int64(1), nil)

	existing := &domain.Note{
		ID:               42,
		VaultID:          1,
		Action:           domain.NoteActionCreate,
		Path:             "a.md",
		PathHash:         "h1",
		Content:          "hello",
		ContentHash:      "ch1",
		Mtime:            1000,
		UpdatedTimestamp: 1500,
	}
	noteRepo.On("GetAllByPathHash", mock.Anything, "h1", int64(1), int64(1)).Return(existing, nil)

	svc := service.NewNoteService(nil, noteRepo, nil, nil, nil, nil, vaultSvc, nil, nil, nil, nil, nil)
	isNew, got, err := svc.ModifyOrCreate(context.Background(), 1, &dto.NoteModifyOrCreateRequest{
		Vault:       "vault",
		Path:        "a.md",
		PathHash:    "h1",
		Content:     "hello",
		ContentHash: "ch1",
		Mtime:       1000,
	}, true)

	require.NoError(t, err)
	assert.False(t, isNew)
	require.NotNil(t, got, "Bug: 幂等短路（内容+mtime 一致）必须返回有效 DTO——返回 nil 导致调用方 ws_note.go:419 空指针 panic")
	assert.Equal(t, "a.md", got.Path)
	assert.Equal(t, "h1", got.PathHash)
	assert.Equal(t, int64(1500), got.UpdatedTimestamp)
	// 短路不应写库
	noteRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}
