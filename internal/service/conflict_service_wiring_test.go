// Package service implements the business logic layer.
// Package service 实现业务逻辑层。
package service

import (
	"context"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	domainmocks "github.com/haierkeys/fast-note-sync-service/internal/domain/mocks"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// fakeVaultServiceForConflictTest is a minimal VaultService stand-in used only to hand back a
// fixed vault ID. It can't be internal/service/mocks.MockVaultService here — that package
// imports internal/service (for the interface types), so importing it back from an
// internal/service test would create an import cycle.
// fakeVaultServiceForConflictTest 是仅用于返回固定 vault ID 的最小 VaultService 替身。
// 这里不能用 internal/service/mocks.MockVaultService ——该包反向导入了 internal/service
// （为了拿接口类型），从 internal/service 的测试里再导入它会形成导入环。
type fakeVaultServiceForConflictTest struct {
	VaultService
	vaultID int64
}

func (f *fakeVaultServiceForConflictTest) MustGetID(ctx context.Context, uid int64, name string) (int64, error) {
	return f.vaultID, nil
}

// TestConflictService_CreateConflictFile_PersistsSyncableCopy is the A1 wiring regression
// (ws_note.go's conflict branch on mergeResult.HasConflict || baseHashNotFound): the
// {name}.conflict.{ts}{ext} copy created for the client's content must be persisted through
// the normal note repository as a plain create, so it flows through standard sync
// distribution like any other note instead of being a side-channel record.
func TestConflictService_CreateConflictFile_PersistsSyncableCopy(t *testing.T) {
	noteRepo := new(domainmocks.MockNoteRepository)
	vaultSvc := &fakeVaultServiceForConflictTest{vaultID: 7}

	var createdNote *domain.Note
	noteRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Note"), int64(42)).
		Run(func(args mock.Arguments) {
			createdNote = args.Get(1).(*domain.Note)
		}).
		Return(&domain.Note{ID: 99}, nil)

	svc := NewConflictService(noteRepo, vaultSvc, zap.NewNop())

	req := &dto.ConflictFileRequest{
		Vault:             "MyVault",
		OriginalPath:      "Notes/ReadMe.md",
		ClientContent:     "client side content",
		ClientContentHash: "client-hash",
		Ctime:             1000,
		Mtime:             2000,
	}

	resp, err := svc.CreateConflictFile(context.Background(), 42, req)
	if err != nil {
		t.Fatalf("CreateConflictFile returned error: %v", err)
	}

	if resp.NoteID != 99 {
		t.Fatalf("resp.NoteID = %d, want 99 (the ID the repo actually created)", resp.NoteID)
	}
	if resp.ConflictPath == req.OriginalPath {
		t.Fatalf("conflict path must differ from the original path, got %q", resp.ConflictPath)
	}

	if createdNote == nil {
		t.Fatal("noteRepo.Create was never invoked — conflict copy was not persisted")
	}
	if createdNote.VaultID != 7 {
		t.Fatalf("created note VaultID = %d, want 7", createdNote.VaultID)
	}
	if createdNote.Action != domain.NoteActionCreate {
		t.Fatalf("created note Action = %q, want %q — anything else would not flow through normal sync distribution", createdNote.Action, domain.NoteActionCreate)
	}
	if createdNote.Content != req.ClientContent {
		t.Fatalf("created note Content = %q, want client content %q", createdNote.Content, req.ClientContent)
	}
	if createdNote.ContentHash != req.ClientContentHash {
		t.Fatalf("created note ContentHash = %q, want %q", createdNote.ContentHash, req.ClientContentHash)
	}
}
