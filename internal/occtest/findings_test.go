package occtest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
)

// TestTombstoneConflictOmitsContent is finding F2: codes 453 (and 452) must never echo the body
// of a soft-deleted note, while keeping the metadata a client needs.
func TestTombstoneConflictOmitsContent(t *testing.T) {
	i := newInstance(t)
	path := "Projects/Tomb.md"
	base := i.seedNote(path, "SECRET tombstone body 12345\n")

	// delete with the current revision
	if env := i.do("DELETE", "/api/note?vault="+testVault+"&path="+urlQueryEscape(path)+
		"&revision="+fmt.Sprint(base), nil, i.apiToken, "occapi"); env.Code != 1 {
		t.Fatalf("delete: code=%d %s", env.Code, env.Message)
	}

	// conditional content write against the tombstone
	env := i.write(path, "try to revive\n", int64p(base))
	if env.Code != 453 {
		t.Fatalf("content write to a tombstone must be 453, got code=%d %s", env.Code, env.Message)
	}
	if !strings.Contains(env.Message, "http=409") {
		t.Fatalf("453 must use HTTP 409, got %s", env.Message)
	}
	var payload struct {
		Path        string `json:"path"`
		Revision    int64  `json:"revision"`
		ContentHash string `json:"contentHash"`
		Mtime       int64  `json:"mtime"`
		Deleted     bool   `json:"deleted"`
		Content     string `json:"content"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode 453 payload: %v", err)
	}
	if !payload.Deleted {
		t.Fatalf("453 payload must report deleted=true, got %+v", payload)
	}
	if payload.Content != "" {
		t.Fatalf("453 payload must not echo the tombstone body, got %q", payload.Content)
	}
	if payload.Path != path || payload.Revision == 0 || payload.ContentHash == "" || payload.Mtime == 0 {
		t.Fatalf("453 payload must keep path/revision/contentHash/mtime, got %+v", payload)
	}

	// legacy (no revision) writes get the same treatment
	env = i.write(path, "legacy revive\n", nil)
	if env.Code != 453 {
		t.Fatalf("legacy write to a tombstone must be 453, got code=%d %s", env.Code, env.Message)
	}
	if err := json.Unmarshal(env.Data, &payload); err == nil && payload.Content != "" {
		t.Fatalf("legacy 453 payload must not echo the tombstone body, got %q", payload.Content)
	}

	// the body is still there once the note is explicitly restored
	var page struct {
		List []struct {
			Path     string `json:"path"`
			Revision int64  `json:"revision"`
		} `json:"list"`
	}
	list := i.do("GET", "/api/notes?vault="+testVault+"&isRecycle=true&page=1&pageSize=20", nil, i.apiToken, "occapi")
	if err := json.Unmarshal(list.Data, &page); err != nil {
		t.Fatalf("decode recycle list: %v", err)
	}
	var tombRev int64
	for _, n := range page.List {
		if n.Path == path {
			tombRev = n.Revision
		}
	}
	if env := i.do("PUT", "/api/note/restore", map[string]any{
		"vault": testVault, "path": path, "revision": tombRev,
	}, i.apiToken, "occapi"); env.Code != 1 {
		t.Fatalf("restore: code=%d %s", env.Code, env.Message)
	}
	if got := i.getNote(path); got.Content != "SECRET tombstone body 12345\n" {
		t.Fatalf("restore must return the preserved body, got %q", got.Content)
	}
}

// TestLiveConflictStillReturnsContent guards the other half of F2: a *live* note's 452 payload
// must still carry the current content, otherwise the documented reconcile-and-retry loop breaks.
func TestLiveConflictStillReturnsContent(t *testing.T) {
	i := newInstance(t)
	path := "Projects/Live.md"
	base := i.seedNote(path, "live-v1\n")
	if env := i.write(path, "live-v2\n", int64p(base)); env.Code != 1 {
		t.Fatalf("advance: code=%d %s", env.Code, env.Message)
	}

	env := i.write(path, "stale\n", int64p(base))
	if env.Code != 452 {
		t.Fatalf("stale write must be 452, got code=%d %s", env.Code, env.Message)
	}
	var payload struct {
		Revision int64  `json:"revision"`
		Deleted  bool   `json:"deleted"`
		Content  string `json:"content"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode 452 payload: %v", err)
	}
	if payload.Deleted || payload.Content != "live-v2\n" || payload.Revision != base+1 {
		t.Fatalf("452 on a live note must carry content/revision, got %+v", payload)
	}

	// MCP surfaces the same payload
	res := i.mcpCall(t, "note_create_or_update", map[string]any{
		"vault": testVault, "path": path, "content": "mcp-stale\n", "revision": base,
	})
	if !res.IsError || !strings.Contains(res.Text, `"code":452`) {
		t.Fatalf("MCP stale write must fail with 452, got %+v", res)
	}
	if !strings.Contains(res.Text, "live-v2") {
		t.Fatalf("MCP 452 payload must carry the live content, got %q", res.Text)
	}

	// and MCP must not echo a tombstone body either
	if env := i.do("DELETE", "/api/note?vault="+testVault+"&path="+urlQueryEscape(path), nil, i.apiToken, "occapi"); env.Code != 1 {
		t.Fatalf("delete: code=%d %s", env.Code, env.Message)
	}
	res = i.mcpCall(t, "note_create_or_update", map[string]any{
		"vault": testVault, "path": path, "content": "mcp-revive\n",
	})
	if !res.IsError || !strings.Contains(res.Text, `"code":453`) {
		t.Fatalf("MCP write to a tombstone must fail with 453, got %+v", res)
	}
	if strings.Contains(res.Text, "live-v2") {
		t.Fatalf("MCP 453 payload must not echo the tombstone body, got %q", res.Text)
	}
}

// TestConflictPayloadDescribesTargetRow is finding F5: when more than one row matches a path
// (a live row plus a renamed tombstone, or historical duplicates), the conflict payload must
// describe the row the compare-and-swap addressed, not whichever row a path lookup returns first.
func TestConflictPayloadDescribesTargetRow(t *testing.T) {
	i := newInstance(t)
	src := "Projects/Ambiguous.md"
	dst := "Projects/Ambiguous-target.md"

	// filler note keeps id 1 busy, so the target note gets id 2
	i.seedNote("Projects/Filler.md", "filler\n")
	base := i.seedNote(src, "live content\n") // id 2, rev 1
	if env := i.write(src, "live content v2\n", int64p(base)); env.Code != 1 {
		t.Fatalf("advance: code=%d %s", env.Code, env.Message)
	}
	live := i.getNote(src) // rev 2

	// Recreate the duplicate state the codebase itself defends against (rename leaves a
	// rename=1 tombstone at the old path; a crash mid-rename or the historical duplicate bug
	// leaves a second row at the same path_hash). The duplicate has a LOWER primary key, so an
	// unordered `First()` path lookup returns it instead of the live row.
	db := i.userDB()
	if _, err := db.Exec(`DELETE FROM note WHERE path = ?`, "Projects/Filler.md"); err != nil {
		t.Fatalf("free id 1: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO note
		(id, vault_id, action, rename, fid, path, path_hash, content, content_hash,
		 content_last_snapshot, content_last_snapshot_hash, version, rev,
		 client_name, client_type, client_version, size, ctime, mtime, updated_timestamp, created_at, updated_at)
		VALUES (1, 1, 'delete', 1, 0, ?, ?, '', 'tombstone-hash', '', '', 0, 50,
			'', '', '', 0, 0, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		src, hashOf(src)); err != nil {
		t.Fatalf("insert duplicate tombstone row: %v", err)
	}

	// a stale conditional rename from the live row must be refused
	env := i.do("POST", "/api/note/rename", map[string]any{
		"vault": testVault, "oldPath": src, "path": dst, "revision": base, // stale (live is base+1)
	}, i.apiToken, "occapi")
	if env.Code != 452 {
		t.Fatalf("stale rename must be 452, got code=%d %s", env.Code, env.Message)
	}
	var payload struct {
		Revision    int64  `json:"revision"`
		ContentHash string `json:"contentHash"`
		Deleted     bool   `json:"deleted"`
		Content     string `json:"content"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.Deleted {
		t.Fatalf("payload described a tombstone (%+v); it must describe the live row the CAS targeted", payload)
	}
	if payload.Revision != live.Revision || payload.ContentHash != live.Hash || payload.Content != live.Content {
		t.Fatalf("payload must describe the CAS target row: got %+v, live row %+v", payload, live)
	}

	// the ambiguous duplicate must not have leaked into normal reads either
	if got := i.getNote(src); got.Content != "live content v2\n" || got.Revision != live.Revision {
		t.Fatalf("reads must still resolve the live row, got %+v", got)
	}
}

// TestHistoryRestoreAdvancesRevision is finding F6: restoring a historical version is a real
// mutation and must invalidate revisions read before it.
func TestHistoryRestoreAdvancesRevision(t *testing.T) {
	i := newInstance(t)
	path := "Projects/Hist.md"
	base := i.seedNote(path, "current content\n") // rev 1

	var noteID int64
	if err := i.userDB().QueryRow(`SELECT id FROM note WHERE path = ?`, path).Scan(&noteID); err != nil {
		t.Fatalf("note id: %v", err)
	}
	// a history row as the delayed snapshot task would have written it (empty diff = full body).
	// Seeding through the repository lets the service create and migrate its own shard.
	history, err := i.app.Repositories.NoteHistoryRepo.Create(context.Background(), &domain.NoteHistory{
		NoteID:      noteID,
		VaultID:     1,
		Path:        path,
		Content:     "historic content\n",
		ContentHash: hashOf("historic content\n"),
		DiffPatch:   "",
		Version:     1,
	}, 1)
	if err != nil {
		t.Fatalf("seed history row: %v", err)
	}
	historyID := history.ID

	env := i.do("PUT", "/api/note/history/restore", map[string]any{
		"vault": testVault, "historyId": historyID,
	}, i.apiToken, "occapi")
	if env.Code != 1 {
		t.Fatalf("history restore: code=%d %s", env.Code, env.Message)
	}
	restored := i.getNote(path)
	if restored.Content != "historic content\n" {
		t.Fatalf("history restore must apply the historic body, got %q", restored.Content)
	}
	if restored.Revision != base+1 {
		t.Fatalf("history restore must advance the revision %d -> %d, got %d", base, base+1, restored.Revision)
	}
	// the revision read before the restore is now stale
	if env := i.write(path, "stale after history\n", int64p(base)); env.Code != 452 {
		t.Fatalf("a revision read before a history restore must be stale, got code=%d %s", env.Code, env.Message)
	}
	if got := i.getNote(path); got.Content != "historic content\n" {
		t.Fatalf("stale write must not overwrite the restored body, got %q", got.Content)
	}
}

// TestFolderDeleteCascadeAdvancesRevision is the second half of F6: deleting a folder cascades a
// soft delete into every note it contains, so previously read revisions must be invalidated.
func TestFolderDeleteCascadeAdvancesRevision(t *testing.T) {
	i := newInstance(t)
	path := "Projects/Cascade/Inner.md"
	if env := i.do("POST", "/api/folder", map[string]any{
		"vault": testVault, "path": "Projects/Cascade",
	}, i.apiToken, "occapi"); env.Code != 1 {
		t.Fatalf("create folder: code=%d %s", env.Code, env.Message)
	}
	base := i.seedNote(path, "inside the folder\n")

	// The recursive delete is the MCP `folder_delete` tool (DELETE /api/folder only removes the
	// folder record itself; that is pre-existing behaviour and out of scope here).
	res := i.mcpCall(t, "folder_delete", map[string]any{"vault": testVault, "path": "Projects/Cascade"})
	if res.IsError {
		t.Fatalf("recursive folder delete failed: %q", res.Text)
	}
	if _, exists := i.noteExists(path); exists {
		t.Fatal("the cascade must soft-delete the contained note")
	}

	// the pre-cascade revision is refused, and the note is not resurrected
	env := i.write(path, "revive after cascade\n", int64p(base))
	if env.Code != 452 && env.Code != 453 {
		t.Fatalf("a revision read before the cascade must be refused, got code=%d %s", env.Code, env.Message)
	}
	if _, exists := i.noteExists(path); exists {
		t.Fatal("a stale write must not resurrect a cascaded note")
	}

	// the tombstone's own revision still restores it
	var page struct {
		List []struct {
			Path     string `json:"path"`
			Revision int64  `json:"revision"`
		} `json:"list"`
	}
	list := i.do("GET", "/api/notes?vault="+testVault+"&isRecycle=true&page=1&pageSize=50", nil, i.apiToken, "occapi")
	if err := json.Unmarshal(list.Data, &page); err != nil {
		t.Fatalf("decode recycle list: %v", err)
	}
	var tombRev int64
	for _, n := range page.List {
		if n.Path == path {
			tombRev = n.Revision
		}
	}
	if tombRev <= base {
		t.Fatalf("the cascade must advance the note revision beyond %d, got %d", base, tombRev)
	}
	if env := i.do("PUT", "/api/note/restore", map[string]any{
		"vault": testVault, "path": path, "revision": tombRev,
	}, i.apiToken, "occapi"); env.Code != 1 {
		t.Fatalf("restore after cascade: code=%d %s", env.Code, env.Message)
	}
	if got := i.getNote(path); got.Content != "inside the folder\n" {
		t.Fatalf("restore must return the preserved body, got %q", got.Content)
	}
}

// TestRowGoneConflictPayload covers the non-racy end of finding N1: the row is gone entirely (the
// tombstone was purged from the recycle bin), so the conflict must report revision 0 and disclose
// nothing — over REST and over MCP, with the same code and shape.
func TestRowGoneConflictPayload(t *testing.T) {
	i := newInstance(t)
	path := "Projects/Gone.md"
	base := i.seedNote(path, "purge me\n")

	if env := i.do("DELETE", "/api/note?vault="+testVault+"&path="+urlQueryEscape(path)+
		"&revision="+fmt.Sprint(base), nil, i.apiToken, "occapi"); env.Code != 1 {
		t.Fatalf("delete: code=%d %s", env.Code, env.Message)
	}
	// Hard-delete the row, exactly as the retention task (DeletePhysicalByTime) or the duplicate
	// cleanup does after the retention window. `DELETE /api/note/recycle-clear` is not a hard
	// delete (it only flags rename=2), so the fixture is applied directly to the note shard.
	var noteID int64
	if err := i.userDB().QueryRow(`SELECT id FROM note WHERE path = ?`, path).Scan(&noteID); err != nil {
		t.Fatalf("note id: %v", err)
	}
	if _, err := i.userDB().Exec(`DELETE FROM note WHERE id = ?`, noteID); err != nil {
		t.Fatalf("hard delete row: %v", err)
	}
	if _, exists := i.noteExists(path); exists {
		t.Fatal("note must be gone after the row was hard-deleted")
	}

	env := i.write(path, "resurrect?\n", int64p(base))
	if env.Code != 452 {
		t.Fatalf("write to a purged row must be 452, got code=%d %s", env.Code, env.Message)
	}
	if !strings.Contains(env.Message, "http=409") {
		t.Fatalf("row-gone conflict must use HTTP 409, got %s", env.Message)
	}
	var payload struct {
		Path        string `json:"path"`
		Revision    int64  `json:"revision"`
		ContentHash string `json:"contentHash"`
		Mtime       int64  `json:"mtime"`
		Deleted     bool   `json:"deleted"`
		Content     string `json:"content"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode row-gone payload: %v", err)
	}
	if payload.Revision != 0 || payload.Path != "" || payload.Content != "" ||
		payload.ContentHash != "" || payload.Mtime != 0 || payload.Deleted {
		t.Fatalf("row-gone payload must be empty apart from revision 0, got %+v", payload)
	}

	// MCP reports the same thing through its own envelope
	res := i.mcpCall(t, "note_create_or_update", map[string]any{
		"vault": testVault, "path": path, "content": "resurrect?\n", "revision": base,
	})
	if !res.IsError || !strings.Contains(res.Text, `"code":452`) {
		t.Fatalf("MCP write to a purged row must fail with 452, got %+v", res)
	}
	if !strings.Contains(res.Text, `"revision":0`) {
		t.Fatalf("MCP row-gone payload must carry revision 0, got %q", res.Text)
	}
	if strings.Contains(res.Text, "purge me") {
		t.Fatalf("MCP row-gone payload must not disclose the old body, got %q", res.Text)
	}

	// and the note was not recreated by either attempt
	if _, exists := i.noteExists(path); exists {
		t.Fatal("a refused conditional write must not recreate a purged note")
	}
}
