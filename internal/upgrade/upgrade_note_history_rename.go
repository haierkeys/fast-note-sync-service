package upgrade

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NoteHistoryRenameMigrate 重命名 user_note_history 数据库文件
// Rename user_note_history database files
type NoteHistoryRenameMigrate struct{}

// Version 返回版本号
func (m *NoteHistoryRenameMigrate) Version() string {
	return "1.6.4"
}

// Description 返回描述
func (m *NoteHistoryRenameMigrate) Description() string {
	return "Rename user_note_history[UID] to user_note_history_[UID] in database filenames"
}

// Up 执行升级
func (m *NoteHistoryRenameMigrate) Up(db *gorm.DB, ctx context.Context, mc *MigrationContext) error {
	logger := mc.Logger

	if mc.DatabaseType != "sqlite" {
		logger.Info("NoteHistoryRenameMigrate: Not using SQLite, skipping file rename")
		return nil
	}

	dbPath := mc.DatabasePath
	dbDir := filepath.Dir(dbPath)
	ext := filepath.Ext(dbPath)
	baseName := strings.TrimSuffix(filepath.Base(dbPath), ext)

	files, err := os.ReadDir(dbDir)
	if err != nil {
		logger.Error("NoteHistoryRenameMigrate: Failed to read database directory", zap.Error(err))
		return err
	}

	// 匹配模式: (baseName)_user_note_history(数字)(扩展名)
	// Pattern: (baseName)_user_note_history(digits)(ext)
	// 且不包含下划线分隔数字
	re := regexp.MustCompile(fmt.Sprintf(`^%s_user_note_history(\d+)%s`, regexp.QuoteMeta(baseName), regexp.QuoteMeta(ext)))

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		matches := re.FindStringSubmatch(name)
		if len(matches) > 1 {
			uid := matches[1]
			oldPath := filepath.Join(dbDir, name)
			newName := fmt.Sprintf("%s_user_note_history_%s%s", baseName, uid, ext)
			newPath := filepath.Join(dbDir, newName)

			logger.Info("Renaming user_note_history database file",
				zap.String("old", name),
				zap.String("new", newName))

			if err := os.Rename(oldPath, newPath); err != nil {
				logger.Error("NoteHistoryRenameMigrate: Failed to rename file",
					zap.String("old", oldPath),
					zap.String("new", newPath),
					zap.Error(err))
				return fmt.Errorf("failed to rename %s to %s: %w", name, newName, err)
			}
		}
	}

	return nil
}
