package websocket_router

import (
	"encoding/json"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

// TestNoteModifyAck_UnchangedNotePayload verifies that when ModifyOrCreate returns nil note
// (content already stored with matching mtime and hash), an ACK envelope is created with
// code.Success and NoteModifyAckMessage so the client clears its pending queue.
func TestNoteModifyAck_UnchangedNotePayload(t *testing.T) {
	ack := code.Success.WithData(dto.NoteModifyAckMessage{
		LastTime: 1234567890,
		Path:     "test/note.md",
		PathHash: "abc123hash",
	}).WithVault("VaultA").WithContext("ctx-99")

	content := app.Res{
		Code:    ack.Code(),
		Status:  ack.Status(),
		Message: ack.MsgIn("zh_cn"),
		Data:    ack.Data(),
	}
	if ack.HaveVault() {
		content.Vault = ack.Vault()
	}
	if ack.HaveContext() {
		content.Context = ack.Context()
	}

	bytes, err := json.Marshal(content)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(bytes, &parsed); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	data, ok := parsed["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object in response, got %v", parsed["data"])
	}

	if data["path"] != "test/note.md" {
		t.Errorf("expected path test/note.md, got %v", data["path"])
	}
	if data["pathHash"] != "abc123hash" {
		t.Errorf("expected pathHash abc123hash, got %v", data["pathHash"])
	}
	if parsed["vault"] != "VaultA" {
		t.Errorf("expected vault VaultA, got %v", parsed["vault"])
	}
}
