package dao

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

// userShareRepository implements domain.UserShareRepository interface
// userShareRepository 实现 domain.UserShareRepository 接口
type userShareRepository struct {
	dao             *Dao
	customPrefixKey string
}

// NewUserShareRepository creates UserShareRepository instance
// NewUserShareRepository 创建 UserShareRepository 实例
func NewUserShareRepository(dao *Dao) domain.UserShareRepository {
	return &userShareRepository{dao: dao, customPrefixKey: "user_share_"}
}

func (r *userShareRepository) GetKey(uid int64) string {
	return r.customPrefixKey + strconv.FormatInt(uid, 10)
}

func init() {
	RegisterModel(ModelConfig{
		Name: "UserShare",
		RepoFactory: func(d *Dao) daoDBCustomKey {
			return NewUserShareRepository(d).(daoDBCustomKey)
		},
	})
}

// userShare gets the share query object
// userShare 获取分享查询对象
func (r *userShareRepository) userShare(uid int64) *query.Query {
	key := r.GetKey(uid)
	return r.dao.QueryWithOnceInit(func(g *gorm.DB) {
		model.AutoMigrate(g, "UserShare")
	}, key+"#userShare", key)
}

// toDomain converts database model to domain model
// toDomain 将数据库模型转换为领域模型
func (r *userShareRepository) toDomain(m *model.UserShare) *domain.UserShare {
	if m == nil {
		return nil
	}
	var res map[string][]string
	_ = json.Unmarshal([]byte(m.Res), &res)

	return &domain.UserShare{
		ID:           m.ID,
		UID:          m.UID,
		ResType:      m.ResType,
		ResID:        m.ResID,
		Resources:    res,
		Status:       m.Status,
		ViewCount:    m.ViewCount,
		LastViewedAt: m.LastViewedAt,
		ExpiresAt:    m.ExpiresAt,
		Password:     m.Password,
		ShortLink:    m.ShortLink,
		CreatedAt:    time.Time(m.CreatedAt),
		UpdatedAt:    time.Time(m.UpdatedAt),
	}
}

// toModel converts domain model to database model
// toModel 将领域模型转换为数据库模型
func (r *userShareRepository) toModel(d *domain.UserShare) *model.UserShare {
	if d == nil {
		return nil
	}
	resBytes, _ := json.Marshal(d.Resources)

	return &model.UserShare{
		ID:           d.ID,
		UID:          d.UID,
		ResType:      d.ResType,
		ResID:        d.ResID,
		Res:          string(resBytes),
		Status:       d.Status,
		ViewCount:    d.ViewCount,
		LastViewedAt: d.LastViewedAt,
		ExpiresAt:    d.ExpiresAt,
		Password:     d.Password,
		ShortLink:    d.ShortLink,
		CreatedAt:    timex.Time(d.CreatedAt),
		UpdatedAt:    timex.Time(d.UpdatedAt),
	}
}

func (r *userShareRepository) Create(ctx context.Context, uid int64, share *domain.UserShare) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		us := r.userShare(uid).UserShare
		m := r.toModel(share)
		if err := us.WithContext(ctx).Create(m); err != nil {
			return err
		}
		share.ID = m.ID // Backfill generated ID // 回填生成的 ID
		return nil
	})
}

func (r *userShareRepository) GetByID(ctx context.Context, uid int64, id int64) (*domain.UserShare, error) {
	us := r.userShare(uid).UserShare
	m, err := us.WithContext(ctx).Where(us.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

func (r *userShareRepository) GetByPath(ctx context.Context, uid int64, vaultID int64, pathHash string) (*domain.UserShare, error) {
	// Get Note (Triggers Notes migration via noteRepo)
	noteRepo := NewNoteRepository(r.dao)
	note, err := noteRepo.GetByPathHash(ctx, pathHash, vaultID, uid)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return nil, nil
	}

	// 3. Get UserShare (Triggers UserShare migration via GetByRes -> userShare(uid))
	// 3. Get UserShare (通过 GetByRes -> userShare(uid) 触发 UserShare 迁移)
	return r.GetByRes(ctx, uid, "note", note.ID)
}

func (r *userShareRepository) GetByRes(ctx context.Context, uid int64, resType string, resID int64) (*domain.UserShare, error) {
	us := r.userShare(uid).UserShare
	m, err := us.WithContext(ctx).Where(us.ResType.Eq(resType), us.ResID.Eq(resID), us.Status.Eq(domain.UserShareStatusActive)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

func (r *userShareRepository) UpdateResources(ctx context.Context, uid int64, id int64, resources map[string][]string) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		us := r.userShare(uid).UserShare
		resBytes, err := json.Marshal(resources)
		if err != nil {
			return err
		}
		_, err = us.WithContext(ctx).Where(us.ID.Eq(id)).Update(us.Res, string(resBytes))
		return err
	})
}

func (r *userShareRepository) UpdateStatus(ctx context.Context, uid int64, id int64, status int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		us := r.userShare(uid).UserShare
		_, err := us.WithContext(ctx).Where(us.ID.Eq(id)).Update(us.Status, status)
		return err
	})
}

func (r *userShareRepository) UpdateStatusByRes(ctx context.Context, uid int64, resType string, resID int64, status int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		us := r.userShare(uid).UserShare
		_, err := us.WithContext(ctx).Where(us.UID.Eq(uid), us.ResType.Eq(resType), us.ResID.Eq(resID), us.Status.Eq(domain.UserShareStatusActive)).Update(us.Status, status)
		return err
	})
}

func (r *userShareRepository) UpdateViewStats(ctx context.Context, uid int64, id int64, viewCountIncr int64, lastViewedAt time.Time) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		us := r.userShare(uid).UserShare
		_, err := us.WithContext(ctx).Where(us.ID.Eq(id)).Updates(map[string]interface{}{
			"view_count":     gorm.Expr("view_count + ?", viewCountIncr),
			"last_viewed_at": lastViewedAt,
		})
		return err
	})
}

func (r *userShareRepository) ListByUID(ctx context.Context, uid int64, sortBy string, sortOrder string, offset, limit int) ([]*domain.UserShare, error) {
	us := r.userShare(uid).UserShare

	// Whitelist sorting field validation
	// 白名单验证排序字段
	allowedFields := map[string]string{
		"created_at": "created_at",
		"updated_at": "updated_at",
		"expires_at": "expires_at",
	}
	field, ok := allowedFields[sortBy]
	if !ok {
		field = "created_at"
	}

	// Validate sorting order
	// 验证排序方向
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	orderClause := field + " " + sortOrder
	var ms []*model.UserShare
	q := us.WithContext(ctx).Where(us.UID.Eq(uid), us.Status.Eq(domain.UserShareStatusActive))
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	err := q.UnderlyingDB().Order(orderClause).Find(&ms).Error
	if err != nil {
		return nil, err
	}
	var ds []*domain.UserShare
	for _, m := range ms {
		ds = append(ds, r.toDomain(m))
	}
	return ds, nil
}

func (r *userShareRepository) UpdateShortLink(ctx context.Context, uid int64, id int64, shortLink string) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		us := r.userShare(uid).UserShare
		_, err := us.WithContext(ctx).Where(us.ID.Eq(id)).Update(us.ShortLink, shortLink)
		return err
	})
}

func (r *userShareRepository) UpdatePassword(ctx context.Context, uid int64, id int64, password string) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		us := r.userShare(uid).UserShare
		_, err := us.WithContext(ctx).Where(us.ID.Eq(id)).Update(us.Password, password)
		return err
	})
}

func (r *userShareRepository) CountByUID(ctx context.Context, uid int64) (int64, error) {
	us := r.userShare(uid).UserShare
	return us.WithContext(ctx).Where(us.UID.Eq(uid), us.Status.Eq(domain.UserShareStatusActive)).Count()
}

// ListActiveNoteResIDs returns note res_ids for all active shares of a user
// ListActiveNoteResIDs retrieves note res_id list for all active shares of a user (queries only user_shares table, no cross-database JOIN)
// ListActiveNoteResIDs 查询该用户所有有效分享中 res_type='note' 的 res_id 列表（只查 user_shares 表，无跨库 JOIN）
func (r *userShareRepository) ListActiveNoteResIDs(ctx context.Context, uid int64) ([]int64, error) {
	us := r.userShare(uid).UserShare
	ms, err := us.WithContext(ctx).
		Where(us.UID.Eq(uid), us.ResType.Eq("note"), us.Status.Eq(domain.UserShareStatusActive)).
		Find()
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(ms))
	for _, m := range ms {
		ids = append(ids, m.ResID)
	}
	return ids, nil
}

// ListChangedNoteResIDs returns note share res_ids changed after since, grouped by status
// ListChangedNoteResIDs 返回 updated_at > since 的 note 分享记录，按状态分组
// ListChangedNoteResIDs returns note share res_ids changed after since, grouped by status
func (r *userShareRepository) ListChangedNoteResIDs(ctx context.Context, uid int64, since time.Time) ([]int64, []int64, error) {
	us := r.userShare(uid).UserShare
	ms, err := us.WithContext(ctx).
		Where(us.UID.Eq(uid), us.ResType.Eq("note"), us.UpdatedAt.Gt(timex.Time(since))).
		Find()
	if err != nil {
		return nil, nil, err
	}
	var active, revoked []int64
	for _, m := range ms {
		switch m.Status {
		case domain.UserShareStatusActive:
			active = append(active, m.ResID)
		case domain.UserShareStatusRevoked:
			revoked = append(revoked, m.ResID)
		}
	}
	return active, revoked, nil
}

// MigrateResID updates res_id and resources JSON when a note/file is renamed (old ID → new ID).
// MigrateResID updates res_id and resources JSON when a note/file is renamed (old ID -> new ID).
// MigrateResID 在笔记/文件重命名时更新分享记录的资源 ID 和资源列表（旧 ID -> 新 ID）。
func (r *userShareRepository) MigrateResID(ctx context.Context, uid int64, oldResID int64, newResID int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		us := r.userShare(uid).UserShare

		// 1. Update res_id for all active shares pointing to oldResID
		// 1. Update res_id for all active shares pointing to oldResID
		// 1. 更新所有指向旧 ID 的有效分享的 res_id
		_, err := us.WithContext(ctx).
			Where(us.UID.Eq(uid), us.ResID.Eq(oldResID), us.Status.Eq(domain.UserShareStatusActive)).
			Update(us.ResID, newResID)
		if err != nil {
			return err
		}

		// 2. Update resources JSON: replace oldResID with newResID in all note/file arrays
		// 2. Update resources JSON: replace oldResID with newResID in all note/file arrays
		// 2. 更新资源 JSON：在 note/file 数组中将旧 ID 替换为新 ID
		oldIDStr := strconv.FormatInt(oldResID, 10)
		newIDStr := strconv.FormatInt(newResID, 10)

		shares, err := us.WithContext(ctx).
			Where(us.UID.Eq(uid), us.Status.Eq(domain.UserShareStatusActive)).
			Find()
		if err != nil {
			return err
		}

		for _, share := range shares {
			var res map[string][]string
			if err := json.Unmarshal([]byte(share.Res), &res); err != nil {
				continue
			}

			changed := false
			for _, ids := range res {
				for i, id := range ids {
					if id == oldIDStr {
						ids[i] = newIDStr
						changed = true
					}
				}
			}

			if changed {
				resBytes, _ := json.Marshal(res)
				_, err := us.WithContext(ctx).
					Where(us.ID.Eq(share.ID)).
					Update(us.Res, string(resBytes))
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// Ensure userShareRepository implements domain.UserShareRepository interface
// 确保 userShareRepository 实现了 domain.UserShareRepository 接口
var _ domain.UserShareRepository = (*userShareRepository)(nil)
