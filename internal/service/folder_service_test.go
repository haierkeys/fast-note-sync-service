package service

import (
	"context"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
)

type mockFolderRepo struct {
	domain.FolderRepository
	folders    []*domain.Folder
	deletedIDs []int64
}

func (m *mockFolderRepo) ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.Folder, error) {
	return m.folders, nil
}

func (m *mockFolderRepo) Delete(ctx context.Context, id, uid int64) error {
	m.deletedIDs = append(m.deletedIDs, id)
	return nil
}

func TestCleanDuplicateFolders(t *testing.T) {
	ctx := context.Background()
	uid := int64(1)
	vaultID := int64(1)

	tests := []struct {
		name           string
		folders        []*domain.Folder
		wantDeletedIDs []int64
	}{
		{
			name: "mixed deleted and active - delete active",
			folders: []*domain.Folder{
				{ID: 1, PathHash: "h1", Action: domain.FolderActionDelete},
				{ID: 2, PathHash: "h1", Action: domain.FolderActionCreate},
			},
			wantDeletedIDs: []int64{2},
		},
		{
			name: "all active - keep max ID",
			folders: []*domain.Folder{
				{ID: 3, PathHash: "h2", Action: domain.FolderActionCreate},
				{ID: 4, PathHash: "h2", Action: domain.FolderActionCreate},
				{ID: 5, PathHash: "h2", Action: domain.FolderActionCreate},
			},
			wantDeletedIDs: []int64{3, 4},
		},
		{
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
			repo := &mockFolderRepo{folders: tt.folders}
			svc := &folderService{folderRepo: repo}

			err := svc.CleanDuplicateFolders(ctx, uid, vaultID)
			if err != nil {
				t.Fatalf("CleanDuplicateFolders failed: %v", err)
			}

			if len(repo.deletedIDs) != len(tt.wantDeletedIDs) {
				t.Errorf("deleted length mismatch: got %v, want %v", repo.deletedIDs, tt.wantDeletedIDs)
				return
			}

			found := make(map[int64]bool)
			for _, id := range repo.deletedIDs {
				found[id] = true
			}

			for _, id := range tt.wantDeletedIDs {
				if !found[id] {
					t.Errorf("expected ID %d to be deleted, but it wasn't", id)
				}
			}
		})
	}
}
