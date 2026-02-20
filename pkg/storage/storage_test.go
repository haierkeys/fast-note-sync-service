package storage_test

import (
	"testing"

	"github.com/haierkeys/fast-note-sync-service/pkg/storage"
	"github.com/haierkeys/fast-note-sync-service/pkg/storage/local_fs"
)

func TestNewClient_Local(t *testing.T) {
	cfg := &storage.Config{
		Type:     storage.LOCAL,
		SavePath: "./uploads",
	}

	client, err := storage.NewClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create local client: %v", err)
	}

	if client == nil {
		t.Fatal("Client is nil")
	}

	// Verify type
	if _, ok := client.(*local_fs.LocalFS); !ok {
		t.Fatal("Client is not *local_fs.LocalFS")
	}
}

func TestNewClient_Invalid(t *testing.T) {
	cfg := &storage.Config{
		Type: "invalid",
	}

	_, err := storage.NewClient(cfg)
	if err == nil {
		t.Fatal("Expected error for invalid storage type")
	}
}
