// Package retirement rejects only explicitly retired (user, vault, path, SHA-256) tuples.
package retirement

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

const EnvPolicy = "FNS_RETIRED_BYTES_POLICY"

type Entry struct {
	UID     int64  `json:"uid"`
	VaultID int64  `json:"vault_id"`
	Path    string `json:"path"`
	SHA256  string `json:"sha256"`
}

type Document struct {
	Version int     `json:"version"`
	Entries []Entry `json:"entries"`
}

type key struct {
	uid, vault int64
	path       string
}

// Policy is immutable after loading. Only the audit file is written at runtime.
// A restore audit entry is evidence of an explicit operation, never a permanent exemption.
type Policy struct {
	entries   map[key]map[string]struct{}
	auditPath string
	auditMu   sync.Mutex
	count     int
}

func LoadFromEnv() (*Policy, error) {
	name := os.Getenv(EnvPolicy)
	if name == "" {
		return nil, nil
	}
	return Load(name)
}

func Load(name string) (*Policy, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("retirement policy open: %w", err)
	}
	defer f.Close()
	var doc Document
	d := json.NewDecoder(f)
	d.DisallowUnknownFields()
	if err := d.Decode(&doc); err != nil {
		return nil, fmt.Errorf("retirement policy decode: %w", err)
	}
	if err := d.Decode(new(any)); !errors.Is(err, io.EOF) {
		return nil, errors.New("retirement policy has trailing JSON")
	}
	if doc.Version != 1 {
		return nil, errors.New("retirement policy version must be 1")
	}
	p := &Policy{entries: make(map[key]map[string]struct{}), auditPath: name + ".audit.jsonl"}
	for i, e := range doc.Entries {
		if e.UID <= 0 || e.VaultID <= 0 || e.Path == "" || path.IsAbs(e.Path) || path.Clean(e.Path) != e.Path || strings.Contains(e.Path, "\\") || e.Path == ".." || strings.HasPrefix(e.Path, "../") {
			return nil, fmt.Errorf("retirement policy entry %d has invalid scope/path", i)
		}
		h, err := hex.DecodeString(e.SHA256)
		if err != nil || len(h) != sha256.Size || e.SHA256 != strings.ToLower(e.SHA256) {
			return nil, fmt.Errorf("retirement policy entry %d has invalid SHA-256", i)
		}
		k := key{e.UID, e.VaultID, e.Path}
		if p.entries[k] == nil {
			p.entries[k] = make(map[string]struct{})
		}
		if _, exists := p.entries[k][e.SHA256]; exists {
			return nil, fmt.Errorf("retirement policy entry %d duplicates a tuple", i)
		}
		p.entries[k][e.SHA256] = struct{}{}
		p.count++
	}
	return p, nil
}

func (p *Policy) Count() int {
	if p == nil {
		return 0
	}
	return p.count
}

func (p *Policy) Covers(uid, vault int64, name string) bool {
	if p == nil {
		return false
	}
	_, ok := p.entries[key{uid, vault, name}]
	return ok
}

func SHA256File(name string) (string, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (p *Policy) CheckFile(uid, vault int64, name, bytesPath string) (blocked bool, digest string, err error) {
	if !p.Covers(uid, vault, name) {
		return false, "", nil
	}
	digest, err = SHA256File(bytesPath)
	if err != nil {
		return false, "", fmt.Errorf("retirement byte verification: %w", err)
	}
	blocked = p.RejectsSHA256(uid, vault, name, digest)
	return blocked, digest, nil
}

// RejectsSHA256 accepts a digest calculated by the server from actual bytes.
// Callers must never pass the client's 32-bit contentHash here.
func (p *Policy) RejectsSHA256(uid, vault int64, name, digest string) bool {
	if p == nil {
		return false
	}
	_, blocked := p.entries[key{uid, vault, name}][digest]
	return blocked
}

// Audit writes and fsyncs authorization before Restore changes the DB. Results
// are appended separately. Failure to record authorization must abort Restore.
func (p *Policy) Audit(uid, vault, fileID int64, name, digest, event, clientType, clientName, clientVersion string) error {
	if !p.Covers(uid, vault, name) {
		return nil
	}
	record := struct {
		Time          string `json:"time_utc"`
		UID           int64  `json:"uid"`
		VaultID       int64  `json:"vault_id"`
		FileID        int64  `json:"file_id"`
		Path          string `json:"path"`
		SHA256        string `json:"sha256"`
		Event         string `json:"event"`
		ClientType    string `json:"client_type"`
		ClientName    string `json:"client_name"`
		ClientVersion string `json:"client_version"`
	}{time.Now().UTC().Format(time.RFC3339Nano), uid, vault, fileID, name, digest, event, clientType, clientName, clientVersion}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	p.auditMu.Lock()
	defer p.auditMu.Unlock()
	f, err := os.OpenFile(p.auditPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(data, '\n')); err != nil {
		return err
	}
	return f.Sync()
}
