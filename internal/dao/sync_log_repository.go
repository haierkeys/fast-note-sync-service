// Package dao implements the data access layer
// Package dao 实现数据访问层
package dao

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

// syncLogRepository implements domain.SyncLogRepository
// syncLogRepository 实现 domain.SyncLogRepository 接口
type syncLogRepository struct {
	dao             *Dao
	customPrefixKey string
	migrateOnce     sync.Map // tracks per-key migration completion // 记录每个 key 是否已完成 AutoMigrate
}

// NewSyncLogRepository creates a SyncLogRepository instance
// NewSyncLogRepository 创建 SyncLogRepository 实例
func NewSyncLogRepository(dao *Dao) domain.SyncLogRepository {
	return &syncLogRepository{dao: dao, customPrefixKey: "user_"}
}

// GetKey returns the database routing key for the given user
// GetKey 返回指定用户的数据库路由键（写入用户库）
func (r *syncLogRepository) GetKey(uid int64) string {
	return r.customPrefixKey + strconv.FormatInt(uid, 10)
}

func init() {
	RegisterModel(ModelConfig{
		Name: "SyncLog",
		RepoFactory: func(d *Dao) daoDBCustomKey {
			return NewSyncLogRepository(d).(daoDBCustomKey)
		},
		IsMainDB: false,
	})
}

// db returns the *gorm.DB for sync_log in the user's database, with one-time AutoMigrate
// db 返回用户库中 sync_log 对应的 *gorm.DB，确保每个用户库只迁移一次
func (r *syncLogRepository) db(uid int64) *gorm.DB {
	key := r.GetKey(uid)
	onceKey := key + "#syncLog"
	if _, loaded := r.migrateOnce.LoadOrStore(onceKey, true); !loaded {
		db := r.dao.ResolveDB(key)
		if db != nil {
			model.AutoMigrate(db, "SyncLog")
		}
	}
	return r.dao.ResolveDB(key)
}

// Create stores a new sync log entry
// Create 存储一条新的同步日志
func (r *syncLogRepository) Create(ctx context.Context, log *domain.SyncLog, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		m := &model.SyncLog{
			UID:           log.UID,
			VaultID:       log.VaultID,
			Type:          string(log.Type),
			Action:        string(log.Action),
			ChangedFields: log.ChangedFields,
			Path:          log.Path,
			PathHash:      log.PathHash,
			Size:          log.Size,
			ClientName:    log.ClientName,
			Status:        log.Status,
			Message:       log.Message,
			CreatedAt:     timex.Time(log.CreatedAt),
		}
		if m.CreatedAt.IsZero() {
			m.CreatedAt = timex.Time(time.Now())
		}
		return r.db(uid).WithContext(ctx).Create(m).Error
	})
}

// List retrieves sync logs for a user with optional filters and pagination
// List 按条件分页查询用户的同步日志
func (r *syncLogRepository) List(ctx context.Context, uid int64, logType, action string, page, pageSize int) ([]*domain.SyncLog, int64, error) {
	db := r.db(uid)

	query := db.WithContext(ctx).Model(&model.SyncLog{})
	if logType != "" {
		query = query.Where("type = ?", logType)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var rows []*model.SyncLog
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	results := make([]*domain.SyncLog, 0, len(rows))
	for _, m := range rows {
		results = append(results, &domain.SyncLog{
			ID:            m.ID,
			UID:           m.UID,
			VaultID:       m.VaultID,
			Type:          domain.SyncLogType(m.Type),
			Action:        domain.SyncLogAction(m.Action),
			ChangedFields: m.ChangedFields,
			Path:          m.Path,
			PathHash:      m.PathHash,
			Size:          m.Size,
			ClientName:    m.ClientName,
			Status:        m.Status,
			Message:       m.Message,
			CreatedAt:     time.Time(m.CreatedAt),
		})
	}
	return results, total, nil
}

// Ensure syncLogRepository implements domain.SyncLogRepository
// 确保 syncLogRepository 实现了 domain.SyncLogRepository 接口
var _ domain.SyncLogRepository = (*syncLogRepository)(nil)
