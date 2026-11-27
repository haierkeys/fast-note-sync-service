package occtest

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"

	internalApp "github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/internal/routers"
)

// mcpResult is the subset of an MCP tool result the tests assert on.
type mcpResult struct {
	IsError  bool
	Text     string
	Revision int64
}

// mcpSession is a minimal MCP StreamableHTTP client: initialize once, then tools/call.
type mcpSession struct {
	url, token, session string
}

func (s *mcpSession) rpc(t *testing.T, method string, params map[string]any) map[string]any {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": method, "params": params,
	})
	if err != nil {
		t.Fatalf("marshal rpc: %v", err)
	}
	req, err := http.NewRequest("POST", s.url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new rpc request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("x-client", "occmcp")
	if s.session != "" {
		req.Header.Set("Mcp-Session-Id", s.session)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("rpc %s: %v", method, err)
	}
	defer resp.Body.Close()
	if sid := resp.Header.Get("Mcp-Session-Id"); sid != "" {
		s.session = sid
	}
	raw, _ := io.ReadAll(resp.Body)
	payload := string(raw)
	// StreamableHTTP may answer as server-sent events; the JSON is on the data: line.
	if strings.HasPrefix(strings.TrimSpace(payload), "event:") {
		for _, line := range strings.Split(payload, "\n") {
			if strings.HasPrefix(line, "data:") {
				payload = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
				break
			}
		}
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatalf("decode rpc %s (%q): %v", method, payload, err)
	}
	return out
}

// mcpCall invokes one MCP tool with a fresh session.
func (i *instance) mcpCall(t *testing.T, tool string, args map[string]any) mcpResult {
	t.Helper()
	s := &mcpSession{url: i.server.URL + "/api/mcp", token: i.mcpToken}
	init := s.rpc(t, "initialize", map[string]any{
		"protocolVersion": "2025-06-18",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "occtest", "version": "1.0"},
	})
	if _, ok := init["result"]; !ok {
		t.Fatalf("mcp initialize failed: %v", init)
	}

	res := s.rpc(t, "tools/call", map[string]any{"name": tool, "arguments": args})
	result, ok := res["result"].(map[string]any)
	if !ok {
		t.Fatalf("mcp tools/call %s: unexpected response %v", tool, res)
	}
	out := mcpResult{}
	if v, ok := result["isError"].(bool); ok {
		out.IsError = v
	}
	// content[0].text carries either the JSON error payload or the human summary; the
	// structured output (if any) sits next to it.
	if content, ok := result["content"].([]any); ok && len(content) > 0 {
		if first, ok := content[0].(map[string]any); ok {
			if txt, ok := first["text"].(string); ok {
				out.Text = txt
			}
		}
	}
	if so, ok := result["structuredContent"].(map[string]any); ok {
		if note, ok := so["note"].(map[string]any); ok {
			if rev, ok := note["revision"].(float64); ok {
				out.Revision = int64(rev)
			}
		}
		if newNote, ok := so["newNote"].(map[string]any); ok {
			if rev, ok := newNote["revision"].(float64); ok {
				out.Revision = int64(rev)
			}
		}
	}
	return out
}

// restart shuts the running instance down and boots a new one against the same database and
// storage directory, imitating a service restart (the acceptance row "restart and restore").
func (i *instance) restart() {
	i.t.Helper()
	i.server.Close()
	_ = i.app.Close()
	if sqlDB, err := i.db.DB(); err == nil {
		_ = sqlDB.Close()
	}

	cfgPath := filepath.Join(i.dir, "config.yaml")
	cfg, _, err := internalApp.LoadConfig(cfgPath)
	if err != nil {
		i.t.Fatalf("restart: load config: %v", err)
	}
	logger := newNopLogger()
	db, err := dao.NewEngine(cfg.Database, logger)
	if err != nil {
		i.t.Fatalf("restart: open db: %v", err)
	}
	appContainer, err := internalApp.NewApp(cfg, logger, db, embed.FS{})
	if err != nil {
		i.t.Fatalf("restart: app container: %v", err)
	}
	uni := ut.New(en.New(), en.New(), zh.New())
	router := routers.NewRouter(embed.FS{}, appContainer, uni)
	srv := httptest.NewServer(router)

	i.db, i.app, i.router, i.server = db, appContainer, router, srv
	i.t.Cleanup(func() {
		srv.Close()
		_ = appContainer.Close()
		if sqlDB, e := db.DB(); e == nil {
			_ = sqlDB.Close()
		}
	})
	_ = gin.ReleaseMode
	_ = fmt.Sprint
}
