package websocket_router

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/retirement"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
)

func TestNoteRetirementAckPrecedesSenderTombstone(t *testing.T) {
	params := &dto.NoteModifyOrCreateRequest{Vault: "Files", Path: "old.md", PathHash: "old-path-hash", Context: "pending-123"}
	current := &dto.NoteDTO{Action: "delete", Path: params.Path, PathHash: params.PathHash, Ctime: 5, Mtime: 10, Size: 20, UpdatedTimestamp: 30}
	var actions []string
	finishRetiredNote(params, current, func(response *code.Code, action ...string) {
		actions = append(actions, action[0])
		if !response.Status() || response.Vault() != "Files" {
			t.Fatal("missing transport success")
		}
		if len(actions) == 1 {
			ack := response.Data().(dto.NoteModifyAckMessage)
			if ack.LastTime != 30 || ack.PathHash != params.PathHash || response.Context() != params.Context {
				t.Fatal("pending note cannot be cleared")
			}
		} else {
			state := response.Data().(dto.NoteSyncDeleteMessage)
			if state.UpdatedTimestamp != 30 || state.Mtime != 10 || state.Path != params.Path {
				t.Fatal("authoritative tombstone metadata lost")
			}
		}
	})
	if want := []string{string(NoteModifyAck), string(NoteSyncDelete)}; !reflect.DeepEqual(actions, want) {
		t.Fatalf("wrong order: %v", actions)
	}
}

func TestNoteRetirementActiveStatePreservesNewParagraphs(t *testing.T) {
	params := &dto.NoteModifyOrCreateRequest{Vault: "Files", Path: "old.md", PathHash: "path-hash", Content: "retired bytes"}
	current := &dto.NoteDTO{Action: "modify", Path: params.Path, PathHash: params.PathHash, Content: "new paragraphs", ContentHash: "current-hash", UpdatedTimestamp: 40}
	var actions []string
	finishRetiredNote(params, current, func(response *code.Code, action ...string) {
		actions = append(actions, action[0])
		if action[0] == string(NoteSyncModify) {
			message := response.Data().(dto.NoteSyncModifyMessage)
			if message.Content != "new paragraphs" || message.ContentHash != "current-hash" {
				t.Fatal("rejected bytes replaced current content")
			}
		}
	})
	if want := []string{string(NoteModifyAck), string(NoteSyncModify)}; !reflect.DeepEqual(actions, want) {
		t.Fatalf("current active content was deleted: %v", actions)
	}
}

func TestNoteRetirementUnavailableCurrentStateStillClearsPendingWithoutDeleting(t *testing.T) {
	var actions []string
	finishRetiredNote(&dto.NoteModifyOrCreateRequest{Vault: "Files", Path: "old.md", Context: "ctx"}, nil, func(response *code.Code, action ...string) {
		actions = append(actions, action[0])
		if response.Context() != "ctx" {
			t.Fatal("missing pending context")
		}
	})
	if want := []string{string(NoteModifyAck)}; !reflect.DeepEqual(actions, want) {
		t.Fatalf("state failure emitted delete: %v", actions)
	}
}

type noteWSRetirementRepo struct {
	domain.NoteRepository
	mu  sync.Mutex
	row *domain.Note
}

func (r *noteWSRetirementRepo) GetByPath(context.Context, string, int64, int64) (*domain.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := *r.row
	return &copy, nil
}
func (r *noteWSRetirementRepo) Update(_ context.Context, note *domain.Note, _ int64) (*domain.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := *note
	copy.UpdatedTimestamp++
	r.row = &copy
	result := copy
	return &result, nil
}
func (r *noteWSRetirementRepo) CountSizeSum(context.Context, int64, int64) (*domain.CountSizeResult, error) {
	return &domain.CountSizeResult{}, nil
}

type noteWSRetirementVault struct{ service.VaultService }

func (noteWSRetirementVault) MustGetID(context.Context, int64, string) (int64, error) { return 1, nil }
func (noteWSRetirementVault) UpdateNoteStats(context.Context, int64, int64, int64, int64) error {
	return nil
}

type noteWSRetirementFolder struct{ service.FolderService }

func (noteWSRetirementFolder) SyncResourceFID(context.Context, int64, int64, []int64, []int64) error {
	return nil
}

func TestNoteRetirementReconcileAfterPostRejectPreReplyNewUpload(t *testing.T) {
	ctx := context.Background()
	path := "moved.md"
	old := "retired old paragraphs"
	digest := sha256.Sum256([]byte(old))
	policy := filepath.Join(t.TempDir(), "policy.json")
	doc, _ := json.Marshal(retirement.Document{Version: 1, Entries: []retirement.Entry{{UID: 1, VaultID: 1, Path: path, SHA256: hex.EncodeToString(digest[:])}}})
	if err := os.WriteFile(policy, doc, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(retirement.EnvPolicy, policy)
	repo := &noteWSRetirementRepo{row: &domain.Note{ID: 9, VaultID: 1, Path: path, PathHash: util.EncodeHash32(path), Content: old, ContentHash: util.EncodeHash32(old), Action: domain.NoteActionDelete, Mtime: 1, UpdatedTimestamp: 10}}
	svc := service.NewNoteService(nil, repo, nil, nil, nil, nil, noteWSRetirementVault{}, noteWSRetirementFolder{}, nil, nil, nil, nil)
	params := &dto.NoteModifyOrCreateRequest{Vault: "Files", Path: path, PathHash: util.EncodeHash32(path), Content: old, ContentHash: util.EncodeHash32(old), Mtime: 2, Context: "old-upload"}
	_, err := service.CheckNoteRetirement(ctx, svc, 1, params)
	var retired *service.RetiredNoteBytesError
	if !errors.As(err, &retired) || retired.Current.Action != "delete" {
		t.Fatalf("missing initial rejection tombstone: %v", err)
	}
	// Controlled postReject / preReply barrier: a genuine new upload commits
	// through the production service before the terminal WS response is built.
	committed := make(chan error, 1)
	go func() {
		fresh := *params
		fresh.Content = "new paragraphs committed between reject and reply"
		_, _, err := svc.WithClient("desktop", "new device", "2.4.0").ModifyOrCreate(ctx, 1, &fresh, true)
		committed <- err
	}()
	if err := <-committed; err != nil {
		t.Fatal(err)
	}
	var actions []string
	err = reconcileRetiredNote(ctx, svc, 1, params, retired.Current, func(response *code.Code, action ...string) {
		actions = append(actions, action[0])
		if action[0] == string(NoteSyncDelete) {
			t.Fatal("late stale deletion erased a committed new note")
		}
		if action[0] == string(NoteSyncModify) && response.Data().(dto.NoteSyncModifyMessage).Content != "new paragraphs committed between reject and reply" {
			t.Fatal("reply did not use authoritative committed body")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{string(NoteModifyAck), string(NoteSyncModify)}; !reflect.DeepEqual(actions, want) {
		t.Fatalf("wrong production reconciliation: %v", actions)
	}
}
