package websocket_router

import (
	"encoding/json"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

// TestA1_ErrorSyncConflict_NotifyPayload is the A1 wiring regression: NoteModify's
// conflict branch (mergeResult.HasConflict || baseHashNotFound) now fires
// code.ErrorSyncConflict.WithData(dto.NoteSyncNeedPushMessage{...}) at the triggering client.
// This reconstructs the exact Res envelope WebsocketClient.ToResponse would send, so a
// regression (wrong code, dropped path, dropped vault) is caught without needing a live WS
// connection — the client's ERROR_SYNC_CONFLICT branch (websocket_manager.ts) keys off
// code===530 and reads data.Path.
func TestA1_ErrorSyncConflict_NotifyPayload(t *testing.T) {
	if got := code.ErrorSyncConflict.Code(); got != 530 {
		t.Fatalf("ErrorSyncConflict code = %d, want 530 (client ERROR_SYNC_CONFLICT)", got)
	}

	notify := code.ErrorSyncConflict.WithData(dto.NoteSyncNeedPushMessage{
		Path:     "Notes/ReadMe.md",
		PathHash: "hash123",
	}).WithVault("MyVault").WithContext("ctx-1")

	// Mirrors the content struct built in WebsocketClient.ToResponse.
	content := app.Res{
		Code:    notify.Code(),
		Status:  notify.Status(),
		Message: notify.MsgIn("zh_cn"),
		Data:    notify.Data(),
	}
	if notify.HaveVault() {
		content.Vault = notify.Vault()
	}
	if notify.HaveContext() {
		content.Context = notify.Context()
	}

	raw, err := json.Marshal(content)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if decoded["code"] != float64(530) {
		t.Fatalf("decoded code = %v, want 530", decoded["code"])
	}
	if decoded["status"] != false {
		t.Fatalf("decoded status = %v, want false (error response)", decoded["status"])
	}

	data, ok := decoded["data"].(map[string]any)
	if !ok {
		t.Fatalf("decoded data missing or wrong shape: %v", decoded["data"])
	}
	if data["path"] != "Notes/ReadMe.md" {
		t.Fatalf("decoded data.path = %v, want Notes/ReadMe.md", data["path"])
	}
	if data["pathHash"] != "hash123" {
		t.Fatalf("decoded data.pathHash = %v, want hash123", data["pathHash"])
	}
	if decoded["vault"] != "MyVault" {
		t.Fatalf("decoded vault = %v, want MyVault", decoded["vault"])
	}
	if decoded["context"] != "ctx-1" {
		t.Fatalf("decoded context = %v, want ctx-1", decoded["context"])
	}
}
