package retirement

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRetirementExactScopeAndActualBytes(t *testing.T) {
	dir := t.TempDir()
	old := []byte("retired bytes")
	h := sha256.Sum256(old)
	e := Entry{UID: 1, VaultID: 1, Path: "Images/原图.png", SHA256: hex.EncodeToString(h[:])}
	doc, _ := json.Marshal(Document{Version: 1, Entries: []Entry{e}})
	name := filepath.Join(dir, "policy.json")
	if err := os.WriteFile(name, doc, 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Load(name)
	if err != nil {
		t.Fatal(err)
	}
	blob := filepath.Join(dir, "blob")
	if err := os.WriteFile(blob, old, 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		uid, vault int64
		path       string
		want       bool
	}{
		{1, 1, e.Path, true}, {2, 1, e.Path, false}, {1, 2, e.Path, false},
		{1, 1, "Images/another.png", false}, {1, 1, "images/原图.png", false},
	} {
		got, _, err := p.CheckFile(tc.uid, tc.vault, tc.path, blob)
		if err != nil || got != tc.want {
			t.Fatalf("%+v: blocked=%v err=%v", tc, got, err)
		}
	}
	if err := os.WriteFile(blob, []byte("intentional new bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	if blocked, _, err := p.CheckFile(1, 1, e.Path, blob); err != nil || blocked {
		t.Fatalf("new bytes rejected: %v %v", blocked, err)
	}
	if _, _, err := p.CheckFile(1, 1, e.Path, filepath.Join(dir, "missing")); err == nil {
		t.Fatal("missing byte proof must fail closed")
	}
}

func TestRetirementConfigurationDisabledOrInvalid(t *testing.T) {
	t.Setenv(EnvPolicy, "")
	if p, err := LoadFromEnv(); p != nil || err != nil {
		t.Fatal("unset policy must be disabled")
	}
	for _, data := range []string{
		`{"version":2,"entries":[]}`, `{"version":1,"unknown":true,"entries":[]}`,
		`{"version":1,"entries":[]} {}`, `{"version":1,"entries":[{"uid":1,"vault_id":1,"path":"../old","sha256":"123"}]}`,
	} {
		name := filepath.Join(t.TempDir(), "bad.json")
		if err := os.WriteFile(name, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(name); err == nil {
			t.Fatalf("invalid policy accepted: %s", data)
		}
	}
}

func TestRetirementMultipleSHAForExactPath(t *testing.T) {
	first := sha256.Sum256([]byte("original CRLF\r\n"))
	second := sha256.Sum256([]byte("original CRLF\n"))
	e := Entry{UID: 1, VaultID: 1, Path: "Images/old.svg", SHA256: hex.EncodeToString(first[:])}
	variant := e
	variant.SHA256 = hex.EncodeToString(second[:])
	name := filepath.Join(t.TempDir(), "policy.json")
	doc, _ := json.Marshal(Document{Version: 1, Entries: []Entry{e, variant}})
	if err := os.WriteFile(name, doc, 0600); err != nil {
		t.Fatal(err)
	}
	p, err := Load(name)
	if err != nil || !p.RejectsSHA256(1, 1, e.Path, e.SHA256) || !p.RejectsSHA256(1, 1, e.Path, variant.SHA256) {
		t.Fatalf("approved newline variant rejected by loader: %v", err)
	}
	doc, _ = json.Marshal(Document{Version: 1, Entries: []Entry{e, e}})
	if err := os.WriteFile(name, doc, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(name); err == nil {
		t.Fatal("duplicate full tuple accepted")
	}
}
