package oidc

import (
	"testing"
	"time"
)

func TestStateStoreConsumeRemovesState(t *testing.T) {
	store := NewStateStore(time.Minute)
	state := State{
		State:        "state-1",
		Nonce:        "nonce-1",
		CodeVerifier: "verifier-1",
		RedirectTo:   "/webgui/",
	}

	store.Save(state)
	got, ok := store.Consume("state-1")
	if !ok {
		t.Fatal("Consume() ok = false, want true")
	}
	if got.Nonce != "nonce-1" || got.CodeVerifier != "verifier-1" || got.RedirectTo != "/webgui/" {
		t.Fatalf("Consume() = %#v", got)
	}

	if _, ok := store.Consume("state-1"); ok {
		t.Fatal("Consume() ok = true after first consume, want false")
	}
}

func TestStateStoreConsumeRejectsExpiredState(t *testing.T) {
	store := NewStateStore(10 * time.Millisecond)
	store.Save(State{
		State:        "state-1",
		Nonce:        "nonce-1",
		CodeVerifier: "verifier-1",
	})

	time.Sleep(20 * time.Millisecond)
	if _, ok := store.Consume("state-1"); ok {
		t.Fatal("Consume() ok = true for expired state, want false")
	}
}
