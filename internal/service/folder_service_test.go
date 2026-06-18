// Package service implements the business logic layer.
// Package service 实现业务逻辑层。
package service

import (
	"context"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	domainmocks "github.com/haierkeys/fast-note-sync-service/internal/domain/mocks"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
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

func TestFolderService_DeleteTree_RecursivelySoftDeletesChildrenAndResources(t *testing.T) {
	ctx := context.Background()
	uid := int64(1)
	vaultID := int64(9)
	vaultName := "vault"

	folderRepo := new(domainmocks.MockFolderRepository)
	noteRepo := new(domainmocks.MockNoteRepository)
	fileRepo := new(domainmocks.MockFileRepository)
	vaultRepo := new(domainmocks.MockVaultRepository)

	root := &domain.Folder{ID: 10, VaultID: vaultID, Action: domain.FolderActionCreate, Path: "Projects", PathHash: util.EncodeHash32("Projects")}
	child := &domain.Folder{ID: 11, VaultID: vaultID, Action: domain.FolderActionCreate, Path: "Projects/Archive", PathHash: util.EncodeHash32("Projects/Archive")}
	note := &domain.Note{ID: 20, VaultID: vaultID, Action: domain.NoteActionCreate, Path: "Projects/Archive/todo.md", PathHash: util.EncodeHash32("Projects/Archive/todo.md")}
	file := &domain.File{ID: 30, VaultID: vaultID, Action: domain.FileActionCreate, Path: "Projects/assets/image.png", PathHash: util.EncodeHash32("Projects/assets/image.png")}

	vaultRepo.On("GetByName", mock.Anything, vaultName, uid).Return(&domain.Vault{ID: vaultID, Name: vaultName}, nil)
	folderRepo.On("GetAllByPathHash", mock.Anything, root.PathHash, vaultID, uid).Return([]*domain.Folder{root}, nil)
	folderRepo.On("ListByPathPrefix", mock.Anything, "Projects", vaultID, uid).Return([]*domain.Folder{child}, nil)
	noteRepo.On("ListByPathPrefix", mock.Anything, "Projects", vaultID, uid).Return([]*domain.Note{note}, nil)
	fileRepo.On("ListByPathPrefix", mock.Anything, "Projects", vaultID, uid).Return([]*domain.File{file}, nil)

	folderRepo.On("Update", mock.Anything, mock.MatchedBy(func(f *domain.Folder) bool {
		return f.ID == child.ID && f.Action == domain.FolderActionDelete
	}), uid).Return(child, nil)
	folderRepo.On("Update", mock.Anything, mock.MatchedBy(func(f *domain.Folder) bool {
		return f.ID == root.ID && f.Action == domain.FolderActionDelete
	}), uid).Return(root, nil)
	noteRepo.On("UpdateDelete", mock.Anything, mock.MatchedBy(func(n *domain.Note) bool {
		return n.ID == note.ID && n.Action == domain.NoteActionDelete && n.Rename == 0
	}), uid).Return(nil)
	fileRepo.On("Update", mock.Anything, mock.MatchedBy(func(f *domain.File) bool {
		return f.ID == file.ID && f.Action == domain.FileActionDelete && f.Rename == 0
	}), uid).Return(file, nil)

	svc := &folderService{
		folderRepo:   folderRepo,
		noteRepo:     noteRepo,
		fileRepo:     fileRepo,
		vaultService: NewVaultService(vaultRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, zap.NewNop()),
	}

	got, err := svc.DeleteTree(ctx, uid, &dto.FolderDeleteRequest{Vault: vaultName, Path: "Projects"})

	assert.NoError(t, err)
	assert.Equal(t, "Projects", got.Path)
	folderRepo.AssertExpectations(t)
	noteRepo.AssertExpectations(t)
	fileRepo.AssertExpectations(t)
	vaultRepo.AssertExpectations(t)
}

func TestFolderService_DeleteTree_RejectsRootPath(t *testing.T) {
	ctx := context.Background()
	uid := int64(1)
	vaultID := int64(9)
	vaultName := "vault"

	folderRepo := new(domainmocks.MockFolderRepository)
	vaultRepo := new(domainmocks.MockVaultRepository)
	vaultRepo.On("GetByName", mock.Anything, vaultName, uid).Return(&domain.Vault{ID: vaultID, Name: vaultName}, nil)

	svc := &folderService{
		folderRepo:   folderRepo,
		vaultService: NewVaultService(vaultRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, zap.NewNop()),
	}

	got, err := svc.DeleteTree(ctx, uid, &dto.FolderDeleteRequest{Vault: vaultName, Path: "/"})

	assert.Nil(t, got)
	assert.Error(t, err)
	folderRepo.AssertNotCalled(t, "GetAllByPathHash", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	vaultRepo.AssertExpectations(t)
}

func TestFolderService_CleanupEmptyAncestors_StopsAtFirstNonEmptyFolder(t *testing.T) {
	ctx := context.Background()
	uid := int64(1)
	vaultID := int64(9)

	folderRepo := new(domainmocks.MockFolderRepository)
	noteRepo := new(domainmocks.MockNoteRepository)
	fileRepo := new(domainmocks.MockFileRepository)

	emptyChild := &domain.Folder{ID: 11, VaultID: vaultID, Action: domain.FolderActionCreate, Path: "Projects/Archive", PathHash: util.EncodeHash32("Projects/Archive")}
	nonEmptyParent := &domain.Folder{ID: 10, VaultID: vaultID, Action: domain.FolderActionCreate, Path: "Projects", PathHash: util.EncodeHash32("Projects")}

	folderRepo.On("GetAllByPathHash", mock.Anything, emptyChild.PathHash, vaultID, uid).Return([]*domain.Folder{emptyChild}, nil)
	folderRepo.On("GetByFID", mock.Anything, emptyChild.ID, vaultID, uid).Return([]*domain.Folder{}, nil)
	noteRepo.On("ListByFIDsCount", mock.Anything, []int64{emptyChild.ID}, vaultID, uid).Return(int64(0), nil)
	fileRepo.On("ListByFIDsCount", mock.Anything, []int64{emptyChild.ID}, vaultID, uid).Return(int64(0), nil)
	folderRepo.On("Update", mock.Anything, mock.MatchedBy(func(f *domain.Folder) bool {
		return f.ID == emptyChild.ID && f.Action == domain.FolderActionDelete
	}), uid).Return(emptyChild, nil)

	folderRepo.On("GetAllByPathHash", mock.Anything, nonEmptyParent.PathHash, vaultID, uid).Return([]*domain.Folder{nonEmptyParent}, nil)
	folderRepo.On("GetByFID", mock.Anything, nonEmptyParent.ID, vaultID, uid).Return([]*domain.Folder{}, nil)
	noteRepo.On("ListByFIDsCount", mock.Anything, []int64{nonEmptyParent.ID}, vaultID, uid).Return(int64(1), nil)

	svc := &folderService{
		folderRepo: folderRepo,
		noteRepo:   noteRepo,
		fileRepo:   fileRepo,
	}

	err := svc.CleanupEmptyAncestors(ctx, uid, vaultID, "Projects/Archive/moved.md")

	assert.NoError(t, err)
	folderRepo.AssertExpectations(t)
	noteRepo.AssertExpectations(t)
	fileRepo.AssertExpectations(t)
}
