package occtest

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/lxzan/gws"

	"github.com/haierkeys/fast-note-sync-service/pkg/util"
)

// wsConn is a minimal client for the service's sync channel. Frames are text and carry the
// `Action|{json}` envelope, exactly like the Obsidian plugin speaks.
type wsConn struct {
	t      *testing.T
	conn   *gws.Conn
	frames chan string
	closed chan struct{}
}

type wsEvent struct{ c *wsConn }

func (e *wsEvent) OnOpen(*gws.Conn) {}

func (e *wsEvent) OnClose(*gws.Conn, error) {
	select {
	case <-e.c.closed:
	default:
		close(e.c.closed)
	}
}

func (e *wsEvent) OnPing(socket *gws.Conn, payload []byte) { _ = socket.WritePong(payload) }

func (e *wsEvent) OnPong(*gws.Conn, []byte) {}

func (e *wsEvent) OnMessage(_ *gws.Conn, message *gws.Message) {
	body := string(message.Bytes())
	_ = message.Close()
	e.c.frames <- body
}

// dialWS connects, authenticates and announces client info - the same handshake the plugin
// performs. `strategy` selects the WebSocket conflict strategy ("" or "manualMerge").
func (i *instance) dialWS(t *testing.T, strategy string) *wsConn {
	t.Helper()
	c := &wsConn{t: t, frames: make(chan string, 128), closed: make(chan struct{})}
	url := strings.Replace(i.server.URL, "http://", "ws://", 1) +
		"/api/user/sync?client=occapi&clientName=occ-ws&clientVersion=1.0"
	conn, _, err := gws.NewClient(&wsEvent{c: c}, &gws.ClientOption{Addr: url})
	if err != nil {
		t.Fatalf("ws dial: %v", err)
	}
	c.conn = conn
	go conn.ReadLoop()

	c.send("Authorization|" + i.apiToken)
	if _, ok := c.awaitAction("Authorization", 10*time.Second); !ok {
		t.Fatal("ws did not acknowledge Authorization")
	}
	info := map[string]any{
		"name": "occ-ws", "version": "1.0", "type": "occapi", "isLinux": true,
		"offlineSyncStrategy": strategy, "protobuf": false,
	}
	payload, _ := json.Marshal(info)
	c.send("ClientInfo|" + string(payload))
	if _, ok := c.awaitAction("ClientInfo", 10*time.Second); !ok {
		t.Fatal("ws did not acknowledge ClientInfo")
	}
	t.Cleanup(func() { _ = conn.WriteClose(1000, nil) })
	return c
}

func (c *wsConn) send(frame string) {
	c.t.Helper()
	if err := c.conn.WriteMessage(gws.OpcodeText, []byte(frame)); err != nil {
		c.t.Fatalf("ws write: %v", err)
	}
}

// awaitAction returns the decoded envelope of the first frame whose action matches, and
// reports whether the frame carried an action prefix at all.
func (c *wsConn) awaitAction(action string, timeout time.Duration) (map[string]any, bool) {
	c.t.Helper()
	deadline := time.After(timeout)
	for {
		select {
		case frame := <-c.frames:
			if !strings.Contains(frame, "|") {
				continue // untyped frame: not what we are waiting for
			}
			parts := strings.SplitN(frame, "|", 2)
			if parts[0] != action {
				continue
			}
			var env map[string]any
			if err := json.Unmarshal([]byte(parts[1]), &env); err != nil {
				c.t.Fatalf("decode %s frame: %v", action, err)
			}
			return env, true
		case <-deadline:
			return nil, false
		case <-c.closed:
			return nil, false
		}
	}
}

// drainUntyped reports the untyped frames received within the window (used to prove that no
// actionless conflict frame was emitted).
func (c *wsConn) drainUntyped(window time.Duration) []string {
	var out []string
	deadline := time.After(window)
	for {
		select {
		case frame := <-c.frames:
			if !strings.Contains(frame, "|") {
				out = append(out, frame)
			}
		case <-deadline:
			return out
		}
	}
}

// TestWSDeleteIgnoresRevision is finding F1 from the independent review: NoteDelete must stay
// unconditional, so a stale `revision` field in the frame may not make it conditional and may
// not produce the untyped conflict frame that NoteDelete's error path would emit.
func TestWSDeleteIgnoresRevision(t *testing.T) {
	i := newInstance(t)
	path := "Projects/WSDelete.md"
	base := i.seedNote(path, "delete-me\n") // revision 1
	if env := i.write(path, "advanced\n", int64p(base)); env.Code != 1 {
		t.Fatalf("advance note: code=%d %s", env.Code, env.Message)
	}
	live := i.getNote(path) // revision 2
	if live.Revision != base+1 {
		t.Fatalf("expected revision %d, got %d", base+1, live.Revision)
	}

	ws := i.dialWS(t, "")
	body, _ := json.Marshal(map[string]any{
		"vault": testVault, "path": path, "pathHash": hashOf(path),
		"revision": base, // deliberately stale
	})
	ws.send("NoteDelete|" + string(body))

	ack, ok := ws.awaitAction("NoteDeleteAck", 10*time.Second)
	if !ok {
		t.Fatalf("NoteDelete over WS must succeed despite a stale revision field; untyped frames=%v",
			ws.drainUntyped(300*time.Millisecond))
	}
	if code, _ := ack["code"].(float64); int(code) != 1 {
		t.Fatalf("NoteDeleteAck code = %v, want 1", ack["code"])
	}
	if _, exists := i.noteExists(path); exists {
		t.Fatal("NoteDelete must have deleted the note")
	}
	if frames := ws.drainUntyped(300 * time.Millisecond); len(frames) != 0 {
		t.Fatalf("no untyped frames expected, got %v", frames)
	}
}

// TestWSRenameIgnoresRevision is the NoteRename half of F1.
func TestWSRenameIgnoresRevision(t *testing.T) {
	i := newInstance(t)
	src := "Projects/WSRename.md"
	dst := "Projects/WSRename-target.md"
	base := i.seedNote(src, "rename-me\n")
	if env := i.write(src, "advanced\n", int64p(base)); env.Code != 1 {
		t.Fatalf("advance note: code=%d %s", env.Code, env.Message)
	}

	ws := i.dialWS(t, "")
	body, _ := json.Marshal(map[string]any{
		"vault": testVault, "path": dst, "pathHash": hashOf(dst),
		"oldPath": src, "oldPathHash": hashOf(src),
		"revision": base, // deliberately stale
	})
	ws.send("NoteRename|" + string(body))

	ack, ok := ws.awaitAction("NoteRenameAck", 10*time.Second)
	if !ok {
		t.Fatalf("NoteRename over WS must succeed despite a stale revision field; untyped frames=%v",
			ws.drainUntyped(300*time.Millisecond))
	}
	if code, _ := ack["code"].(float64); int(code) != 1 {
		t.Fatalf("NoteRenameAck code = %v, want 1", ack["code"])
	}
	renamed, exists := i.noteExists(dst)
	if !exists {
		t.Fatal("rename target must exist")
	}
	if renamed.Content != "advanced\n" {
		t.Fatalf("rename must carry the current content, got %q", renamed.Content)
	}
	if _, exists := i.noteExists(src); exists {
		t.Fatal("source path must be gone after rename")
	}
}

// TestWSModifyIgnoresRevisionAndKeepsManualMerge pins the untouched WebSocket conflict engine:
// `revision` is ignored on NoteModify, baseHash/manualMerge still answers 530, and resolving the
// conflict still writes.
func TestWSModifyIgnoresRevisionAndKeepsManualMerge(t *testing.T) {
	i := newInstance(t)
	path := "Projects/WSModify.md"
	i.seedNote(path, "base\n")
	current := i.getNote(path)

	ws := i.dialWS(t, "manualMerge")

	// A stale `revision` must not turn this into a REST-style conditional write: the WS engine
	// only looks at baseHash.
	body, _ := json.Marshal(map[string]any{
		"vault": testVault, "path": path, "pathHash": hashOf(path),
		"baseHash": current.Hash, "baseHashMissing": false,
		"content": "from-ws\n", "contentHash": hashOf("from-ws\n"),
		"mtime": current.Mtime + 1000, "ctime": current.Mtime,
		"context": "occ-ws-test", "revision": 999999, // nonsense for REST, ignored here
	})
	ws.send("NoteModify|" + string(body))
	ack, ok := ws.awaitAction("NoteModifyAck", 10*time.Second)
	if !ok {
		t.Fatalf("NoteModify with a bogus revision must still be processed by the WS engine; untyped=%v",
			ws.drainUntyped(300*time.Millisecond))
	}
	if code, _ := ack["code"].(float64); int(code) != 1 {
		t.Fatalf("NoteModifyAck code = %v, want 1", ack["code"])
	}
	if got := i.getNote(path); got.Content != "from-ws\n" {
		t.Fatalf("WS write must be applied, got %q", got.Content)
	}

	// server-side change while the client is "offline"
	mid := i.getNote(path)
	if env := i.write(path, "server-edit\n", int64p(mid.Revision)); env.Code != 1 {
		t.Fatalf("server write: code=%d %s", env.Code, env.Message)
	}

	// stale baseHash + manualMerge must still yield 530 with both versions
	staleBody, _ := json.Marshal(map[string]any{
		"vault": testVault, "path": path, "pathHash": hashOf(path),
		"baseHash": mid.Hash, "baseHashMissing": false,
		"content": "offline-edit\n", "contentHash": hashOf("offline-edit\n"),
		"mtime": mid.Mtime + 5000, "ctime": mid.Mtime, "context": "occ-ws-test",
	})
	ws.send("NoteModify|" + string(staleBody))
	conflict, ok := ws.awaitAction("NoteSyncNeedPush", 10*time.Second)
	if !ok {
		t.Fatal("manualMerge must still answer NoteSyncNeedPush for a stale baseHash")
	}
	if code, _ := conflict["code"].(float64); int(code) != 530 {
		t.Fatalf("conflict code = %v, want 530", conflict["code"])
	}
	data, _ := conflict["data"].(map[string]any)
	if data == nil || data["serverContent"] == nil || data["baseContent"] == nil {
		t.Fatalf("530 payload must carry serverContent and baseContent, got %v", conflict["data"])
	}
	if got := i.getNote(path); got.Content != "server-edit\n" {
		t.Fatalf("a blocked WS push must not overwrite the server copy, got %q", got.Content)
	}
}

// hashOf is the same 32-bit hash the server and the plugin use for pathHash/contentHash.
func hashOf(s string) string { return util.EncodeHash32(s) }
