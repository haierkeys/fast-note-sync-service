package websocket_router

import (
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

func TestRetirementUploadAckReleasesSessionBeforeDeleteReconciliation(t *testing.T) {
	temp := filepath.Join(t.TempDir(), "rejected-upload")
	if err := os.WriteFile(temp, []byte("retired bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	cancelled := false
	session := &FileUploadBinaryChunkSession{ID: "session-1", Vault: "Files", Path: "Images/old.png", PathHash: "path-hash", Context: "ctx-1", SavePath: temp, CancelFunc: func() { cancelled = true }}
	current := &dto.FileDTO{Action: "delete", Path: session.Path, PathHash: session.PathHash, Ctime: 10, Mtime: 20, Size: 30, UpdatedTimestamp: 40}
	var sequence []string
	send := func(response *code.Code, actions ...string) {
		sequence = append(sequence, actions[0])
		if response.Vault() != "Files" || !response.Status() {
			t.Fatal("missing successful vault response")
		}
		if actions[0] == string(FileUploadAck) {
			ack := response.Data().(dto.FileUploadAckMessage)
			if ack.Path != session.Path || ack.PathHash != session.PathHash || ack.LastTime != 40 || response.Context() != "ctx-1" {
				t.Fatalf("bad pending-upload release ack: %+v", ack)
			}
		} else {
			message := response.Data().(dto.FileSyncDeleteMessage)
			if message.Mtime != 20 || message.UpdatedTimestamp != 40 || message.PathHash != "path-hash" {
				t.Fatal("tombstone metadata lost")
			}
			if !cancelled || session.SavePath != "" {
				t.Fatal("reconciliation preceded cleanup")
			}
		}
	}
	finishRetiredUpload(session, current, send, func() {
		if !session.isCompleted {
			t.Fatal("session removal preceded completion")
		}
		sequence = append(sequence, "RemoveSession")
	})
	if want := []string{string(FileUploadAck), "RemoveSession", string(FileSyncDelete)}; !reflect.DeepEqual(sequence, want) {
		t.Fatalf("completion order %v, want %v", sequence, want)
	}
	if _, err := os.Stat(temp); !os.IsNotExist(err) {
		t.Fatal("rejected temporary bytes retained")
	}
}

func TestRetirementUploadPreservesCurrentActiveNewContent(t *testing.T) {
	session := &FileUploadBinaryChunkSession{Vault: "Files", Path: "Images/old.png", PathHash: "path-hash", ContentHash: "rejected-old-hash"}
	current := &dto.FileDTO{Action: "modify", Path: session.Path, PathHash: session.PathHash, ContentHash: "current-new-hash", UpdatedTimestamp: 50}
	var sequence []string
	finishRetiredUpload(session, current, func(response *code.Code, actions ...string) {
		sequence = append(sequence, actions[0])
		if actions[0] == string(FileSyncUpdate) && response.Data().(dto.FileSyncModifyMessage).ContentHash != "current-new-hash" {
			t.Fatal("returned rejected bytes instead of active content")
		}
	}, func() { sequence = append(sequence, "RemoveSession") })
	if want := []string{string(FileUploadAck), "RemoveSession", string(FileSyncUpdate)}; !reflect.DeepEqual(sequence, want) {
		t.Fatalf("active reconciliation %v", sequence)
	}
}

func TestRetirementUploadStateReadFailureStillReleasesSlot(t *testing.T) {
	session := &FileUploadBinaryChunkSession{Vault: "Files", Path: "Images/old.png"}
	var sequence []string
	finishRetiredUpload(session, nil, func(_ *code.Code, actions ...string) { sequence = append(sequence, actions[0]) }, func() { sequence = append(sequence, "RemoveSession") })
	if want := []string{string(FileUploadAck), "RemoveSession"}; !reflect.DeepEqual(sequence, want) || !session.isCompleted {
		t.Fatalf("slot not released: %v", sequence)
	}
}

func TestRetirementUploadCompletionClaimIsUniqueAndStable(t *testing.T) {
	session := &FileUploadBinaryChunkSession{Size: 10, TotalChunks: 2, uploadedChunks: map[uint32]struct{}{}}
	if _, _, _, complete := claimFileUploadCompletion(session, 0, 5); complete {
		t.Fatal("first chunk finalized incomplete bytes")
	}
	var claims atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, _, _, complete := claimFileUploadCompletion(session, 1, 5); complete {
				claims.Add(1)
			}
		}()
	}
	wg.Wait()
	if claims.Load() != 1 || session.UploadedBytes != 10 || session.UploadedChunks != 2 || !session.isFinalizing {
		t.Fatalf("unstable completion: claims=%d bytes=%d chunks=%d", claims.Load(), session.UploadedBytes, session.UploadedChunks)
	}
}
