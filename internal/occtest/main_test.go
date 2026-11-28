package occtest

import (
	"os"
	"testing"
)

// TestMain runs the whole package inside a throwaway directory.
//
// The service resolves note storage as the relative path "storage/vault/u_<uid>/note/n_<id>/…"
// (internal/dao/dao_helper.go), so tests must control the process working directory. Each test
// still chdirs into its own temp dir; this outer directory only guarantees that a background
// goroutine which finishes after its test has cleaned up cannot create "storage/…" inside the
// repository and turn a passing run into a dirty working tree.
func TestMain(m *testing.M) {
	orig, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}
	tmp, tmpErr := os.MkdirTemp("", "occtest-root-")
	if tmpErr == nil {
		_ = os.Chdir(tmp)
	}
	code := m.Run()
	_ = os.Chdir(orig)
	if tmpErr == nil {
		_ = os.RemoveAll(tmp)
	}
	os.Exit(code)
}
