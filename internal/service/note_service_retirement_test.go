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
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/retirement"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"gorm.io/gorm"
)

type noteRetirementRepo struct {
	domain.NoteRepository
	mu                                     sync.Mutex
	row                                    *domain.Note
	extraRow                               *domain.Note
	writes, mtimeWrites, actionMtimeWrites int
	readError                              error
}

func (r *noteRetirementRepo) GetByPath(_ context.Context, path string, _, _ int64) (*domain.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.readError != nil {
		return nil, r.readError
	}
	if r.extraRow != nil && r.extraRow.Path == path {
		copy := *r.extraRow
		return &copy, nil
	}
	if r.row == nil || r.row.Path != path {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *r.row
	return &copy, nil
}
func (r *noteRetirementRepo) GetAllByPathHash(ctx context.Context, hash string, vault, uid int64) (*domain.Note, error) {
	r.mu.Lock()
	if r.extraRow != nil && r.extraRow.PathHash == hash {
		path := r.extraRow.Path
		r.mu.Unlock()
		return r.GetByPath(ctx, path, vault, uid)
	}
	if r.row == nil || r.row.PathHash != hash {
		r.mu.Unlock()
		return nil, gorm.ErrRecordNotFound
	}
	path := r.row.Path
	r.mu.Unlock()
	return r.GetByPath(ctx, path, vault, uid)
}
func (r *noteRetirementRepo) GetByID(ctx context.Context, _ int64, uid int64) (*domain.Note, error) {
	r.mu.Lock()
	path := r.row.Path
	r.mu.Unlock()
	return r.GetByPath(ctx, path, 1, uid)
}
func (r *noteRetirementRepo) GetByPathHash(ctx context.Context, hash string, vault, uid int64) (*domain.Note, error) {
	return r.GetAllByPathHash(ctx, hash, vault, uid)
}
func (r *noteRetirementRepo) GetByPathHashIncludeRecycle(ctx context.Context, hash string, vault, uid int64, _ bool) (*domain.Note, error) {
	return r.GetAllByPathHash(ctx, hash, vault, uid)
}
func (r *noteRetirementRepo) Update(_ context.Context, note *domain.Note, _ int64) (*domain.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := *note
	copy.UpdatedTimestamp++
	r.row = &copy
	r.writes++
	result := copy
	return &result, nil
}
func (r *noteRetirementRepo) Create(ctx context.Context, note *domain.Note, uid int64) (*domain.Note, error) {
	return r.Update(ctx, note, uid)
}
func (r *noteRetirementRepo) UpdateDelete(ctx context.Context, note *domain.Note, uid int64) error {
	updated, err := r.Update(ctx, note, uid)
	if err == nil {
		note.UpdatedTimestamp = updated.UpdatedTimestamp
	}
	return err
}
func (r *noteRetirementRepo) UpdateMtime(_ context.Context, mtime, _, _ int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mtimeWrites++
	r.row.Mtime = mtime
	return nil
}
func (r *noteRetirementRepo) UpdateActionMtime(_ context.Context, action domain.NoteAction, mtime, _, _ int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.actionMtimeWrites++
	r.row.Action = action
	r.row.Mtime = mtime
	return nil
}
func (r *noteRetirementRepo) CountSizeSum(context.Context, int64, int64) (*domain.CountSizeResult, error) {
	return &domain.CountSizeResult{}, nil
}

type noteRetirementVault struct{ VaultService }

func (noteRetirementVault) MustGetID(context.Context, int64, string) (int64, error) { return 1, nil }
func (noteRetirementVault) UpdateNoteStats(context.Context, int64, int64, int64, int64) error {
	return nil
}

type noteRetirementFolder struct{ FolderService }

func (noteRetirementFolder) SyncResourceFID(context.Context, int64, int64, []int64, []int64) error {
	return nil
}

type noteRetirementShares struct{ domain.UserShareRepository }

func (noteRetirementShares) UpdateStatusByRes(context.Context, int64, string, int64, int64) error {
	return nil
}

func noteRetirementFixture(t *testing.T) (*noteService, *noteRetirementRepo, *dto.NoteModifyOrCreateRequest, string) {
	t.Helper()
	path := "课程/已搬移笔记.md"
	content := "原始 UTF-8 内容 😀\r\n![[旧附件.png]]"
	digest := sha256.Sum256([]byte(content))
	policy := filepath.Join(t.TempDir(), "policy.json")
	doc, _ := json.Marshal(retirement.Document{Version: 1, Entries: []retirement.Entry{{UID: 1, VaultID: 1, Path: path, SHA256: hex.EncodeToString(digest[:])}}})
	if err := os.WriteFile(policy, doc, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(retirement.EnvPolicy, policy)
	repo := &noteRetirementRepo{row: &domain.Note{ID: 7, VaultID: 1, Path: path, PathHash: util.EncodeHash32(path), Content: content, ContentHash: util.EncodeHash32(content), Action: domain.NoteActionDelete, Mtime: 10, Ctime: 5, UpdatedTimestamp: 100}}
	svc := NewNoteService(nil, repo, nil, nil, noteRetirementShares{}, nil, noteRetirementVault{}, noteRetirementFolder{}, nil, nil, nil, nil).(*noteService)
	params := &dto.NoteModifyOrCreateRequest{Vault: "Files", Path: path, PathHash: util.EncodeHash32(path), Content: content, ContentHash: util.EncodeHash32(content), Mtime: 999, Ctime: 5}
	return svc, repo, params, policy
}

func TestNoteRetirementRejectsActualBytesBeforeAnyMetadataOrContentWrite(t *testing.T) {
	for _, http := range []bool{false, true} {
		t.Run(map[bool]string{false: "ws", true: "http-save"}[http], func(t *testing.T) {
			svc, repo, params, _ := noteRetirementFixture(t)
			params.ContentHash = "forged-weak-hash"
			_, _, err := svc.ModifyOrCreate(context.Background(), 1, params, !http)
			var retired *RetiredNoteBytesError
			if !errors.As(err, &retired) || retired.Current.Action != "delete" {
				t.Fatalf("not rejected: %v", err)
			}
			if repo.writes != 0 || repo.actionMtimeWrites != 0 || repo.mtimeWrites != 0 {
				t.Fatal("retired bytes mutated database")
			}
		})
	}
}

func TestNoteRetirementRejectsBeforeMergeAndCannotTrustPrefetchedState(t *testing.T) {
	svc, repo, params, _ := noteRetirementFixture(t)
	protected, err := CheckNoteRetirement(context.Background(), svc, 1, params)
	var retired *RetiredNoteBytesError
	if !protected || !errors.As(err, &retired) {
		t.Fatalf("early WS guard missed old bytes: %v", err)
	}
	stale := *repo.row
	stale.Action = domain.NoteActionModify
	params.Content = "intentional new content"
	// Same weak hash / mtime must not reuse the stale pre-Delete snapshot.
	params.Mtime = stale.Mtime
	_, result, err := svc.ModifyOrCreate(context.Background(), 1, params, true, &stale)
	if err != nil || result.Action != "create" || repo.writes != 1 || repo.row.Content != params.Content {
		t.Fatalf("fresh write lost: %+v %v", result, err)
	}
	if repo.row.ContentHash != util.EncodeHash32(params.Content) {
		t.Fatal("stored weak hash does not reflect actual new bytes")
	}
}

func TestNoteRetirementMetadataOnlyNeverReactivatesDeletedNote(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "disabled", true: "enabled"}[enabled], func(t *testing.T) {
			svc, repo, params, _ := noteRetirementFixture(t)
			if !enabled {
				svc.retirement = nil
			}
			params.Content = ""
			_, result, err := svc.ModifyOrCreate(context.Background(), 1, params, true)
			if enabled {
				var retired *RetiredNoteBytesError
				if !errors.As(err, &retired) {
					t.Fatalf("metadata not acknowledged rejection: %v", err)
				}
			} else if err != nil || result.Action != "delete" {
				t.Fatalf("metadata path: %+v %v", result, err)
			}
			if repo.writes != 0 || repo.actionMtimeWrites != 0 || repo.row.Action != domain.NoteActionDelete {
				t.Fatal("metadata reactivated tombstone")
			}
		})
	}
}

func TestNoteRetirementAllowsDifferentBytesAndExplicitEmptyDocument(t *testing.T) {
	for _, content := range []string{"intentional new bytes 😀", ""} {
		t.Run(content, func(t *testing.T) {
			svc, repo, params, _ := noteRetirementFixture(t)
			params.Content = content
			if content == "" {
				params.ContentHash = "0"
			}
			_, result, err := svc.ModifyOrCreate(context.Background(), 1, params, true)
			if err != nil || result.Action != "create" || repo.row.Content != content || repo.writes != 1 {
				t.Fatalf("new bytes blocked: %+v %v", result, err)
			}
		})
	}
}

func TestNoteRetirementUpdateCheckDoesNotTrustWeakHashOrMutateMtime(t *testing.T) {
	svc, repo, params, _ := noteRetirementFixture(t)
	repo.row.Action = domain.NoteActionModify
	mode, _, err := svc.UpdateCheck(context.Background(), 1, &dto.NoteUpdateCheckRequest{Vault: params.Vault, Path: params.Path, PathHash: params.PathHash, ContentHash: params.ContentHash, Mtime: 999})
	if err != nil || mode != "UpdateContent" || repo.mtimeWrites != 0 {
		t.Fatalf("unsafe metadata shortcut: %s %v", mode, err)
	}
}

func TestNoteRetirementUnlistedScopeKeepsNormalSaveBehavior(t *testing.T) {
	for _, scope := range []string{"path", "uid", "disabled"} {
		t.Run(scope, func(t *testing.T) {
			svc, repo, params, _ := noteRetirementFixture(t)
			uid := int64(1)
			if scope == "path" {
				params.Path = "different.md"
				params.PathHash = util.EncodeHash32(params.Path)
				repo.row.Path = params.Path
				repo.row.PathHash = params.PathHash
			}
			if scope == "uid" {
				uid = 2
			}
			if scope == "disabled" {
				svc.retirement = nil
			}
			_, result, err := svc.ModifyOrCreate(context.Background(), uid, params, false)
			if err != nil || result.Action != "create" || repo.writes != 1 {
				t.Fatalf("unlisted scope denied: %v", err)
			}
		})
	}
}

func TestNoteRetirementRestoreRaceAuditDeleteAndNoPermanentExemption(t *testing.T) {
	svc, repo, params, policy := noteRetirementFixture(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	start := make(chan struct{})
	errs := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		_, err := svc.WithClient("desktop", "explicit user", "2.4.0").Restore(ctx, 1, &dto.NoteRestoreRequest{Vault: params.Vault, Path: params.Path, PathHash: params.PathHash})
		errs <- err
	}()
	go func() {
		defer wg.Done()
		<-start
		_, _, err := svc.WithClient("mobile", "old device", "2.4.0").ModifyOrCreate(ctx, 1, params, true)
		var retired *RetiredNoteBytesError
		if !errors.As(err, &retired) {
			errs <- errors.New("old upload not rejected")
		} else {
			errs <- nil
		}
	}()
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if repo.row.Action != domain.NoteActionModify || repo.writes != 1 {
		t.Fatal("Restore was overwritten by concurrent upload")
	}
	audit, err := os.ReadFile(policy + ".audit.jsonl")
	if err != nil || !strings.Contains(string(audit), "note_restore_authorized") || !strings.Contains(string(audit), "note_restore_committed") {
		t.Fatalf("durable audit missing: %s %v", audit, err)
	}
	_, err = svc.Delete(ctx, 1, &dto.NoteDeleteRequest{Vault: params.Vault, Path: params.Path, PathHash: params.PathHash})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = svc.ModifyOrCreate(ctx, 1, params, true)
	var retired *RetiredNoteBytesError
	if !errors.As(err, &retired) || repo.row.Action != domain.NoteActionDelete || repo.writes != 2 {
		t.Fatalf("Restore left permanent exemption: %v", err)
	}
}

func TestNoteRetirementRestoreAuditFailurePreventsMutation(t *testing.T) {
	svc, repo, params, policy := noteRetirementFixture(t)
	if err := os.Mkdir(policy+".audit.jsonl", 0700); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Restore(context.Background(), 1, &dto.NoteRestoreRequest{Vault: params.Vault, Path: params.Path, PathHash: params.PathHash})
	if err == nil || repo.writes != 0 || repo.row.Action != domain.NoteActionDelete {
		t.Fatalf("unaudited Restore committed: %v", err)
	}
}

func TestNoteRetirementResponseReReadsCurrentActiveInsteadOfStaleTombstone(t *testing.T) {
	svc, repo, params, _ := noteRetirementFixture(t)
	ctx := context.Background()
	_, _, err := svc.ModifyOrCreate(ctx, 1, params, true)
	var retired *RetiredNoteBytesError
	if !errors.As(err, &retired) {
		t.Fatal(err)
	}
	params.Content = "new paragraphs created after rejection"
	_, _, err = svc.ModifyOrCreate(ctx, 1, params, true)
	if err != nil {
		t.Fatal(err)
	}
	err = WithNoteRetirementState(ctx, svc, 1, params.Vault, params.Path, retired.Current, func(current *dto.NoteDTO) {
		if current.Action == "delete" || current.Content != params.Content {
			t.Fatal("late stale tombstone would erase current active content")
		}
	})
	if err != nil || repo.writes != 1 {
		t.Fatal(err)
	}
}

func TestNoteRetirementResponseReadFailureNeverDeliversStaleDeletion(t *testing.T) {
	svc, repo, params, _ := noteRetirementFixture(t)
	repo.readError = errors.New("storage unavailable")
	delivered := false
	err := WithNoteRetirementState(context.Background(), svc, 1, params.Vault, params.Path, &dto.NoteDTO{Action: "delete"}, func(*dto.NoteDTO) { delivered = true })
	if err == nil || delivered {
		t.Fatal("failed authoritative lookup emitted stale delete")
	}
}

func TestNoteRetirementRejectsPathHashSpoofing(t *testing.T) {
	svc, repo, params, _ := noteRetirementFixture(t)
	params.PathHash = "spoofed"
	params.Content = "new bytes"
	_, _, err := svc.ModifyOrCreate(context.Background(), 1, params, false)
	if err == nil || repo.writes != 0 {
		t.Fatal("protected exact path/hash mismatch accepted")
	}
}

func TestNoteRetirementKnownOldBytesAckableEvenWhenStateReadFails(t *testing.T) {
	svc, repo, params, _ := noteRetirementFixture(t)
	repo.readError = errors.New("storage unavailable")
	protected, err := CheckNoteRetirement(context.Background(), svc, 1, params)
	var retired *RetiredNoteBytesError
	if !protected || !errors.As(err, &retired) || retired.Current != nil {
		t.Fatalf("old bytes were not terminally rejected without a stale state: %v", err)
	}
	_, _, err = svc.ModifyOrCreate(context.Background(), 1, params, false)
	if !errors.As(err, &retired) || retired.Current != nil || repo.writes != 0 {
		t.Fatalf("failed read admitted bytes or supplied an unsafe deletion: %v", err)
	}
}

func TestNoteRetirementRenameCannotBypassRetiredTarget(t *testing.T) {
	svc, repo, params, _ := noteRetirementFixture(t)
	source := *repo.row
	source.ID = 8
	source.Path = "elsewhere.md"
	source.PathHash = util.EncodeHash32(source.Path)
	source.Action = domain.NoteActionModify
	repo.extraRow = &source
	_, _, err := svc.Rename(context.Background(), 1, &dto.NoteRenameRequest{Vault: params.Vault, Path: params.Path, PathHash: params.PathHash, OldPath: source.Path, OldPathHash: source.PathHash})
	var retired *RetiredNoteBytesError
	if !errors.As(err, &retired) || repo.writes != 0 || repo.row.Action != domain.NoteActionDelete || repo.extraRow.Action != domain.NoteActionModify {
		t.Fatalf("Rename activated retired target or deleted its source: %v", err)
	}
}

func TestNoteRetirementDeleteCASUsesRealBytesAndGeneration(t *testing.T) {
	for _, scenario := range []string{"new-content", "new-generation", "match", "legacy-no-cas"} {
		t.Run(scenario, func(t *testing.T) {
			svc, repo, params, _ := noteRetirementFixture(t)
			svc.retirement = nil // CAS also protects duplicate notes outside the policy.
			repo.row.Action = domain.NoteActionModify
			digest := sha256.Sum256([]byte(params.Content))
			request := &dto.NoteDeleteRequest{Vault: params.Vault, Path: params.Path, PathHash: params.PathHash, ExpectedSHA256: strings.ToUpper(hex.EncodeToString(digest[:])), ExpectedUpdatedTimestamp: 100}
			if scenario == "new-content" {
				repo.row.Content = "new server paragraphs"
			} // weak hash deliberately unchanged
			if scenario == "new-generation" {
				repo.row.UpdatedTimestamp = 101
			}
			if scenario == "legacy-no-cas" {
				request.ExpectedSHA256 = ""
				request.ExpectedUpdatedTimestamp = 0
			}
			result, err := svc.Delete(context.Background(), 1, request)
			if scenario == "new-content" || scenario == "new-generation" {
				var conflict *code.Code
				if !errors.As(err, &conflict) || conflict.Code() != 530 || repo.writes != 0 || repo.row.Action != domain.NoteActionModify {
					t.Fatalf("CAS erased concurrent data: %v", err)
				}
				current := conflict.Data().(*dto.NoteDTO)
				if current.Content != repo.row.Content || current.UpdatedTimestamp != repo.row.UpdatedTimestamp {
					t.Fatal("530 lacks current note state")
				}
			} else if err != nil || result.Action != "delete" || repo.writes != 1 {
				t.Fatalf("matching / compatible Delete failed: %+v %v", result, err)
			}
		})
	}
}

func TestNoteRetirementStateLockCoversResponseEnqueue(t *testing.T) {
	svc, repo, params, _ := noteRetirementFixture(t)
	writerStarted := make(chan struct{})
	writerDone := make(chan error, 1)
	queued := false
	err := WithNoteRetirementState(context.Background(), svc, 1, params.Vault, params.Path, nil, func(current *dto.NoteDTO) {
		if current.Action != "delete" {
			t.Fatal("wrong initial state")
		}
		go func() {
			close(writerStarted)
			fresh := *params
			fresh.Content = "new data while reply enqueues"
			_, _, err := svc.WithClient("desktop", "concurrent writer", "2.4.0").ModifyOrCreate(context.Background(), 1, &fresh, true)
			writerDone <- err
		}()
		<-writerStarted
		select {
		case err := <-writerDone:
			t.Fatalf("writer committed before current-state response enqueue: %v", err)
		case <-time.After(100 * time.Millisecond):
		}
		queued = true // Represents the synchronous ToResponse enqueue boundary.
	})
	if err != nil || !queued {
		t.Fatal(err)
	}
	if err := <-writerDone; err != nil {
		t.Fatal(err)
	}
	if repo.row.Action == domain.NoteActionDelete || repo.row.Content != "new data while reply enqueues" {
		t.Fatal("writer did not proceed after reply enqueue")
	}
}
