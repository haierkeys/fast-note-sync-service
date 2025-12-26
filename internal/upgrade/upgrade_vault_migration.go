package upgrade

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// VaultMigrate 修改 vault 表的 note_count 字段,去掉 NOT NULL 约束
type VaultMigrate struct{}

// Version 返回版本号
func (m *VaultMigrate) Version() string {
	return "0.8.10"
}

// Description 返回描述
func (m *VaultMigrate) Description() string {
	return "Remove NOT NULL constraint from vault.note_count field"
}

// Up 执行升级
func (m *VaultMigrate) Up(db *gorm.DB, ctx context.Context) error {

	// 0. 重命名数据库文件
	global.Logger.Info("Step 0: Renaming database files from db_note_ prefix to db_user_ prefix")
	dbDir := filepath.Dir(global.Config.Database.Path)
	files, err := os.ReadDir(dbDir)
	if err != nil {
		global.Logger.Error("Step 0 failed: unable to read database directory", zap.Error(err))
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if strings.HasPrefix(name, "db_note_") {

			oldPath := filepath.Join(dbDir, name)
			newName := "db_user_" + strings.TrimPrefix(name, "db_note_")
			newPath := filepath.Join(dbDir, newName)

			global.Logger.Info("Renaming file", zap.String("old", name), zap.String("new", newName))
			if err := os.Rename(oldPath, newPath); err != nil {
				// 记录错误但尝试继续，或者根据需求决定是否中断
				// 这里选择如果重命名失败则报错停止，保证数据一致性
				global.Logger.Error("Step 0 failed: unable to rename file", zap.Error(err))
				return fmt.Errorf("failed to rename file %s to %s: %w", name, newName, err)
			}
		}
	}
	global.Logger.Info("Step 0: Database files renamed successfully")

	global.Logger.Info("VaultMigrate Up - Starting vault table migration")
	svc := service.NewBackground(ctx)

	// SQLite 不支持直接 ALTER COLUMN,需要重建表
	// 1. 创建新表
	global.Logger.Info("Step 1: Creating new vault_new table")
	createNewTable := `
		CREATE TABLE vault_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			vault TEXT DEFAULT '',
			note_count INTEGER DEFAULT 0,
			note_size INTEGER DEFAULT 0,
			file_count INTEGER DEFAULT 0,
			file_size INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT NULL,
			updated_at DATETIME DEFAULT NULL
		)
	`
	if err := svc.ExposeExecuteSQL(createNewTable); err != nil {
		global.Logger.Error("Step 1 failed", zap.Error(err))
		return err
	}
	global.Logger.Info("Step 1: vault_new table created successfully")

	// 2. 复制数据
	global.Logger.Info("Step 2: Copying data from vault to vault_new")
	copyData := `
		INSERT INTO vault_new (id, vault, note_count, note_size, file_count, file_size, created_at, updated_at)
		SELECT id, vault, note_count, size as note_size, file_count, file_size, created_at, updated_at
		FROM vault
	`
	if err := svc.ExposeExecuteSQL(copyData); err != nil {
		global.Logger.Error("Step 2 failed", zap.Error(err))
		return err
	}
	global.Logger.Info("Step 2: Data copied successfully")

	// 3. 删除旧表
	global.Logger.Info("Step 3: Dropping old vault table")
	if err := svc.ExposeExecuteSQL("DROP TABLE vault"); err != nil {
		global.Logger.Error("Step 3 failed", zap.Error(err))
		return err
	}

	global.Logger.Info("Step 3: Old vault table dropped successfully")

	// 4. 重命名新表
	global.Logger.Info("Step 4: Renaming vault_new to vault")
	if err := svc.ExposeExecuteSQL("ALTER TABLE vault_new RENAME TO vault"); err != nil {
		global.Logger.Error("Step 4 failed", zap.Error(err))
		return err
	}
	global.Logger.Info("Step 4: Table renamed successfully")

	// 5. 重建索引
	global.Logger.Info("Step 5: Recreating index idx_vault_uid")
	createIndex := `CREATE INDEX idx_vault_uid ON vault (vault ASC)`
	if err := svc.ExposeExecuteSQL(createIndex); err != nil {
		global.Logger.Error("Step 5 failed", zap.Error(err))
		return err
	}
	global.Logger.Info("Step 5: Index recreated successfully")

	global.Logger.Info("VaultMigrate Up - Completed successfully")
	return nil
}
