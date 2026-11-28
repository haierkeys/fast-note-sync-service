// Package occtest holds the optimistic-concurrency regression suite.
//
// It lives in its own package on purpose: this project has several *_test.go files in
// internal/dao and internal/routers/api_router that import the cgo sqlite driver
// (gorm.io/driver/sqlite → mattn/go-sqlite3), so those packages cannot be compiled on a host
// without a C toolchain. This package only uses the production pure-Go driver
// (github.com/glebarez/sqlite) and boots the real application container, the real router and
// a real SQLite database, so the tests exercise the same code path as a deployed service.
package occtest

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite" // pure-Go sqlite driver, same one the service uses
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"embed"
	internalApp "github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/internal/routers"
	"github.com/haierkeys/fast-note-sync-service/internal/upgrade"
)

const testVault = "Projects"

// envelope mirrors the service's response envelope: the HTTP status is always accompanied by
// an application-level code, so tests assert on both.
type envelope struct {
	Code    int             `json:"code"`
	Status  bool            `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type instance struct {
	t      *testing.T
	app    *internalApp.App
	db     *gorm.DB
	router *gin.Engine
	server *httptest.Server
	dir    string
	// tokens
	apiToken string // scope p:rest,ws c:occapi f:*
	mcpToken string // scope p:mcp c:occmcp f:*
	webToken string // scope p:rest c:webgui f:* (admin surface: vault/token management)
}

type noteView struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Revision int64  `json:"revision"`
	Version  int64  `json:"version"`
	Mtime    int64  `json:"mtime"`
	Hash     string `json:"contentHash"`
}

// newInstance boots a full service against a throwaway SQLite database.
func newInstance(t *testing.T) *instance {
	t.Helper()
	gin.SetMode(gin.ReleaseMode)

	// Not t.TempDir(): the service starts background goroutines (vault size counters, link
	// index, FTS, delayed history) that may still write for a few milliseconds after Close(),
	// and t.TempDir's single RemoveAll would then fail with ENOTEMPTY and fail an otherwise
	// passing test. Use a manual directory removed with retries (registered first so it runs
	// last, after the app has been closed).
	dir, mkErr := os.MkdirTemp("", "occtest-")
	if mkErr != nil {
		t.Fatalf("create temp dir: %v", mkErr)
	}
	t.Cleanup(func() { removeWithRetry(t, dir) })

	// The note store resolves paths relative to the working directory
	// ("storage/vault/u_<uid>/note/n_<id>/content.txt"), so run inside the temp dir to keep
	// the repository clean and every test self-contained.
	t.Chdir(dir)
	cfgPath := filepath.Join(dir, "config.yaml")
	cfgYAML := fmt.Sprintf(`server:
  run-mode: release
  http-port: "127.0.0.1:0"
app:
  temp-path: %q
database:
  type: sqlite
  path: %q
  auto-migrate: true
  enable-write-queue: true
log:
  level: error
  file: %q
  production: false
security:
  auth-token-key: "occ-test-token-key"
  token-expiry: "30d"
user:
  register-is-enable: true
  admin-uid: 0
`,
		filepath.Join(dir, "temp"),
		filepath.Join(dir, "db", "db.sqlite3"),
		filepath.Join(dir, "logs", "log.log"),
	)
	if err := os.WriteFile(cfgPath, []byte(cfgYAML), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	for _, d := range []string{filepath.Join(dir, "temp"), filepath.Join(dir, "db"), filepath.Join(dir, "logs")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	cfg, _, err := internalApp.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	logger := newNopLogger()

	db, err := dao.NewEngine(cfg.Database, logger)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := upgrade.Execute(db, logger, internalApp.Version, &cfg.Database, &cfg.UserDatabase); err != nil {
		t.Fatalf("upgrade: %v", err)
	}
	appContainer, err := internalApp.NewApp(cfg, logger, db, embed.FS{})
	if err != nil {
		t.Fatalf("app container: %v", err)
	}

	uni := ut.New(en.New(), en.New(), zh.New())
	router := routers.NewRouter(embed.FS{}, appContainer, uni)
	srv := httptest.NewServer(router)

	inst := &instance{t: t, app: appContainer, db: db, router: router, server: srv, dir: dir}
	t.Cleanup(func() {
		srv.Close()
		_ = appContainer.Close()
		if sqlDB, e := db.DB(); e == nil {
			_ = sqlDB.Close()
		}
		// give fire-and-forget goroutines a moment to finish their last write
		time.Sleep(100 * time.Millisecond)
	})

	inst.bootstrap()
	return inst
}

// bootstrap registers the first user (which self-closes registration), creates the test vault
// and issues one token per credential shape the tests need.
func (i *instance) bootstrap() {
	i.t.Helper()

	reg := i.doRaw("POST", "/api/user/register", map[string]any{
		"username": "occadmin", "password": "Test-Pass-123!",
		"confirmPassword": "Test-Pass-123!", "email": "occ@example.invalid",
	}, "", "webgui")
	if reg.Code != 1 {
		i.t.Fatalf("register: code=%d msg=%s", reg.Code, reg.Message)
	}
	var regData struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(reg.Data, &regData)
	i.webToken = regData.Token

	if r := i.do("POST", "/api/vault", map[string]any{"vault": testVault}, i.webToken, "webgui"); r.Code != 1 && r.Code != 2 {
		i.t.Fatalf("create vault: code=%d msg=%s", r.Code, r.Message)
	}

	issue := func(clientType, protocol, client, function string) string {
		body := map[string]any{
			"clientType": clientType, "vaults": testVault, "expiredDays": 30,
		}
		if protocol != "" {
			body["protocol"] = protocol
			body["client"] = client
			body["function"] = function
		} else {
			body["scope"] = "rest,ws"
		}
		r := i.do("POST", "/api/token", body, i.webToken, "webgui")
		if r.Code != 1 {
			i.t.Fatalf("issue token: code=%d msg=%s", r.Code, r.Message)
		}
		var d struct {
			Token string `json:"token"`
		}
		_ = json.Unmarshal(r.Data, &d)
		return d.Token
	}
	i.apiToken = issue("occapi", "", "", "")
	i.mcpToken = issue("occmcp", "mcp", "occmcp", "*")
}

func (i *instance) doRaw(method, path string, body any, token, client string) envelope {
	i.t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			i.t.Fatalf("marshal body: %v", err)
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, i.server.URL+path, rdr)
	if err != nil {
		i.t.Fatalf("new request: %v", err)
	}
	// Only set Content-Type when a body is present: gin binds bodyless DELETE/GET requests
	// from the query string only when no JSON content type is announced.
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("token", token)
	}
	req.Header.Set("x-client", client)
	req.Header.Set("User-Agent", "occ-test/1.0")

	resp, err := i.server.Client().Do(req)
	if err != nil {
		i.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		i.t.Fatalf("%s %s: decode %q: %v", method, path, string(raw), err)
	}
	env.Message = fmt.Sprintf("%s (http=%d)", env.Message, resp.StatusCode)
	return env
}

func (i *instance) do(method, path string, body any, token, client string) envelope {
	return i.doRaw(method, path, body, token, client)
}

// getNote reads a note through the REST API with the API token.
func (i *instance) getNote(path string) noteView {
	i.t.Helper()
	env := i.do("GET", "/api/note?vault="+testVault+"&path="+urlQueryEscape(path), nil, i.apiToken, "occapi")
	if env.Code != 1 {
		i.t.Fatalf("get note %s: code=%d %s", path, env.Code, env.Message)
	}
	var n noteView
	if err := json.Unmarshal(env.Data, &n); err != nil {
		i.t.Fatalf("decode note: %v", err)
	}
	return n
}

func (i *instance) noteExists(path string) (noteView, bool) {
	env := i.do("GET", "/api/note?vault="+testVault+"&path="+urlQueryEscape(path), nil, i.apiToken, "occapi")
	if env.Code != 1 {
		return noteView{}, false
	}
	var n noteView
	_ = json.Unmarshal(env.Data, &n)
	return n, true
}

// write creates/updates a note. revision(nil) = legacy unconditional write.
func (i *instance) write(path, content string, revision *int64) envelope {
	body := map[string]any{
		"vault": testVault, "path": path, "content": content,
		"mtime": nowMillis(), "ctime": nowMillis(),
	}
	if revision != nil {
		body["revision"] = *revision
	}
	return i.do("POST", "/api/note", body, i.apiToken, "occapi")
}

// seedNote creates a note through the API and returns its initial revision.
func (i *instance) seedNote(path, content string) int64 {
	i.t.Helper()
	env := i.write(path, content, nil)
	if env.Code != 1 {
		i.t.Fatalf("seed %s: code=%d %s", path, env.Code, env.Message)
	}
	return i.getNote(path).Revision
}

// userDB opens the per-user SQLite database that holds the note rows (the service shards
// note metadata into db_user_<uid>.sqlite3 even with a single configured database).
func (i *instance) userDB() *sql.DB {
	i.t.Helper()
	path := filepath.Join(i.dir, "storage", "database", "db_user_1.sqlite3")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		i.t.Fatalf("open user db %s: %v", path, err)
	}
	i.t.Cleanup(func() { _ = db.Close() })
	return db
}

// openDB opens one of the per-domain SQLite shards the service creates for a user
// (e.g. db_user_note_history_1.sqlite3) so tests can assert on or seed internal state.
func (i *instance) openDB(name string) *sql.DB {
	i.t.Helper()
	path := filepath.Join(i.dir, "storage", "database", name)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		i.t.Fatalf("open %s: %v", name, err)
	}
	i.t.Cleanup(func() { _ = db.Close() })
	return db
}

func (i *instance) contentFile(path string) string {
	i.t.Helper()
	var id int64
	if err := i.userDB().QueryRow(`SELECT id FROM note WHERE path = ?`, path).Scan(&id); err != nil || id == 0 {
		i.t.Fatalf("contentFile: note row for %s: %v (id=%d)", path, err, id)
	}
	file := filepath.Join(i.dir, "storage", "vault", "u_1", "note", fmt.Sprintf("n_%d", id), "content.txt")
	b, err := os.ReadFile(file)
	if err != nil {
		i.t.Fatalf("read content file %s: %v", file, err)
	}
	return string(b)
}

func int64p(v int64) *int64 { return &v }

func newNopLogger() *zap.Logger { return zap.NewNop() }

// removeWithRetry deletes a test directory, tolerating writes that a background goroutine may
// perform just after the application container has been closed.
func removeWithRetry(t *testing.T, dir string) {
	t.Helper()
	var err error
	for i := 0; i < 60; i++ {
		if err = os.RemoveAll(dir); err == nil {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Logf("warning: could not remove temp dir %s: %v", dir, err)
}

func urlQueryEscape(s string) string { return url.QueryEscape(s) }

func nowMillis() int64 { return time.Now().UnixMilli() }
