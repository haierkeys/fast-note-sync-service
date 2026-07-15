package upgrade

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BackupRetentionDefaultMigrate backfills existing backup_config rows whose
// retention_days is still 0 (the old runtime default) to 10, matching the
// design intent in scripts/db.sql. RetentionDays == 0 was treated as "keep
// forever" by finishTask, causing unbounded backup history/storage growth.
// BackupRetentionDefaultMigrate 将 retention_days 仍为 0（旧的运行时默认值）的
// backup_config 记录回填为 10，与 scripts/db.sql 的设计意图保持一致。
// RetentionDays == 0 曾被 finishTask 视为"永久保留"，导致备份历史/存储无限增长。
type BackupRetentionDefaultMigrate struct{}

// Version returns the migration version
// Version 返回升级版本号
func (m *BackupRetentionDefaultMigrate) Version() string {
	return "3.5.2"
}

// Description returns the migration description
// Description 返回升级描述
func (m *BackupRetentionDefaultMigrate) Description() string {
	return "Backfill backup_config.retention_days from 0 to 10 (design default)"
}

// Up runs the migration
// Up 执行升级操作
func (m *BackupRetentionDefaultMigrate) Up(db *gorm.DB, ctx context.Context, mc *MigrationContext) error {
	if mc.Dao == nil {
		return fmt.Errorf("dao is nil in migration context")
	}

	uids, err := mc.Dao.GetAllUserUIDs()
	if err != nil {
		return fmt.Errorf("failed to get all user UIDs: %w", err)
	}

	for _, uid := range uids {
		key := "user_backup_" + fmt.Sprintf("%d", uid)
		userDB := mc.Dao.ResolveDB(key)
		if userDB == nil {
			mc.Logger.Warn("user db connection is nil, skipping", zap.Int64("uid", uid))
			continue
		}

		err = userDB.Transaction(func(tx *gorm.DB) error {
			if !tx.Migrator().HasTable("backup_config") {
				return nil
			}
			err := tx.WithContext(ctx).Table("backup_config").Where("retention_days = 0").Update("retention_days", 10).Error
			if err != nil {
				return fmt.Errorf("failed to backfill backup_config retention_days for uid %d: %w", uid, err)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}
