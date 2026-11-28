package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/domain/mocks"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

// TestNoteConflictData_TargetRowGone covers finding N1 from the targeted review: when the row a
// conditional write addressed no longer exists, the conflict payload must report "row gone"
// (revision 0) and must NOT fall back to a pathHash lookup that could describe another row.
func TestNoteConflictData_TargetRowGone(t *testing.T) {
	const (
		uid      = int64(1)
		vaultID  = int64(1)
		targetID = int64(42)
		pathHash = "-1572999001"
	)

	repo := &mocks.MockNoteRepository{}
	// the targeted row is gone…
	repo.On("GetByID", mock.Anything, targetID, uid).
		Return(nil, errors.New("record not found"))
	// …while a different row (rename tombstone / historical duplicate) still matches the path.
	// It must never appear in the payload, so the lookup must not even be attempted.
	repo.On("GetAllByPathHash", mock.Anything, pathHash, vaultID, uid).
		Return(&domain.Note{
			ID:          7,
			Path:        "Projects/Someone-Elses.md",
			PathHash:    pathHash,
			Content:     "secret of another row\n",
			ContentHash: "other-hash",
			Rev:         99,
			Action:      domain.NoteActionDelete,
		}, nil)

	svc := &noteService{noteRepo: repo}
	data := svc.noteConflictData(context.Background(), uid, vaultID, targetID, pathHash, nil)

	if data.Revision != 0 {
		t.Fatalf("row-gone payload must carry revision 0, got %d", data.Revision)
	}
	if data.Path != "" || data.Content != "" || data.ContentHash != "" || data.Mtime != 0 || data.Deleted {
		t.Fatalf("row-gone payload must disclose nothing about any row, got %+v", data)
	}
	repo.AssertNotCalled(t, "GetAllByPathHash", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	// the code contract is unchanged: a row-gone conflict is still 452
	env := svc.noteConflictCode(context.Background(), uid, vaultID, targetID, pathHash, nil)
	if env.Code() != code.ErrorNoteRevisionConflict.Code() {
		t.Fatalf("row-gone conflict must stay code 452, got %d", env.Code())
	}
	if payload, ok := env.Data().(*dto.NoteConflictData); !ok || payload.Revision != 0 || payload.Content != "" {
		t.Fatalf("452 payload must be the row-gone payload, got %#v", env.Data())
	}
}

// TestNoteConflictData_UsesTargetRowByID: a live target row is described from its primary-key read,
// never from the (ambiguous) path lookup.
func TestNoteConflictData_UsesTargetRowByID(t *testing.T) {
	repo := &mocks.MockNoteRepository{}
	repo.On("GetByID", mock.Anything, int64(42), int64(1)).
		Return(&domain.Note{ID: 42, Path: "Projects/Live.md", Content: "live body\n",
			ContentHash: "live-hash", Rev: 5, Mtime: 1234, Action: domain.NoteActionModify}, nil)

	svc := &noteService{noteRepo: repo}
	data := svc.noteConflictData(context.Background(), 1, 1, 42, "-1572999001", nil)

	if data.Path != "Projects/Live.md" || data.Revision != 5 || data.ContentHash != "live-hash" ||
		data.Mtime != 1234 || data.Deleted || data.Content != "live body\n" {
		t.Fatalf("live-row payload mismatch: %+v", data)
	}
	repo.AssertNotCalled(t, "GetAllByPathHash", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

// TestNoteConflictData_PathFallbackOnlyWithoutTargetRow documents the one authorised use of the
// path lookup: an operation with no identified target row (noteID == 0).
func TestNoteConflictData_PathFallbackOnlyWithoutTargetRow(t *testing.T) {
	repo := &mocks.MockNoteRepository{}
	repo.On("GetAllByPathHash", mock.Anything, "-1572999001", int64(1), int64(1)).
		Return(&domain.Note{ID: 3, Path: "Projects/Created-Elsewhere.md", Content: "body\n",
			ContentHash: "h", Rev: 2, Action: domain.NoteActionModify}, nil)

	svc := &noteService{noteRepo: repo}
	data := svc.noteConflictData(context.Background(), 1, 1, 0, "-1572999001", nil)

	if data.Revision != 2 || data.Path != "Projects/Created-Elsewhere.md" || data.Content != "body\n" {
		t.Fatalf("path fallback (noteID == 0) mismatch: %+v", data)
	}
	repo.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything, mock.Anything)
}

// TestNoteConflictData_NoRowAtAll: without a target id and without a matching path, the payload is
// the same row-gone shape.
func TestNoteConflictData_NoRowAtAll(t *testing.T) {
	repo := &mocks.MockNoteRepository{}
	repo.On("GetAllByPathHash", mock.Anything, "nope", int64(1), int64(1)).
		Return(nil, errors.New("record not found"))

	svc := &noteService{noteRepo: repo}
	data := svc.noteConflictData(context.Background(), 1, 1, 0, "nope", nil)
	if data.Revision != 0 || data.Content != "" || data.Path != "" {
		t.Fatalf("expected row-gone payload, got %+v", data)
	}
}
