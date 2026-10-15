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

type retirementReplyRepo struct {
	domain.FileRepository
	mu  sync.Mutex
	row *domain.File
}

func (r *retirementReplyRepo) GetByPath(context.Context, string, int64, int64) (*domain.File, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := *r.row
	return &copy, nil
}
func (r *retirementReplyRepo) Update(_ context.Context, row *domain.File, _ int64) (*domain.File, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := *row
	copy.UpdatedTimestamp++
	r.row = &copy
	result := copy
	return &result, nil
}
func (*retirementReplyRepo) CountSizeSum(context.Context, int64, int64) (*domain.CountSizeResult, error) {
	return &domain.CountSizeResult{}, nil
}

type retirementReplyVault struct{ service.VaultService }

func (retirementReplyVault) MustGetID(context.Context, int64, string) (int64, error) { return 1, nil }
func (retirementReplyVault) UpdateFileStats(context.Context, int64, int64, int64, int64) error {
	return nil
}

type retirementReplyFolder struct{ service.FolderService }

func (retirementReplyFolder) SyncResourceFID(context.Context, int64, int64, []int64, []int64) error {
	return nil
}

// Controlled handoff: old upload is rejected, another goroutine commits a new
// active generation, and only then the actual WS reconciliation helper replies.
// A handler that reuses RetiredBytesError.Current would send a stale delete here.
func TestRetirementWSReplyAfterConcurrentNewUploadOrRestore(t *testing.T) {
	for _, restore := range []bool{false, true} {
		t.Run(map[bool]string{false: "new-upload", true: "explicit-restore"}[restore], func(t *testing.T) {
			dir := t.TempDir()
			name := "Images/old.png"
			hash := util.EncodeHash32(name)
			old := []byte("approved retired bytes")
			sum := sha256.Sum256(old)
			policy := filepath.Join(dir, "policy.json")
			doc, _ := json.Marshal(retirement.Document{Version: 1, Entries: []retirement.Entry{{UID: 1, VaultID: 1, Path: name, SHA256: hex.EncodeToString(sum[:])}}})
			if err := os.WriteFile(policy, doc, 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv(retirement.EnvPolicy, policy)
			blob := filepath.Join(dir, "old-blob")
			temp := filepath.Join(dir, "upload-temp")
			if err := os.WriteFile(blob, old, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(temp, old, 0600); err != nil {
				t.Fatal(err)
			}
			repo := &retirementReplyRepo{row: &domain.File{ID: 7, VaultID: 1, Path: name, PathHash: hash, ContentHash: "old", SavePath: blob, Action: domain.FileActionDelete, UpdatedTimestamp: 10}}
			svc := service.NewFileService(nil, repo, nil, retirementReplyVault{}, retirementReplyFolder{}, nil, nil, nil, nil)
			ctx := context.Background()
			params := &dto.FileUpdateRequest{Vault: "Files", Path: name, PathHash: hash, ContentHash: "old", SavePath: temp, Mtime: 2}
			_, _, err := svc.UploadComplete(ctx, 1, params)
			var rejected *service.RetiredBytesError
			if !errors.As(err, &rejected) || rejected.Current.Action != "delete" {
				t.Fatalf("missing old tombstone rejection: %v", err)
			}
			rejectedHook := make(chan struct{})
			committed := make(chan error, 1)
			go func() {
				<-rejectedHook
				if restore {
					_, err := svc.Restore(ctx, 1, &dto.FileRestoreRequest{Vault: "Files", Path: name, PathHash: hash})
					committed <- err
					return
				}
				newBlob := filepath.Join(dir, "new-blob")
				if err := os.WriteFile(newBlob, []byte("intentional new content"), 0600); err != nil {
					committed <- err
					return
				}
				newParams := *params
				newParams.SavePath = newBlob
				newParams.ContentHash = "new"
				_, _, err := svc.UploadComplete(ctx, 1, &newParams)
				committed <- err
			}()
			close(rejectedHook)
			if err := <-committed; err != nil {
				t.Fatal(err)
			}
			session := &FileUploadBinaryChunkSession{Vault: "Files", Path: name, PathHash: hash, Context: "old-upload", SavePath: temp}
			var sequence []string
			err = reconcileRetiredUpload(ctx, svc, 1, session, rejected, func(response *code.Code, actions ...string) {
				sequence = append(sequence, actions[0])
				if actions[0] == string(FileSyncDelete) {
					t.Fatal("stale tombstone deleted newer active generation")
				}
				if actions[0] == string(FileSyncUpdate) {
					message := response.Data().(dto.FileSyncModifyMessage)
					want := "new"
					if restore {
						want = "old"
					}
					if message.ContentHash != want {
						t.Fatalf("reply content %q, want %q", message.ContentHash, want)
					}
				}
			}, func() { sequence = append(sequence, "RemoveSession") })
			if err != nil {
				t.Fatal(err)
			}
			if want := []string{string(FileUploadAck), "RemoveSession", string(FileSyncUpdate)}; !reflect.DeepEqual(sequence, want) {
				t.Fatalf("reply sequence %v", sequence)
			}
		})
	}
}
