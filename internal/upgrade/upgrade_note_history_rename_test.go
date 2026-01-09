package upgrade

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/global"
	"go.uber.org/zap"
)

func TestNoteHistoryRenameMigrate_Up(t *testing.T) {
	// 1. Setup environment
	dbDir := t.TempDir()

	// Create several files to test matching logic
	oldFile1 := filepath.Join(dbDir, "db_user_note_history1.sqlite3")
	oldFile2 := filepath.Join(dbDir, "db_user_note_history123.sqlite3")
	otherFile := filepath.Join(dbDir, "db_user_note_history_already_has_underscore.sqlite3")
	unrelatedFile := filepath.Join(dbDir, "db.sqlite3")
	unrelatedUserFile := filepath.Join(dbDir, "db_user_1.sqlite3")

	_ = os.WriteFile(oldFile1, []byte("dummy1"), 0644)
	_ = os.WriteFile(oldFile2, []byte("dummy2"), 0644)
	_ = os.WriteFile(otherFile, []byte("dummy3"), 0644)
	_ = os.WriteFile(unrelatedFile, []byte("dummy4"), 0644)
	_ = os.WriteFile(unrelatedUserFile, []byte("dummy5"), 0644)

	// 2. Mock global config using ConfigLoad
	configContent := fmt.Sprintf(`
database:
  type: sqlite
  path: %s
`, filepath.ToSlash(filepath.Join(dbDir, "db.sqlite3")))
	configPath := filepath.Join(dbDir, "config.yaml")
	_ = os.WriteFile(configPath, []byte(configContent), 0644)

	_, err := global.ConfigLoad(configPath)
	if err != nil {
		t.Fatalf("ConfigLoad failed: %v", err)
	}
	global.Logger, _ = zap.NewDevelopment()

	// 3. Run migration logic
	migrate := &NoteHistoryRenameMigrate{}
	err = migrate.Up(nil, context.Background())
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// 4. Verify results
	tests := []struct {
		path   string
		exists bool
	}{
		{filepath.Join(dbDir, "db_user_note_history_1.sqlite3"), true},
		{filepath.Join(dbDir, "db_user_note_history_123.sqlite3"), true},
		{filepath.Join(dbDir, "db_user_note_history1.sqlite3"), false},
		{filepath.Join(dbDir, "db_user_note_history123.sqlite3"), false},
		{otherFile, true},
		{unrelatedFile, true},
		{unrelatedUserFile, true},
	}

	for _, tt := range tests {
		_, err := os.Stat(tt.path)
		exists := !os.IsNotExist(err)
		if exists != tt.exists {
			t.Errorf("File %s exists=%v, want %v", tt.path, exists, tt.exists)
		}
	}
}
