package oauth

import (
	"errors"
	"testing"
)

func TestMapOAuthScopesToFNS_scopesGenerateMCP3DRBACScope(t *testing.T) {
	got, err := MapOAuthScopesToFNS("cursor", []string{"notes:read", "notes:write", "files:read", "files:write"})
	if err != nil {
		t.Fatalf("MapOAuthScopesToFNS() error = %v", err)
	}

	want := "p:mcp c:cursor f:note_r,note_w,file_r,file_w"
	if got != want {
		t.Fatalf("MapOAuthScopesToFNS() = %q, want %q", got, want)
	}
}

func TestMapOAuthScopesToFNS_vaultReadGrantsNoteRead(t *testing.T) {
	got, err := MapOAuthScopesToFNS("claude", []string{"vaults:read"})
	if err != nil {
		t.Fatalf("MapOAuthScopesToFNS() error = %v", err)
	}

	want := "p:mcp c:claude f:note_r"
	if got != want {
		t.Fatalf("MapOAuthScopesToFNS() = %q, want %q", got, want)
	}
}

func TestMapOAuthScopesToFNS_unknownScopeIsInsufficientScope(t *testing.T) {
	_, err := MapOAuthScopesToFNS("cursor", []string{"notes:delete"})
	if !errors.Is(err, ErrInsufficientScope) {
		t.Fatalf("MapOAuthScopesToFNS() error = %v, want ErrInsufficientScope", err)
	}
}
