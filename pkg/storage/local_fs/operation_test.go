package local_fs

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLocalFS_SendFile(t *testing.T) {
	// Setup temporary directory
	tempDir := t.TempDir()

	// Create LocalFS client
	client, err := NewClient(&Config{
		SavePath: tempDir,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Prepare data
	filename := "test_file.txt"
	content := "hello world"
	modTime := time.Date(2023, 10, 1, 12, 0, 0, 0, time.UTC)
	reader := strings.NewReader(content)

	// Call SendFile
	savedPath, err := client.SendFile(filename, reader, "text/plain", modTime)
	if err != nil {
		t.Fatalf("SendFile failed: %v", err)
	}

	// Verify file existence
	if _, err := os.Stat(savedPath); os.IsNotExist(err) {
		t.Fatalf("File not found at %s", savedPath)
	}

	// Verify content
	savedContent, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}
	if string(savedContent) != content {
		t.Errorf("Content mismatch: expected %s, got %s", content, string(savedContent))
	}

	// Verify modification time
	fileInfo, err := os.Stat(savedPath)
	if err != nil {
		t.Fatalf("Failed to stat saved file: %v", err)
	}
	// Allow small difference due to filesystem precision/delays, but here we set explicit time
	// Windows/Linux precision differs, but usually it should be close enough or exact
	if !fileInfo.ModTime().Equal(modTime) {
		// Some filesystems might have different precision, check difference
		diff := fileInfo.ModTime().Sub(modTime)
		if diff < -time.Second || diff > time.Second {
			t.Errorf("ModTime mismatch: expected %v, got %v (diff %v)", modTime, fileInfo.ModTime(), diff)
		}
	}
}

func TestLocalFS_SendContent(t *testing.T) {
	// Setup temporary directory
	tempDir := t.TempDir()

	// Create LocalFS client
	client, err := NewClient(&Config{
		SavePath: tempDir,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Prepare data
	// Test with a subdirectory to ensure SendContent creates directories
	filename := "subdir/test_content.txt"
	content := []byte("hello content")
	modTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	// Call SendContent
	savedPath, err := client.SendContent(filename, content, modTime)
	if err != nil {
		t.Fatalf("SendContent failed: %v", err)
	}

	// Verify file existence
	if _, err := os.Stat(savedPath); os.IsNotExist(err) {
		t.Fatalf("File not found at %s", savedPath)
	}

	// Verify content
	savedContent, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}
	if !bytes.Equal(savedContent, content) {
		t.Errorf("Content mismatch: expected %s, got %s", content, string(savedContent))
	}

	// Verify modification time
	fileInfo, err := os.Stat(savedPath)
	if err != nil {
		t.Fatalf("Failed to stat saved file: %v", err)
	}
	if !fileInfo.ModTime().Equal(modTime) {
		diff := fileInfo.ModTime().Sub(modTime)
		if diff < -time.Second || diff > time.Second {
			t.Errorf("ModTime mismatch: expected %v, got %v (diff %v)", modTime, fileInfo.ModTime(), diff)
		}
	}
}
