package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/upgrade"
	"go.uber.org/zap"
)

func main() {
	// 1. Setup environment
	dbDir := "storage/database"
	_ = os.MkdirAll(dbDir, 0755)

	oldFile := filepath.Join(dbDir, "db_user_note_history999.sqlite3")
	_ = os.WriteFile(oldFile, []byte("dummy data"), 0644)
	fmt.Printf("Created dummy file: %s\n", oldFile)

	// 2. Mock global config and logger
	global.Config = &global.config{
		Database: global.Database{
			Type: "sqlite",
			Path: filepath.Join(dbDir, "db.sqlite3"),
		},
	}
	global.Logger, _ = zap.NewDevelopment()

	// 3. Run migration logic
	migrate := &upgrade.NoteHistoryRenameMigrate{}
	err := migrate.Up(nil, context.Background())
	if err != nil {
		fmt.Printf("Migration failed: %v\n", err)
		os.Exit(1)
	}

	// 4. Verify results
	newFile := filepath.Join(dbDir, "db_user_note_history_999.sqlite3")
	if _, err := os.Stat(newFile); err == nil {
		fmt.Printf("Success! File renamed to: %s\n", newFile)
		// Clean up
		_ = os.Remove(newFile)
	} else {
		fmt.Printf("Failure! New file not found: %s\n", newFile)
		// Clean up old if it still exists
		_ = os.Remove(oldFile)
		os.Exit(1)
	}
}
