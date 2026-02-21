package upgrade

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/service"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GitSyncConfigMigrate 增加 git_sync_config 表
type GitSyncConfigMigrate struct{}

// Version 返回版本号
func (m *GitSyncConfigMigrate) Version() string {
	return "0.8.11"
}

// Description 返回描述
func (m *GitSyncConfigMigrate) Description() string {
	return "Add git_sync_config table for git repository sync task"
}

// Up 执行升级
func (m *GitSyncConfigMigrate) Up(db *gorm.DB, ctx context.Context, mc *MigrationContext) error {
	logger := mc.Logger
	logger.Info("GitSyncConfigMigrate Up - Starting git_sync_config table creation")
	dbUtils := service.NewDBUtils(db, ctx)

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS "git_sync_config" (
		"id" integer PRIMARY KEY AUTOINCREMENT,
		"uid" integer NOT NULL DEFAULT 0,
		"vault_id" integer NOT NULL DEFAULT 0,
		"repo_url" text DEFAULT '',
		"username" text DEFAULT '',
		"password" text DEFAULT '',
		"branch" text DEFAULT 'main',
		"is_enabled" integer DEFAULT 0,
		"delay" integer DEFAULT 0,
		"last_sync_time" datetime DEFAULT NULL,
		"last_status" integer DEFAULT 0,
		"last_message" text DEFAULT '',
		"created_at" datetime DEFAULT NULL,
		"updated_at" datetime DEFAULT NULL
	);
	`
	// 我们将此配置放在用户数据库中，与 BackupConfig 类似的逻辑
	// 如果需要全局管理，则应放在主库。但根据项目目前的模式，用户配置通常在用户库。
	if err := dbUtils.UserExecuteSQL(createTableSQL); err != nil {
		logger.Error("GitSyncConfigMigrate Up - Failed to create table", zap.Error(err))
		return err
	}

	createIndexSQL := `CREATE INDEX IF NOT EXISTS "idx_git_sync_config_uid" ON "git_sync_config" ("uid");`
	if err := dbUtils.UserExecuteSQL(createIndexSQL); err != nil {
		logger.Error("GitSyncConfigMigrate Up - Failed to create index", zap.Error(err))
		return err
	}

	logger.Info("GitSyncConfigMigrate Up - Completed successfully")
	return nil
}
