package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

type backupRepository struct {
	dao *Dao
}

// NewBackupRepository 创建 BackupRepository 实例
func NewBackupRepository(dao *Dao) domain.BackupRepository {
	return &backupRepository{dao: dao}
}

func (r *backupRepository) GetKey(uid int64) string {
	return "user_backup_" + fmt.Sprintf("%d", uid)
}

func (r *backupRepository) backup(uid int64) *query.Query {
	return r.dao.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "BackupConfig")
		model.AutoMigrate(g, "BackupHistory")
	}, r.GetKey(uid)+"#backup", r.GetKey(uid))
}

func (r *backupRepository) configToDomain(m *model.BackupConfig) *domain.BackupConfig {
	if m == nil {
		return nil
	}
	return &domain.BackupConfig{
		ID:             m.ID,
		UID:            m.UID,
		VaultID:        m.VaultID,
		Type:           m.Type,
		StorageIds:     m.StorageIds,
		IsEnabled:      m.IsEnabled == 1,
		CronStrategy:   m.CronStrategy,
		CronExpression: m.CronExpression,
		RetentionDays:  int(m.RetentionDays),
		LastRunTime:    m.LastRunTime,
		NextRunTime:    m.NextRunTime,
		LastStatus:     int(m.LastStatus),
		LastMessage:    m.LastMessage,
		CreatedAt:      time.Time(m.CreatedAt),
		UpdatedAt:      time.Time(m.UpdatedAt),
	}
}

func (r *backupRepository) configToModel(d *domain.BackupConfig) *model.BackupConfig {
	if d == nil {
		return nil
	}
	isEnabled := int64(0)
	if d.IsEnabled {
		isEnabled = 1
	}
	return &model.BackupConfig{
		ID:             d.ID,
		UID:            d.UID,
		VaultID:        d.VaultID,
		Type:           d.Type,
		StorageIds:     d.StorageIds,
		IsEnabled:      isEnabled,
		CronStrategy:   d.CronStrategy,
		CronExpression: d.CronExpression,
		RetentionDays:  int64(d.RetentionDays),
		LastRunTime:    d.LastRunTime,
		NextRunTime:    d.NextRunTime,
		LastStatus:     int64(d.LastStatus),
		LastMessage:    d.LastMessage,
		CreatedAt:      timex.Time(d.CreatedAt),
		UpdatedAt:      timex.Time(d.UpdatedAt),
	}
}

func (r *backupRepository) historyToDomain(m *model.BackupHistory) *domain.BackupHistory {
	if m == nil {
		return nil
	}
	return &domain.BackupHistory{
		ID:        m.ID,
		UID:       m.UID,
		ConfigID:  m.ConfigID,
		StorageID: m.StorageID,
		Type:      m.Type,
		StartTime: m.StartTime,
		EndTime:   m.EndTime,
		Status:    int(m.Status),
		FileSize:  m.FileSize,
		FileCount: m.FileCount,
		Message:   m.Message,
		FilePath:  m.FilePath,
		CreatedAt: time.Time(m.CreatedAt),
		UpdatedAt: time.Time(m.UpdatedAt),
	}
}

func (r *backupRepository) historyToModel(d *domain.BackupHistory) *model.BackupHistory {
	if d == nil {
		return nil
	}
	return &model.BackupHistory{
		ID:        d.ID,
		UID:       d.UID,
		ConfigID:  d.ConfigID,
		StorageID: d.StorageID,
		Type:      d.Type,
		StartTime: d.StartTime,
		EndTime:   d.EndTime,
		Status:    int64(d.Status),
		FileSize:  d.FileSize,
		FileCount: d.FileCount,
		Message:   d.Message,
		FilePath:  d.FilePath,
		CreatedAt: timex.Time(d.CreatedAt),
		UpdatedAt: timex.Time(d.UpdatedAt),
	}
}

func (r *backupRepository) GetByID(ctx context.Context, id, uid int64) (*domain.BackupConfig, error) {
	q := r.backup(uid).BackupConfig
	m, err := q.WithContext(ctx).Where(q.UID.Eq(uid), q.ID.Eq(id)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.configToDomain(m), nil
}

func (r *backupRepository) DeleteConfig(ctx context.Context, id, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		q := r.backup(uid).BackupConfig
		// Limit to UID for safety
		_, err := q.WithContext(ctx).Where(q.UID.Eq(uid), q.ID.Eq(id)).Delete()
		return err
	})
}

func (r *backupRepository) ListConfigs(ctx context.Context, uid int64) ([]*domain.BackupConfig, error) {
	q := r.backup(uid).BackupConfig
	configs, err := q.WithContext(ctx).Where(q.UID.Eq(uid)).Order(q.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	var result []*domain.BackupConfig
	for _, m := range configs {
		result = append(result, r.configToDomain(m))
	}
	return result, nil
}

func (r *backupRepository) SaveConfig(ctx context.Context, config *domain.BackupConfig, uid int64) (*domain.BackupConfig, error) {
	var result *domain.BackupConfig
	err := r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		q := r.backup(uid).BackupConfig
		m := r.configToModel(config)
		m.UID = uid

		// 如果 ID > 0，执行更新逻辑
		if config.ID > 0 {
			// 检查 ID 是否属于当前用户
			old, err := q.WithContext(ctx).Where(q.UID.Eq(uid), q.ID.Eq(config.ID)).First()
			if err != nil {
				return err // RecordNotFound or other error
			}
			m.CreatedAt = old.CreatedAt
			m.UpdatedAt = timex.Now()
			// Update columns
			if err := q.WithContext(ctx).Save(m); err != nil {
				return err
			}
		} else {
			// ID == 0，执行创建新配置逻辑
			m.CreatedAt = timex.Now()
			m.UpdatedAt = timex.Now()
			if err := q.WithContext(ctx).Create(m); err != nil {
				return err
			}
		}
		result = r.configToDomain(m)
		return nil
	})
	return result, err
}

func (r *backupRepository) ListEnabledConfigs(ctx context.Context) ([]*domain.BackupConfig, error) {
	// 这是一个跨库操作，需要在外部循环所有用户。
	// 但在 Repository 层，我们只实现针对特定库的操作。
	// 这里其实有点矛盾，因为 ListEnabledConfigs 的语义通常是“全局”。
	// 按照 dao.go 的逻辑，我们可以先获取所有 UID，然后逐个查。
	uids, err := r.dao.GetAllUserUIDs()
	if err != nil {
		return nil, err
	}

	var allConfigs []*domain.BackupConfig
	for _, uid := range uids {
		q := r.backup(uid).BackupConfig
		configs, err := q.WithContext(ctx).Where(q.UID.Eq(uid), q.IsEnabled.Eq(1)).Find()
		if err != nil {
			continue
		}
		for _, m := range configs {
			allConfigs = append(allConfigs, r.configToDomain(m))
		}
	}
	return allConfigs, nil
}

func (r *backupRepository) UpdateNextRunTime(ctx context.Context, id, uid int64, nextRun time.Time) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		q := r.backup(uid).BackupConfig
		_, err := q.WithContext(ctx).Where(q.ID.Eq(id)).Update(q.NextRunTime, nextRun)
		return err
	})
}

// 修改接口定义以支持无 UID 调用 (如果 ID 包含在 Config 中)
func (r *backupRepository) CreateHistory(ctx context.Context, history *domain.BackupHistory, uid int64) (*domain.BackupHistory, error) {
	var result *domain.BackupHistory
	err := r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		q := r.backup(uid).BackupHistory
		m := r.historyToModel(history)
		m.UID = uid
		m.CreatedAt = timex.Now()
		m.UpdatedAt = timex.Now()
		if err := q.WithContext(ctx).Save(m); err != nil {
			return err
		}
		result = r.historyToDomain(m)
		return nil
	})
	return result, err
}

func (r *backupRepository) ListHistory(ctx context.Context, uid int64, configID int64, page, pageSize int) ([]*domain.BackupHistory, int64, error) {
	q := r.backup(uid).BackupHistory
	offset := (page - 1) * pageSize
	modelList, count, err := q.WithContext(ctx).Where(q.UID.Eq(uid), q.ConfigID.Eq(configID)).Order(q.ID.Desc()).FindByPage(offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	var list []*domain.BackupHistory
	for _, m := range modelList {
		list = append(list, r.historyToDomain(m))
	}
	return list, count, nil
}

func (r *backupRepository) ListOldHistory(ctx context.Context, uid int64, configID int64, cutoffTime time.Time) ([]*domain.BackupHistory, error) {
	q := r.backup(uid).BackupHistory
	modelList, err := q.WithContext(ctx).Where(q.ConfigID.Eq(configID), q.UID.Eq(uid), q.CreatedAt.Lt(timex.Time(cutoffTime))).Find()
	if err != nil {
		return nil, err
	}

	var list []*domain.BackupHistory
	for _, m := range modelList {
		list = append(list, r.historyToDomain(m))
	}
	return list, nil
}

func (r *backupRepository) DeleteOldHistory(ctx context.Context, uid int64, configID int64, cutoffTime time.Time) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		q := r.backup(uid).BackupHistory
		// Delete history records created before cutoffTime
		_, err := q.WithContext(ctx).Where(q.ConfigID.Eq(configID), q.UID.Eq(uid), q.CreatedAt.Lt(timex.Time(cutoffTime))).Delete()
		return err
	})
}

// 确保 backupRepository 实现了 domain.BackupRepository 接口
var _ domain.BackupRepository = (*backupRepository)(nil)
