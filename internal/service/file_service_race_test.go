package service

import (
	"context"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
)

// mockFileRepoWithRename supports the race condition test with RenamedToID
type mockFileRepoWithRename struct {
	domain.FileRepository
	files   []*domain.File
	nextID  int64
	byID    map[int64]*domain.File
	byHash  map[string]*domain.File
	byPath  map[string]*domain.File
}

func (m *mockFileRepoWithRename) GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*domain.File, error) {
	if f, ok := m.byHash[pathHash]; ok && f.VaultID == vaultID {
		return f, nil
	}
	return nil, nil
}

func (m *mockFileRepoWithRename) GetByPath(ctx context.Context, path string, vaultID, uid int64) (*domain.File, error) {
	if f, ok := m.byPath[path]; ok && f.VaultID == vaultID {
		return f, nil
	}
	return nil, nil
}

func (m *mockFileRepoWithRename) GetByID(ctx context.Context, id, uid int64) (*domain.File, error) {
	if f, ok := m.byID[id]; ok {
		return f, nil
	}
	return nil, nil
}

func (m *mockFileRepoWithRename) Create(ctx context.Context, file *domain.File, uid int64) (*domain.File, error) {
	file.ID = m.nextID
	m.nextID++
	m.files = append(m.files, file)
	m.byID[file.ID] = file
	m.byPath[file.Path] = file
	m.byHash[file.PathHash] = file
	return file, nil
}

func (m *mockFileRepoWithRename) Update(ctx context.Context, file *domain.File, uid int64) (*domain.File, error) {
	if f, ok := m.byID[file.ID]; ok {
		f.Path = file.Path
		f.PathHash = file.PathHash
		f.SavePath = file.SavePath
		f.ContentHash = file.ContentHash
		f.Action = file.Action
		f.Rename = file.Rename
		f.RenamedToID = file.RenamedToID
		f.Size = file.Size
		f.Mtime = file.Mtime
		f.Ctime = file.Ctime
		f.FID = file.FID
		f.UpdatedTimestamp = file.UpdatedTimestamp
		return f, nil
	}
	return nil, nil
}

func (m *mockFileRepoWithRename) UpdateMtime(ctx context.Context, mt int64, id, uid int64) error { return nil }
func (m *mockFileRepoWithRename) UpdateActionMtime(ctx context.Context, action domain.FileAction, mt int64, id, uid int64) error { return nil }
func (m *mockFileRepoWithRename) UpdateFID(ctx context.Context, id, fid, uid int64) error { return nil }
func (m *mockFileRepoWithRename) Delete(ctx context.Context, id, uid int64) error { return nil }
func (m *mockFileRepoWithRename) CountSizeSum(ctx context.Context, vaultID, uid int64) (*domain.CountSizeResult, error) { return nil, nil }

// mockVault returns a fixed vault ID
type mockVault struct{}

func (s *mockVault) GetByName(ctx context.Context, uid int64, name string) (*domain.Vault, error) {
	return &domain.Vault{ID: 1, Name: name}, nil
}
func (s *mockVault) GetOrCreate(ctx context.Context, uid int64, name string) (*domain.Vault, error) {
	return &domain.Vault{ID: 1, Name: name}, nil
}
func (s *mockVault) MustGetID(ctx context.Context, uid int64, name string) (int64, error) {
	return 1, nil
}
func (s *mockVault) Create(ctx context.Context, uid int64, name string) (*dto.VaultDTO, error) {
	return &dto.VaultDTO{ID: 1}, nil
}
func (s *mockVault) Update(ctx context.Context, uid int64, id int64, name string) (*dto.VaultDTO, error) {
	return &dto.VaultDTO{ID: id}, nil
}
func (s *mockVault) Get(ctx context.Context, uid int64, id int64) (*dto.VaultDTO, error) {
	return &dto.VaultDTO{ID: id}, nil
}
func (s *mockVault) List(ctx context.Context, uid int64) ([]*dto.VaultDTO, error) { return nil, nil }
func (s *mockVault) Delete(ctx context.Context, uid int64, id int64) error { return nil }
func (s *mockVault) UpdateNoteStats(ctx context.Context, noteSize, noteCount, vaultID, uid int64) error { return nil }
func (s *mockVault) UpdateFileStats(ctx context.Context, fileSize, fileCount, vaultID, uid int64) error { return nil }

type mockFolder struct{}

func (s *mockFolder) Get(ctx context.Context, uid int64, params *dto.FolderGetRequest) (*dto.FolderDTO, error) {
	return nil, nil
}
func (s *mockFolder) List(ctx context.Context, uid int64, params *dto.FolderListRequest) ([]*dto.FolderDTO, error) {
	return nil, nil
}
func (s *mockFolder) ListByUpdatedTimestamp(ctx context.Context, uid int64, vault string, lastTime int64) ([]*dto.FolderDTO, error) {
	return nil, nil
}
func (s *mockFolder) UpdateOrCreate(ctx context.Context, uid int64, params *dto.FolderCreateRequest) (*dto.FolderDTO, error) {
	return nil, nil
}
func (s *mockFolder) Delete(ctx context.Context, uid int64, params *dto.FolderDeleteRequest) (*dto.FolderDTO, error) {
	return nil, nil
}
func (s *mockFolder) Rename(ctx context.Context, uid int64, params *dto.FolderRenameRequest) (*dto.FolderDTO, *dto.FolderDTO, error) {
	return nil, nil, nil
}
func (s *mockFolder) ListNotes(ctx context.Context, uid int64, params *dto.FolderContentRequest, pager *app.Pager) ([]*dto.NoteNoContentDTO, int, error) {
	return nil, 0, nil
}
func (s *mockFolder) ListFiles(ctx context.Context, uid int64, params *dto.FolderContentRequest, pager *app.Pager) ([]*dto.FileDTO, int, error) {
	return nil, 0, nil
}
func (s *mockFolder) EnsurePathFID(ctx context.Context, uid int64, vaultID int64, path string) (int64, error) {
	return 1, nil
}
func (s *mockFolder) SyncResourceFID(ctx context.Context, uid int64, vaultID int64, noteIDs []int64, fileIDs []int64) error {
	return nil
}
func (s *mockFolder) GetTree(ctx context.Context, uid int64, params *dto.FolderTreeRequest) (*dto.FolderTreeResponse, error) {
	return nil, nil
}
func (s *mockFolder) CleanDuplicateFolders(ctx context.Context, uid int64, vaultID int64) error {
	return nil
}

// TestRenameUploadRaceCondition verifies that when Rename marks record A as Delete
// with Rename=1, and UploadComplete arrives, the fix routes the update to record B
// via RenamedToID, preventing record B's SavePath from becoming stale.
func TestRenameUploadRaceCondition(t *testing.T) {
	ctx := context.Background()
	uid := int64(1)
	vaultID := int64(1)

	// Record A: being uploaded, Rename marks it Delete, sets Rename=1
	// Rename will set RenamedToID = 101 after creating record B
	recordA := &domain.File{
		ID:          100, VaultID: vaultID,
		Action:      domain.FileActionDelete, Rename: 1,
		Path:        "dir/old_name.webp", PathHash: "h_old",
		SavePath:    "storage/temp/upload_xxx",
		RenamedToID: 101, // Will be set by Rename after creating record B
	}

	// Record B: created by Rename, reuses A's SavePath
	recordB := &domain.File{
		ID:          101, VaultID: vaultID,
		Action:      domain.FileActionCreate, Rename: 0,
		Path:        "dir/new_name.webp", PathHash: "h_new",
		SavePath:    "storage/temp/upload_xxx",
	}

	mockRepo := &mockFileRepoWithRename{
		files:  []*domain.File{recordA, recordB},
		nextID: 102,
		byID:   map[int64]*domain.File{100: recordA, 101: recordB},
		byPath: map[string]*domain.File{
			"dir/old_name.webp": recordA,
			"dir/new_name.webp": recordB,
		},
		byHash: map[string]*domain.File{"h_old": recordA, "h_new": recordB},
	}

	svc := NewFileService(mockRepo, nil, &mockVault{}, &mockFolder{}, nil, nil, nil).(*fileService)

	params := &dto.FileUpdateRequest{
		Vault:       "vault",
		Path:        "dir/old_name.webp",
		PathHash:    "h_old",
		ContentHash: "ch",
		SavePath:    "storage/temp/upload_xxx",
		Size:        100,
		Ctime:       1000,
		Mtime:       1000,
	}

	_, dto, err := svc.UpdateOrCreate(ctx, uid, params, true)
	if err != nil {
		t.Fatalf("UpdateOrCreate failed: %v", err)
	}

	// Record B (ID=101) must be updated, not A (ID=100)
	if dto == nil || dto.ID != 101 {
		t.Errorf("expected returned DTO to be record B (ID=101), got %v", dto)
	}
}
