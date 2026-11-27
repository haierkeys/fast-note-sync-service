package occtest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
)

// TestReadRevisionBothClients covers the first row of the acceptance matrix: two clients that
// read the same note observe the same revision.
func TestReadRevisionBothClients(t *testing.T) {
	i := newInstance(t)
	rev := i.seedNote("Projects/Occ.md", "v1\n")
	if rev == 0 {
		t.Fatalf("a freshly created note must expose a non-zero revision, got %d", rev)
	}
	a, b := i.getNote("Projects/Occ.md"), i.getNote("Projects/Occ.md")
	if a.Revision != b.Revision || a.Revision != rev {
		t.Fatalf("both readers must see revision %d, got %d and %d", rev, a.Revision, b.Revision)
	}
	// the same value is exposed by the list endpoint, which is what an agent usually reads first
	env := i.do("GET", "/api/notes?vault="+testVault+"&page=1&pageSize=20", nil, i.apiToken, "occapi")
	if env.Code != 1 {
		t.Fatalf("list: code=%d %s", env.Code, env.Message)
	}
	var page struct {
		List []struct {
			Path     string `json:"path"`
			Revision int64  `json:"revision"`
		} `json:"list"`
	}
	if err := json.Unmarshal(env.Data, &page); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	found := false
	for _, n := range page.List {
		if n.Path == "Projects/Occ.md" {
			found = true
			if n.Revision != rev {
				t.Fatalf("list revision %d != note revision %d", n.Revision, rev)
			}
		}
	}
	if !found {
		t.Fatal("note missing from list response")
	}
}

// TestFirstWriteAdvancesRevisionAndSecondIsRejected is the reproduction of OBSIDIAN-FNS-002
// Test B, now expected to be a conflict instead of a silent overwrite.
func TestFirstWriteAdvancesRevisionAndSecondIsRejected(t *testing.T) {
	i := newInstance(t)
	path := "Projects/Concurrent.md"
	base := i.seedNote(path, "base\n")

	// both clients read the same revision
	revA, revB := i.getNote(path).Revision, i.getNote(path).Revision
	if revA != base || revB != base {
		t.Fatalf("both clients must read revision %d, got %d and %d", base, revA, revB)
	}

	// first client writes with the revision it read
	env := i.write(path, "edit-from-A\n", int64p(revA))
	if env.Code != 1 {
		t.Fatalf("first conditional write must succeed: code=%d %s", env.Code, env.Message)
	}
	var written noteView
	if err := json.Unmarshal(env.Data, &written); err != nil {
		t.Fatalf("decode write response: %v", err)
	}
	if written.Revision != base+1 {
		t.Fatalf("revision must advance to %d, got %d", base+1, written.Revision)
	}

	// second client writes with the now-stale revision
	env = i.write(path, "edit-from-B\n", int64p(revB))
	if env.Code != 452 {
		t.Fatalf("stale conditional write must be rejected with 452, got code=%d %s", env.Code, env.Message)
	}
	if !strings.Contains(env.Message, "http=409") {
		t.Fatalf("a rejected conditional write must use HTTP 409, got %s", env.Message)
	}
	var conflict struct {
		Path     string `json:"path"`
		Revision int64  `json:"revision"`
		Content  string `json:"content"`
		Deleted  bool   `json:"deleted"`
	}
	if err := json.Unmarshal(env.Data, &conflict); err != nil {
		t.Fatalf("decode conflict data: %v", err)
	}
	if conflict.Revision != base+1 || conflict.Content != "edit-from-A\n" || conflict.Deleted {
		t.Fatalf("conflict payload must describe the stored state, got %+v", conflict)
	}

	// the note is untouched, both in the API and on disk
	after := i.getNote(path)
	if after.Content != "edit-from-A\n" || after.Revision != base+1 {
		t.Fatalf("stale write must not change the note: %+v", after)
	}
	if got := i.contentFile(path); got != "edit-from-A\n" {
		t.Fatalf("content file must be unchanged, got %q", got)
	}

	// the losing client re-reads and retries successfully
	env = i.write(path, "edit-from-B\n", int64p(after.Revision))
	if env.Code != 1 {
		t.Fatalf("retry with the current revision must succeed: code=%d %s", env.Code, env.Message)
	}
	if got := i.getNote(path); got.Content != "edit-from-B\n" || got.Revision != base+2 {
		t.Fatalf("retry must apply and advance the revision: %+v", got)
	}
}

// TestConcurrentWritersOnlyOneSucceeds is the stress version: N writers race on the same
// revision. Exactly one may win; every other request must be a conflict and the note must
// end up containing exactly the winner's payload.
func TestConcurrentWritersOnlyOneSucceeds(t *testing.T) {
	i := newInstance(t)
	path := "Projects/Stress.md"
	base := i.seedNote(path, "base\n")

	const writers = 24
	var wg sync.WaitGroup
	results := make([]envelope, writers)
	for n := 0; n < writers; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			content := fmt.Sprintf("writer-%02d\n", n)
			results[n] = i.write(path, content, int64p(base))
		}(n)
	}
	wg.Wait()

	successes, conflicts := 0, 0
	winner := ""
	for n, r := range results {
		switch r.Code {
		case 1:
			successes++
			winner = fmt.Sprintf("writer-%02d\n", n)
		case 452:
			conflicts++
		default:
			t.Fatalf("writer %d got unexpected code %d: %s", n, r.Code, r.Message)
		}
	}
	if successes != 1 || conflicts != writers-1 {
		t.Fatalf("exactly one writer may succeed: successes=%d conflicts=%d", successes, conflicts)
	}
	final := i.getNote(path)
	if final.Revision != base+1 {
		t.Fatalf("revision must advance exactly once (%d), got %d", base+1, final.Revision)
	}
	if final.Content != winner {
		t.Fatalf("note must hold the winner's payload %q, got %q", winner, final.Content)
	}
	if got := i.contentFile(path); got != winner {
		t.Fatalf("content file must match the winner, got %q", got)
	}
}

// TestStaleReadModifyWriteHelpers covers append / prepend / replace / frontmatter: they must
// refuse to compose on top of a revision the caller no longer holds.
func TestStaleReadModifyWriteHelpers(t *testing.T) {
	i := newInstance(t)
	path := "Projects/Helpers.md"
	base := i.seedNote(path, "# Title\n\nbody\n")

	type tc struct {
		name string
		path string
		body map[string]any
	}
	cases := []tc{
		{"append", "/api/note/append", map[string]any{"vault": testVault, "path": path, "content": "\nappended\n"}},
		{"prepend", "/api/note/prepend", map[string]any{"vault": testVault, "path": path, "content": "prepended\n\n"}},
		{"replace", "/api/note/replace", map[string]any{"vault": testVault, "path": path, "find": "moved-on", "replace": "MOVED-ON", "all": true}},
		{"frontmatter", "/api/note/frontmatter", map[string]any{"vault": testVault, "path": path,
			"updates": map[string]any{"status": "occ"}, "remove": []string{}}},
	}

	// move the note on so that `base` is stale
	if env := i.write(path, "# Title\n\nmoved-on\n", int64p(base)); env.Code != 1 {
		t.Fatalf("advance note: code=%d %s", env.Code, env.Message)
	}
	current := i.getNote(path)

	method := map[string]string{
		"append": "POST", "prepend": "POST", "replace": "POST", "frontmatter": "PATCH",
	}
	for _, c := range cases {
		body := map[string]any{}
		for k, v := range c.body {
			body[k] = v
		}
		body["revision"] = base // stale
		env := i.do(method[c.name], c.path, body, i.apiToken, "occapi")
		if env.Code != 452 {
			t.Fatalf("%s with a stale revision must conflict, got code=%d %s", c.name, env.Code, env.Message)
		}
		if !strings.Contains(env.Message, "http=409") {
			t.Fatalf("%s conflict must use HTTP 409, got %s", c.name, env.Message)
		}
		after := i.getNote(path)
		if after.Content != current.Content || after.Revision != current.Revision {
			t.Fatalf("%s must not modify the note: %+v", c.name, after)
		}

		// the same call with the current revision is applied
		body["revision"] = after.Revision
		if env := i.do(method[c.name], c.path, body, i.apiToken, "occapi"); env.Code != 1 {
			t.Fatalf("%s with the current revision must succeed: code=%d %s", c.name, env.Code, env.Message)
		}
		applied := i.getNote(path)
		if applied.Revision != after.Revision+1 {
			t.Fatalf("%s must advance the revision once: %d -> %d", c.name, after.Revision, applied.Revision)
		}
		current = applied // the next case must start from the note's new state
	}
}

// TestConcurrentAppendsKeepExactlyOneEdit proves the read-modify-write helpers cannot lose an
// update: N simultaneous appends on the same revision apply at most once.
func TestConcurrentAppendsKeepExactlyOneEdit(t *testing.T) {
	i := newInstance(t)
	path := "Projects/Append.md"
	base := i.seedNote(path, "start\n")

	const writers = 12
	var wg sync.WaitGroup
	results := make([]envelope, writers)
	for n := 0; n < writers; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			results[n] = i.do("POST", "/api/note/append", map[string]any{
				"vault": testVault, "path": path,
				"content":  fmt.Sprintf("append-%02d\n", n),
				"revision": base,
			}, i.apiToken, "occapi")
		}(n)
	}
	wg.Wait()

	successes, conflicts := 0, 0
	for _, r := range results {
		switch r.Code {
		case 1:
			successes++
		case 452:
			conflicts++
		default:
			t.Fatalf("unexpected code %d: %s", r.Code, r.Message)
		}
	}
	if successes != 1 || conflicts != writers-1 {
		t.Fatalf("exactly one append may win: successes=%d conflicts=%d", successes, conflicts)
	}
	final := i.getNote(path)
	if strings.Count(final.Content, "append-") != 1 {
		t.Fatalf("exactly one append must be present, got %q", final.Content)
	}
	if final.Revision != base+1 {
		t.Fatalf("revision must advance once, got %d (base %d)", final.Revision, base)
	}
}

// TestDeleteThenStaleWriteCannotResurrect covers the delete/tombstone rules.
func TestDeleteThenStaleWriteCannotResurrect(t *testing.T) {
	i := newInstance(t)
	path := "Projects/Deleted.md"
	base := i.seedNote(path, "base\n")

	// delete with the current revision
	env := i.do("DELETE", "/api/note?vault="+testVault+"&path="+urlQueryEscape(path)+"&revision="+fmt.Sprint(base), nil, i.apiToken, "occapi")
	if env.Code != 1 {
		t.Fatalf("conditional delete must succeed: code=%d %s", env.Code, env.Message)
	}
	if _, ok := i.noteExists(path); ok {
		t.Fatal("note must be gone after delete")
	}

	// a stale writer using the pre-delete revision must not resurrect it
	env = i.write(path, "stale-write\n", int64p(base))
	if env.Code != 452 && env.Code != 453 {
		t.Fatalf("stale write after delete must conflict, got code=%d %s", env.Code, env.Message)
	}
	if !strings.Contains(env.Message, "http=409") {
		t.Fatalf("conflict must use HTTP 409, got %s", env.Message)
	}
	if _, ok := i.noteExists(path); ok {
		t.Fatal("stale write must not resurrect a deleted note")
	}

	// a stale delete of a *live* note is a plain revision conflict and changes nothing
	other := "Projects/DeleteStale.md"
	otherRev := i.seedNote(other, "keep-me\n")
	if env := i.write(other, "advanced\n", int64p(otherRev)); env.Code != 1 {
		t.Fatalf("advance other note: code=%d %s", env.Code, env.Message)
	}
	env = i.do("DELETE", "/api/note?vault="+testVault+"&path="+urlQueryEscape(other)+"&revision="+fmt.Sprint(otherRev), nil, i.apiToken, "occapi")
	if env.Code != 452 {
		t.Fatalf("stale delete must be a revision conflict, got code=%d %s", env.Code, env.Message)
	}
	if got, ok := i.noteExists(other); !ok || got.Content != "advanced\n" {
		t.Fatalf("stale delete must not remove the note: %+v", got)
	}

	// a legacy unconditional write must not resurrect it either (the fix for the
	// OBSIDIAN-FNS-002 resurrection finding)
	env = i.write(path, "legacy-write\n", nil)
	if env.Code != 453 {
		t.Fatalf("unconditional write to a tombstone must be refused with 453, got code=%d %s", env.Code, env.Message)
	}
	if _, ok := i.noteExists(path); ok {
		t.Fatal("legacy write must not resurrect a deleted note")
	}

	// restore is the authorised revival path; a stale revision cannot restore a newer tombstone
	env = i.do("PUT", "/api/note/restore", map[string]any{
		"vault": testVault, "path": path, "revision": base - 1,
	}, i.apiToken, "occapi")
	if env.Code != 452 {
		t.Fatalf("restore with a stale revision must conflict, got code=%d %s", env.Code, env.Message)
	}

	// find the tombstone revision through the recycle bin and restore with it
	list := i.do("GET", "/api/notes?vault="+testVault+"&isRecycle=true&page=1&pageSize=20", nil, i.apiToken, "occapi")
	var page struct {
		List []struct {
			Path     string `json:"path"`
			Revision int64  `json:"revision"`
		} `json:"list"`
	}
	if err := json.Unmarshal(list.Data, &page); err != nil {
		t.Fatalf("decode recycle list: %v", err)
	}
	var tombRev int64
	for _, n := range page.List {
		if n.Path == path {
			tombRev = n.Revision
		}
	}
	if tombRev == 0 {
		t.Fatal("deleted note must be listed in the recycle bin with its revision")
	}
	env = i.do("PUT", "/api/note/restore", map[string]any{
		"vault": testVault, "path": path, "revision": tombRev,
	}, i.apiToken, "occapi")
	if env.Code != 1 {
		t.Fatalf("restore with the tombstone revision must succeed: code=%d %s", env.Code, env.Message)
	}
	restored := i.getNote(path)
	if restored.Content != "base\n" || restored.Revision != tombRev+1 {
		t.Fatalf("restore must return the last content and advance the revision: %+v", restored)
	}
}

// TestRenameConflictLeavesNoDuplicate covers rename/move: two clients renaming the same note
// from the same revision cannot both succeed.
func TestRenameConflictLeavesNoDuplicate(t *testing.T) {
	i := newInstance(t)
	src := "Projects/Rename.md"
	base := i.seedNote(src, "content\n")

	first := i.do("POST", "/api/note/rename", map[string]any{
		"vault": testVault, "oldPath": src, "path": "Projects/Renamed-A.md", "revision": base,
	}, i.apiToken, "occapi")
	if first.Code != 1 {
		t.Fatalf("first rename must succeed: code=%d %s", first.Code, first.Message)
	}

	second := i.do("POST", "/api/note/rename", map[string]any{
		"vault": testVault, "oldPath": src, "path": "Projects/Renamed-B.md", "revision": base,
	}, i.apiToken, "occapi")
	if second.Code == 1 {
		t.Fatal("second rename from the same revision must not succeed")
	}
	if second.Code != 452 && second.Code != 430 {
		t.Fatalf("second rename must conflict or report the source as gone, got code=%d %s", second.Code, second.Message)
	}
	if _, ok := i.noteExists("Projects/Renamed-B.md"); ok {
		t.Fatal("conflicting rename must not create a second copy")
	}
	if _, ok := i.noteExists("Projects/Renamed-A.md"); !ok {
		t.Fatal("the successful rename target must exist")
	}
}

// TestRestAndMCPShareTheGuard proves the MCP surface inherits exactly the same check (one
// implementation, not a parallel one).
func TestRestAndMCPShareTheGuard(t *testing.T) {
	i := newInstance(t)
	path := "Projects/Shared.md"
	base := i.seedNote(path, "base\n")

	// REST client writes first
	if env := i.write(path, "from-rest\n", int64p(base)); env.Code != 1 {
		t.Fatalf("REST write: code=%d %s", env.Code, env.Message)
	}

	// MCP client, holding the same (now stale) revision, is rejected by the same guard
	res := i.mcpCall(t, "note_create_or_update", map[string]any{
		"vault": testVault, "path": path, "content": "from-mcp\n", "revision": base,
	})
	if !res.IsError {
		t.Fatalf("MCP write with a stale revision must fail, got %+v", res)
	}
	var payload struct {
		Code int `json:"code"`
		Data struct {
			Revision int64  `json:"revision"`
			Content  string `json:"content"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(res.Text), &payload); err != nil {
		t.Fatalf("MCP conflict must be machine-readable JSON, got %q: %v", res.Text, err)
	}
	if payload.Code != 452 || payload.Data.Revision != base+1 || payload.Data.Content != "from-rest\n" {
		t.Fatalf("MCP conflict payload mismatch: %+v", payload)
	}
	if got := i.getNote(path); got.Content != "from-rest\n" {
		t.Fatalf("MCP conflict must not change the note: %+v", got)
	}

	// MCP retry with the current revision succeeds, and MCP reads expose the revision
	res = i.mcpCall(t, "note_create_or_update", map[string]any{
		"vault": testVault, "path": path, "content": "from-mcp\n", "revision": base + 1,
	})
	if res.IsError {
		t.Fatalf("MCP retry must succeed, got %q", res.Text)
	}
	if res.Revision != base+2 {
		t.Fatalf("MCP write must report the new revision %d, got %d", base+2, res.Revision)
	}

	// and REST sees the MCP write
	if got := i.getNote(path); got.Content != "from-mcp\n" || got.Revision != base+2 {
		t.Fatalf("REST must observe the MCP write: %+v", got)
	}
}

// TestLegacyUnconditionalWritesDocumentedBehaviour pins the compatibility policy: a request
// without a revision is still accepted (unchanged behaviour) and still advances the revision,
// so a later conditional writer is protected from it.
func TestLegacyUnconditionalWritesDocumentedBehaviour(t *testing.T) {
	i := newInstance(t)
	path := "Projects/Legacy.md"
	base := i.seedNote(path, "v1\n")

	// a conditional client reads `base`
	reader := i.getNote(path).Revision

	// a legacy client overwrites without a revision
	if env := i.write(path, "legacy-overwrite\n", nil); env.Code != 1 {
		t.Fatalf("legacy write must still succeed: code=%d %s", env.Code, env.Message)
	}
	advanced := i.getNote(path)
	if advanced.Revision != base+1 {
		t.Fatalf("legacy write must advance the revision (%d -> %d), got %d", base, base+1, advanced.Revision)
	}

	// the conditional client is now protected: its stale revision is refused
	env := i.write(path, "conditional\n", int64p(reader))
	if env.Code != 452 {
		t.Fatalf("conditional write after a legacy write must conflict, got code=%d %s", env.Code, env.Message)
	}
	if got := i.getNote(path); got.Content != "legacy-overwrite\n" {
		t.Fatalf("refused write must not overwrite the legacy content: %+v", got)
	}
}

// TestMigrationBaselineZeroAndRestart makes sure pre-existing rows (created before the rev
// column existed) work with revision 0, and that the revision survives a restart of the
// service against the same database + storage.
func TestMigrationBaselineZeroAndRestart(t *testing.T) {
	i := newInstance(t)
	path := "Projects/Restart.md"
	rev := i.seedNote(path, "v1\n")

	// simulate a row that predates the rev column
	if _, err := i.userDB().Exec(`UPDATE note SET rev = 0 WHERE path = ?`, path); err != nil {
		t.Fatalf("reset rev: %v", err)
	}
	if got := i.getNote(path).Revision; got != 0 {
		t.Fatalf("a legacy row must present revision 0, got %d", got)
	}
	if env := i.write(path, "v2\n", int64p(0)); env.Code != 1 {
		t.Fatalf("revision 0 must be a usable baseline: code=%d %s", env.Code, env.Message)
	}
	if env := i.write(path, "v3\n", int64p(0)); env.Code != 452 {
		t.Fatalf("revision 0 must be stale after the first write: code=%d %s", env.Code, env.Message)
	}

	// restart: same database, same storage directory, fresh process-level state
	rev = i.getNote(path).Revision
	i.restart()
	if got := i.getNote(path); got.Revision != rev || got.Content != "v2\n" {
		t.Fatalf("revision must survive a restart: want revision %d content %q, got %+v", rev, "v2\n", got)
	}
	if env := i.write(path, "v4\n", int64p(rev)); env.Code != 1 {
		t.Fatalf("conditional writes must keep working after a restart: code=%d %s", env.Code, env.Message)
	}
	if env := i.write(path, "v5\n", int64p(rev)); env.Code != 452 {
		t.Fatalf("the pre-restart revision must be stale after the write: code=%d %s", env.Code, env.Message)
	}
}

// TestMissingRevisionOnAbsentNote: a conditional write to a note that does not exist cannot
// be satisfied and must not silently create it.
func TestMissingRevisionOnAbsentNote(t *testing.T) {
	i := newInstance(t)
	env := i.write("Projects/Absent.md", "content\n", int64p(3))
	if env.Code != 452 {
		t.Fatalf("conditional create-with-revision must conflict, got code=%d %s", env.Code, env.Message)
	}
	if _, ok := i.noteExists("Projects/Absent.md"); ok {
		t.Fatal("a conditional write must not create a missing note")
	}
	// legacy create still works
	if env := i.write("Projects/Absent.md", "content\n", nil); env.Code != 1 {
		t.Fatalf("legacy create must still work: code=%d %s", env.Code, env.Message)
	}
}

var _ = http.StatusConflict

// TestNoOpReplaceStillChecksRevision documents a subtle case: a find/replace that would change
// nothing must not report success to a caller whose revision is already stale.
func TestNoOpReplaceStillChecksRevision(t *testing.T) {
	i := newInstance(t)
	path := "Projects/NoOp.md"
	base := i.seedNote(path, "alpha\n")

	// move the note on
	if env := i.write(path, "beta\n", int64p(base)); env.Code != 1 {
		t.Fatalf("advance: code=%d %s", env.Code, env.Message)
	}

	// a replace that matches nothing (no-op) with the stale revision must still conflict
	env := i.do("POST", "/api/note/replace", map[string]any{
		"vault": testVault, "path": path,
		"find": "does-not-exist", "replace": "x", "all": true,
		"revision": base,
	}, i.apiToken, "occapi")
	if env.Code != 452 {
		t.Fatalf("no-op replace with a stale revision must conflict, got code=%d %s", env.Code, env.Message)
	}

	// the same no-op with the current revision is a success and does not advance the revision
	cur := i.getNote(path)
	env = i.do("POST", "/api/note/replace", map[string]any{
		"vault": testVault, "path": path,
		"find": "does-not-exist", "replace": "x", "all": true,
		"revision": cur.Revision,
	}, i.apiToken, "occapi")
	if env.Code != 1 {
		t.Fatalf("no-op replace with the current revision must succeed: code=%d %s", env.Code, env.Message)
	}
	if got := i.getNote(path); got.Revision != cur.Revision {
		t.Fatalf("a no-op must not advance the revision: %d -> %d", cur.Revision, got.Revision)
	}
}

// TestConflictLeavesRowUntouched asserts that a rejected write changes no persisted state at
// all: not the content file, not the hash, not mtime, not the revision.
func TestConflictLeavesRowUntouched(t *testing.T) {
	i := newInstance(t)
	path := "Projects/Untouched.md"
	base := i.seedNote(path, "stable\n")
	if env := i.write(path, "newer\n", int64p(base)); env.Code != 1 {
		t.Fatalf("advance: code=%d %s", env.Code, env.Message)
	}
	before := i.getNote(path)
	beforeFile := i.contentFile(path)
	var beforeMtime, beforeRev int64
	var beforeHash string
	if err := i.userDB().QueryRow(`SELECT mtime, rev, content_hash FROM note WHERE path = ?`, path).
		Scan(&beforeMtime, &beforeRev, &beforeHash); err != nil {
		t.Fatalf("read row: %v", err)
	}

	env := i.write(path, "conflicting\n", int64p(base))
	if env.Code != 452 {
		t.Fatalf("stale write must conflict: code=%d %s", env.Code, env.Message)
	}

	after := i.getNote(path)
	if after.Content != before.Content || after.Revision != before.Revision ||
		after.Hash != before.Hash || after.Mtime != before.Mtime {
		t.Fatalf("conflict must not change the note: before=%+v after=%+v", before, after)
	}
	if got := i.contentFile(path); got != beforeFile {
		t.Fatalf("conflict must not change the content file: %q != %q", got, beforeFile)
	}
	var afterMtime, afterRev int64
	var afterHash string
	if err := i.userDB().QueryRow(`SELECT mtime, rev, content_hash FROM note WHERE path = ?`, path).
		Scan(&afterMtime, &afterRev, &afterHash); err != nil {
		t.Fatalf("read row: %v", err)
	}
	if afterMtime != beforeMtime || afterRev != beforeRev || afterHash != beforeHash {
		t.Fatalf("conflict must not change the row: mtime %d->%d rev %d->%d hash %q->%q",
			beforeMtime, afterMtime, beforeRev, afterRev, beforeHash, afterHash)
	}
}

// TestMCPDeleteRestoreRenameRevision covers the lifecycle tools on the MCP surface.
func TestMCPDeleteRestoreRenameRevision(t *testing.T) {
	i := newInstance(t)
	path := "Projects/McpLife.md"
	base := i.seedNote(path, "life\n")

	// stale delete is rejected and the note survives
	res := i.mcpCall(t, "note_delete", map[string]any{"vault": testVault, "path": path, "revision": base + 41})
	if !res.IsError || !strings.Contains(res.Text, `"code":452`) {
		t.Fatalf("stale MCP delete must be a 452 conflict, got %+v", res)
	}
	if _, ok := i.noteExists(path); !ok {
		t.Fatal("stale MCP delete must not delete the note")
	}

	// current-revision delete succeeds
	res = i.mcpCall(t, "note_delete", map[string]any{"vault": testVault, "path": path, "revision": base})
	if res.IsError {
		t.Fatalf("MCP delete with the current revision must succeed: %+v", res)
	}
	if _, ok := i.noteExists(path); ok {
		t.Fatal("MCP delete must remove the note")
	}

	// restore with the tombstone revision works
	var tombRev int64
	var p struct {
		List []struct {
			Path     string `json:"path"`
			Revision int64  `json:"revision"`
		} `json:"list"`
	}
	list := i.do("GET", "/api/notes?vault="+testVault+"&isRecycle=true&page=1&pageSize=20", nil, i.apiToken, "occapi")
	if err := json.Unmarshal(list.Data, &p); err != nil {
		t.Fatalf("decode recycle list: %v", err)
	}
	for _, n := range p.List {
		if n.Path == path {
			tombRev = n.Revision
		}
	}
	res = i.mcpCall(t, "note_restore", map[string]any{"vault": testVault, "path": path, "revision": tombRev})
	if res.IsError {
		t.Fatalf("MCP restore with the tombstone revision must succeed: %+v", res)
	}

	// rename with a stale revision is rejected, with the current one succeeds
	cur := i.getNote(path)
	res = i.mcpCall(t, "note_rename", map[string]any{
		"vault": testVault, "oldPath": path, "newPath": "Projects/McpLife2.md", "revision": base,
	})
	if !res.IsError {
		t.Fatalf("stale MCP rename must fail, got %+v", res)
	}
	res = i.mcpCall(t, "note_rename", map[string]any{
		"vault": testVault, "oldPath": path, "newPath": "Projects/McpLife2.md", "revision": cur.Revision,
	})
	if res.IsError {
		t.Fatalf("MCP rename with the current revision must succeed: %+v", res)
	}
	if _, ok := i.noteExists("Projects/McpLife2.md"); !ok {
		t.Fatal("MCP rename must create the target path")
	}
}

// TestStressRepeatedParallelWrites hammers the compare-and-swap across several rounds with a
// wide fan-out. It is the substitute for the race detector, which needs a C toolchain this
// environment does not have (see the report): every round must produce exactly one winner and
// exactly one revision step, which is only possible if the check and the write are atomic.
func TestStressRepeatedParallelWrites(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in -short mode")
	}
	i := newInstance(t)
	path := "Projects/StressRounds.md"
	rev := i.seedNote(path, "base\n")

	const writers = 48
	for round := 0; round < 3; round++ {
		var wg sync.WaitGroup
		results := make([]envelope, writers)
		start := make(chan struct{})
		for n := 0; n < writers; n++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				<-start
				results[n] = i.write(path, fmt.Sprintf("round-%d-writer-%02d\n", round, n), int64p(rev))
			}(n)
		}
		close(start)
		wg.Wait()

		successes, conflicts := 0, 0
		for _, r := range results {
			switch r.Code {
			case 1:
				successes++
			case 452:
				conflicts++
			default:
				t.Fatalf("round %d: unexpected code %d: %s", round, r.Code, r.Message)
			}
		}
		if successes != 1 || conflicts != writers-1 {
			t.Fatalf("round %d: want 1 success and %d conflicts, got %d/%d", round, writers-1, successes, conflicts)
		}
		got := i.getNote(path)
		if got.Revision != rev+1 {
			t.Fatalf("round %d: revision must advance exactly once: want %d, got %d", round, rev+1, got.Revision)
		}
		if !strings.HasPrefix(got.Content, fmt.Sprintf("round-%d-", round)) {
			t.Fatalf("round %d: content must come from this round, got %q", round, got.Content)
		}
		rev = got.Revision
	}
}
