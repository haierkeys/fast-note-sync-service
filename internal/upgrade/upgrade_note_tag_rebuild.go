package upgrade

import (
	"context"
	"fmt"

	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NoteTagIndexRebuildMigrate builds the note_tag index for every existing
// note. The table is created empty by AutoMigrate, so without this backfill
// note_list_tags would report nothing until each note happens to be written
// again. The scan is DB-only (note.content is stored in the database), so it
// does not touch the filesystem or the note body loader.
//
// NoteTagIndexRebuildMigrate 为既有的每篇笔记构建 note_tag 索引。
// AutoMigrate 只建空表，若不回填，note_list_tags 在每篇笔记被重新写入之前都会返回空。
// 该扫描只读数据库（note.content 存于库中），不触碰文件系统与正文加载器。
type NoteTagIndexRebuildMigrate struct{}

// Version returns the migration version
// Version 返回升级版本号
func (m *NoteTagIndexRebuildMigrate) Version() string {
	return "3.6.2"
}

// Description returns the migration description
// Description 返回升级描述
func (m *NoteTagIndexRebuildMigrate) Description() string {
	return "Backfill the note_tag index from existing note content"
}

// Up runs the migration
// Up 执行升级操作
func (m *NoteTagIndexRebuildMigrate) Up(db *gorm.DB, ctx context.Context, mc *MigrationContext) error {
	if mc.Dao == nil {
		return fmt.Errorf("dao is nil in migration context")
	}

	uids, err := mc.Dao.GetAllUserUIDs()
	if err != nil {
		return fmt.Errorf("failed to get all user UIDs: %w", err)
	}

	repo := dao.NewNoteTagRepository(mc.Dao)
	for _, uid := range uids {
		userDB := mc.Dao.ResolveDB("user_" + fmt.Sprintf("%d", uid))
		if userDB == nil {
			mc.Logger.Warn("user db connection is nil, skipping", zap.Int64("uid", uid))
			continue
		}
		if !userDB.Migrator().HasTable("vault") || !userDB.Migrator().HasTable("note") {
			continue
		}

		var vaultIDs []int64
		if err := userDB.WithContext(ctx).Table("vault").
			Where("uid = ? AND is_deleted = ?", uid, false).
			Pluck("id", &vaultIDs).Error; err != nil {
			return fmt.Errorf("failed to list vaults for uid %d: %w", uid, err)
		}

		for _, vaultID := range vaultIDs {
			written, err := repo.RebuildForVault(ctx, vaultID, uid)
			if err != nil {
				return fmt.Errorf("failed to rebuild tag index for uid %d vault %d: %w", uid, vaultID, err)
			}
			mc.Logger.Info("note tag index rebuilt",
				zap.Int64("uid", uid),
				zap.Int64("vaultId", vaultID),
				zap.Int("tags", written))
		}
	}

	return nil
}
