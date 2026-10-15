package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/retirement"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"gorm.io/gorm"
)

type retirementTestRepo struct {
	domain.FileRepository
	mu                                     sync.Mutex
	row                                    *domain.File
	writes, mtimeWrites, actionMtimeWrites int
}

func (r *retirementTestRepo) GetByPath(_ context.Context, name string, _, _ int64) (*domain.File, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.row == nil || r.row.Path != name {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *r.row
	return &copy, nil
}

func (r *retirementTestRepo) GetByPathHash(ctx context.Context, hash string, vault, uid int64) (*domain.File, error) {
	r.mu.Lock()
	if r.row == nil || r.row.PathHash != hash {
		r.mu.Unlock()
		return nil, gorm.ErrRecordNotFound
	}
	name := r.row.Path
	r.mu.Unlock()
	return r.GetByPath(ctx, name, vault, uid)
}

func (r *retirementTestRepo) Update(_ context.Context, row *domain.File, _ int64) (*domain.File, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := *row
	copy.UpdatedTimestamp++
	r.row = &copy
	r.writes++
	result := copy
	return &result, nil
}
func (r *retirementTestRepo) Create(ctx context.Context, row *domain.File, uid int64) (*domain.File, error) {
	return r.Update(ctx, row, uid)
}
func (r *retirementTestRepo) UpdateMtime(_ context.Context, mtime, _, _ int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mtimeWrites++
	r.row.Mtime = mtime
	return nil
}
func (r *retirementTestRepo) UpdateActionMtime(_ context.Context, action domain.FileAction, mtime, _, _ int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.actionMtimeWrites++
	r.row.Action = action
	r.row.Mtime = mtime
	return nil
}
func (r *retirementTestRepo) CountSizeSum(context.Context, int64, int64) (*domain.CountSizeResult, error) {
	return &domain.CountSizeResult{}, nil
}

type retirementTestVault struct{ VaultService }

func (retirementTestVault) MustGetID(context.Context, int64, string) (int64, error) { return 1, nil }
func (retirementTestVault) UpdateFileStats(context.Context, int64, int64, int64, int64) error {
	return nil
}

type retirementTestFolder struct{ FolderService }

func (retirementTestFolder) SyncResourceFID(context.Context, int64, int64, []int64, []int64) error {
	return nil
}
func (retirementTestFolder) CleanupEmptyAncestors(context.Context, int64, int64, string) error {
	return nil
}

func retirementFixture(t *testing.T) (*fileService, *retirementTestRepo, *dto.FileUpdateRequest, string) {
	t.Helper()
	dir := t.TempDir()
	old := []byte("retired full bytes")
	h := sha256.Sum256(old)
	name := "Images/原图.png"
	doc, _ := json.Marshal(retirement.Document{Version: 1, Entries: []retirement.Entry{{UID: 1, VaultID: 1, Path: name, SHA256: hex.EncodeToString(h[:])}}})
	policy := filepath.Join(dir, "policy.json")
	if err := os.WriteFile(policy, doc, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(retirement.EnvPolicy, policy)
	blob := filepath.Join(dir, "persisted-old")
	if err := os.WriteFile(blob, old, 0600); err != nil {
		t.Fatal(err)
	}
	upload := filepath.Join(dir, "uploaded-old")
	if err := os.WriteFile(upload, old, 0600); err != nil {
		t.Fatal(err)
	}
	repo := &retirementTestRepo{row: &domain.File{ID: 7, VaultID: 1, Path: name, PathHash: util.EncodeHash32(name), ContentHash: "same-weak-hash", SavePath: blob, Action: domain.FileActionDelete, Mtime: 1, UpdatedTimestamp: 100}}
	svc := NewFileService(nil, repo, nil, retirementTestVault{}, retirementTestFolder{}, nil, nil, nil, nil).(*fileService)
	params := &dto.FileUpdateRequest{Vault: "Files", Path: name, PathHash: util.EncodeHash32(name), ContentHash: "same-weak-hash", SavePath: upload, Mtime: 2}
	return svc, repo, params, policy
}

func TestRetirementRejectsUploadAndPersistedBlobBeforeMtimeMutation(t *testing.T) {
	for _, skip := range []bool{false, true} {
		t.Run(map[bool]string{false: "uploaded-bytes", true: "persisted-skip-upload"}[skip], func(t *testing.T) {
			svc, repo, params, _ := retirementFixture(t)
			if skip {
				params.SavePath = ""
			}
			_, _, err := svc.UploadComplete(context.Background(), 1, params)
			var rejected *RetiredBytesError
			if !errors.As(err, &rejected) || rejected.Current.Action != "delete" {
				t.Fatalf("want terminal tombstone rejection, got %v", err)
			}
			if repo.writes != 0 || repo.actionMtimeWrites != 0 || repo.row.Action != domain.FileActionDelete {
				t.Fatal("rejected old bytes mutated row")
			}
		})
	}
}

func TestRetirementNewBytesAllowedEvenWithSameWeakHash(t *testing.T) {
	svc, repo, params, _ := retirementFixture(t)
	if err := os.WriteFile(params.SavePath, []byte("intentional new bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	params.Mtime = repo.row.Mtime
	_, result, err := svc.UploadComplete(context.Background(), 1, params)
	if err != nil || result.Action != "create" || repo.writes != 1 || repo.actionMtimeWrites != 0 {
		t.Fatalf("new bytes not committed: %+v %v", result, err)
	}
	if repo.row.SavePath != params.SavePath {
		t.Fatal("new bytes discarded by metadata shortcut")
	}
}

func TestRetirementNewBytesReplaceActiveWeakHashMatch(t *testing.T) {
	svc, repo, params, _ := retirementFixture(t)
	repo.row.Action = domain.FileActionModify
	if err := os.WriteFile(params.SavePath, []byte("different actual bytes with same weak metadata"), 0600); err != nil {
		t.Fatal(err)
	}
	params.Mtime = repo.row.Mtime
	_, result, err := svc.UploadComplete(context.Background(), 1, params)
	if err != nil || result.Action != "modify" || repo.writes != 1 || repo.row.SavePath != params.SavePath {
		t.Fatalf("active weak-hash shortcut dropped valid new bytes: %+v %v", result, err)
	}
}

func TestRetirementHTTPStyleOldBytesNeverOverwriteActiveNewContent(t *testing.T) {
	svc, repo, params, _ := retirementFixture(t)
	newBlob := filepath.Join(t.TempDir(), "current-new")
	if err := os.WriteFile(newBlob, []byte("intentional current content"), 0600); err != nil {
		t.Fatal(err)
	}
	repo.row.SavePath = newBlob
	repo.row.Action = domain.FileActionModify
	repo.row.ContentHash = "current-new-hash"
	_, _, err := svc.UpdateOrCreate(context.Background(), 1, params, false)
	var rejected *RetiredBytesError
	if !errors.As(err, &rejected) || rejected.Current.Action != "modify" || rejected.Current.ContentHash != "current-new-hash" || repo.writes != 0 || repo.row.SavePath != newBlob {
		t.Fatalf("stale HTTP bytes replaced active content: %v", err)
	}
}

func TestRetirementUnlistedPathAllowed(t *testing.T) {
	svc, repo, params, _ := retirementFixture(t)
	params.Path = "Images/unlisted.png"
	params.PathHash = util.EncodeHash32(params.Path)
	repo.row.Path = params.Path
	repo.row.PathHash = params.PathHash
	_, result, err := svc.UpdateOrCreate(context.Background(), 1, params, false)
	if err != nil || result.Action != "create" || repo.writes != 1 {
		t.Fatalf("unlisted path rejected: %v", err)
	}
}

func TestRetirementUpdateCheckNeverTrustsWeakHash(t *testing.T) {
	svc, repo, params, _ := retirementFixture(t)
	repo.row.Action = domain.FileActionModify
	mode, _, err := svc.UpdateCheck(context.Background(), 1, &dto.FileUpdateCheckRequest{Vault: params.Vault, Path: params.Path, PathHash: params.PathHash, ContentHash: params.ContentHash, Mtime: 2})
	if err != nil || mode != "UpdateContent" || repo.mtimeWrites != 0 {
		t.Fatalf("weak hash allowed metadata shortcut: %s %v", mode, err)
	}
}

func TestRetirementDisabledMetadataOnlyKeepsTombstone(t *testing.T) {
	svc, repo, params, _ := retirementFixture(t)
	svc.retirement = nil
	params.SavePath = ""
	_, result, err := svc.UploadComplete(context.Background(), 1, params)
	if err != nil || result.Action != "delete" || repo.writes != 0 || repo.actionMtimeWrites != 0 {
		t.Fatalf("mtime-only revived tombstone: %+v %v", result, err)
	}
}

func TestRetirementRestoreRaceAuditAndNoPermanentExemption(t *testing.T) {
	svc, repo, params, policy := retirementFixture(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	start := make(chan struct{})
	errorsOut := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		_, err := svc.Restore(ctx, 1, &dto.FileRestoreRequest{Vault: params.Vault, Path: params.Path, PathHash: params.PathHash})
		errorsOut <- err
	}()
	go func() {
		defer wg.Done()
		<-start
		_, _, err := svc.UploadComplete(ctx, 1, params)
		var rejected *RetiredBytesError
		if !errors.As(err, &rejected) {
			errorsOut <- errors.New("old upload not rejected")
		} else {
			errorsOut <- nil
		}
	}()
	close(start)
	wg.Wait()
	close(errorsOut)
	for err := range errorsOut {
		if err != nil {
			t.Fatal(err)
		}
	}
	if repo.row.Action != domain.FileActionModify || repo.writes != 1 {
		t.Fatal("racing old upload invalidated explicit Restore")
	}
	stale := &dto.FileDTO{Action: "delete", Path: params.Path}
	if err := WithRetirementState(ctx, svc, 1, params.Vault, params.Path, stale, func(current *dto.FileDTO) {
		if current.Action != "modify" {
			t.Fatal("reconciled stale tombstone after Restore")
		}
	}); err != nil {
		t.Fatal(err)
	}
	audit, err := os.ReadFile(policy + ".audit.jsonl")
	if err != nil || !strings.Contains(string(audit), "restore_authorized") || !strings.Contains(string(audit), "restore_committed") {
		t.Fatalf("missing persistent Restore audit: %s %v", audit, err)
	}
	if _, err := svc.Delete(ctx, 1, &dto.FileDeleteRequest{Vault: params.Vault, Path: params.Path, PathHash: params.PathHash}); err != nil {
		t.Fatal(err)
	}
	_, _, err = svc.UploadComplete(ctx, 1, params)
	var rejected *RetiredBytesError
	if !errors.As(err, &rejected) || repo.row.Action != domain.FileActionDelete {
		t.Fatal("Restore created a permanent upload exemption")
	}
}

func TestRetirementDeleteCASPreservesConcurrentNewBytesAndRestoreGeneration(t *testing.T) {
	svc, repo, params, _ := retirementFixture(t)
	oldDigest, err := retirement.SHA256File(repo.row.SavePath)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate an intentional new upload after the cleanup client's precheck.
	newBlob := filepath.Join(t.TempDir(), "new-current")
	if err := os.WriteFile(newBlob, []byte("new content after approval"), 0600); err != nil {
		t.Fatal(err)
	}
	repo.row.SavePath = newBlob
	repo.row.Action = domain.FileActionModify
	repo.row.UpdatedTimestamp = 200
	deleteParams := &dto.FileDeleteRequest{Vault: params.Vault, Path: params.Path, PathHash: params.PathHash, ExpectedSHA256: oldDigest}
	_, err = svc.Delete(context.Background(), 1, deleteParams)
	var conflict *code.Code
	if !errors.As(err, &conflict) || conflict.Code() != 530 || repo.writes != 0 || repo.row.Action != domain.FileActionModify {
		t.Fatalf("SHA CAS deleted newer bytes: %v", err)
	}
	deleteParams.ExpectedSHA256, err = retirement.SHA256File(newBlob)
	if err != nil {
		t.Fatal(err)
	}
	deleteParams.ExpectedUpdatedTimestamp = 100
	_, err = svc.Delete(context.Background(), 1, deleteParams)
	if !errors.As(err, &conflict) || repo.writes != 0 {
		t.Fatal("generation CAS deleted a newer Restore/upload generation")
	}
	deleteParams.ExpectedUpdatedTimestamp = 200
	result, err := svc.Delete(context.Background(), 1, deleteParams)
	if err != nil || result.Action != "delete" || repo.writes != 1 {
		t.Fatalf("matching CAS refused authorized delete: %v", err)
	}
}

func TestRetirementRenameCannotReturnOldBytesToRetiredDestination(t *testing.T) {
	svc, repo, params, _ := retirementFixture(t)
	repo.row.Path = "Images/source.png"
	repo.row.PathHash = util.EncodeHash32(repo.row.Path)
	repo.row.Action = domain.FileActionModify
	_, _, err := svc.Rename(context.Background(), 1, &dto.FileRenameRequest{Vault: params.Vault, OldPath: repo.row.Path, OldPathHash: repo.row.PathHash, Path: params.Path, PathHash: params.PathHash})
	var rejected *RetiredBytesError
	if !errors.As(err, &rejected) || repo.writes != 0 || repo.row.Action != domain.FileActionModify {
		t.Fatalf("rename bypassed retirement: %v", err)
	}
}

func TestRetirementMissingRecordStillReturnsSyntheticTombstone(t *testing.T) {
	svc, repo, params, _ := retirementFixture(t)
	repo.row = nil
	_, _, err := svc.UploadComplete(context.Background(), 1, params)
	var rejected *RetiredBytesError
	if !errors.As(err, &rejected) || rejected.Current.Path != params.Path || rejected.Current.Action != "delete" || rejected.Current.UpdatedTimestamp != 0 || repo.writes != 0 {
		t.Fatalf("missing-row retirement failed: %v", err)
	}
}
