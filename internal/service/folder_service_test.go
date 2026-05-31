// Package service implements the business logic layer.
// Package service 实现业务逻辑层。
package service

import (
	"context"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	domainmocks "github.com/haierkeys/fast-note-sync-service/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestCleanDuplicateFolders uses table-driven tests to verify dedup logic.
// TestCleanDuplicateFolders 使用表驱动测试验证重复文件夹清理逻辑。
func TestCleanDuplicateFolders(t *testing.T) {
	ctx := context.Background()
	uid := int64(1)
	vaultID := int64(1)

	tests := []struct {
		name           string           // test case name / 测试用例名称
		folders        []*domain.Folder // input folders / 输入的文件夹列表
		wantDeletedIDs []int64          // expected deleted IDs (order-insensitive) / 期望被删除的 ID（顺序无关）
	}{
		{
			// When a hash has both delete and create records, the active ones should be deleted.
			// 当同一 hash 既有 delete 也有 create 记录时，应删除 active 记录。
			name: "mixed deleted and active - delete active",
			folders: []*domain.Folder{
				{ID: 1, PathHash: "h1", Action: domain.FolderActionDelete},
				{ID: 2, PathHash: "h1", Action: domain.FolderActionCreate},
			},
			wantDeletedIDs: []int64{2},
		},
		{
			// When all records are active, keep the highest ID and delete the rest.
			// 当所有记录都是 active 时，保留最大 ID，删除其余记录。
			name: "all active - keep max ID",
			folders: []*domain.Folder{
				{ID: 3, PathHash: "h2", Action: domain.FolderActionCreate},
				{ID: 4, PathHash: "h2", Action: domain.FolderActionCreate},
				{ID: 5, PathHash: "h2", Action: domain.FolderActionCreate},
			},
			wantDeletedIDs: []int64{3, 4},
		},
		{
			// When each hash has only one record, nothing should be deleted.
			// 当每个 hash 只有一条记录时，不应删除任何内容。
			name: "no duplicates - delete nothing",
			folders: []*domain.Folder{
				{ID: 6, PathHash: "h3", Action: domain.FolderActionCreate},
				{ID: 7, PathHash: "h4", Action: domain.FolderActionCreate},
			},
			wantDeletedIDs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(domainmocks.MockFolderRepository)

			// Stub ListByUpdatedTimestamp to return the test fixture folders.
			// Stub ListByUpdatedTimestamp 返回测试固定数据的文件夹列表。
			mockRepo.On("ListByUpdatedTimestamp", mock.Anything, int64(0), vaultID, uid).
				Return(tt.folders, nil)

			// Stub Delete for each expected deleted ID.
			// 为每个期望删除的 ID Stub Delete。
			for _, id := range tt.wantDeletedIDs {
				mockRepo.On("Delete", mock.Anything, id, uid).Return(nil)
			}

			svc := &folderService{folderRepo: mockRepo}
			err := svc.CleanDuplicateFolders(ctx, uid, vaultID)

			assert.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})
	}
}
