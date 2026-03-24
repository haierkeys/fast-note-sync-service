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

// userShareRepository 实现 domain.UserShareRepository 接口
type userShareRepository struct {
	dao             *Dao
	customPrefixKey string
}

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

// userShare 获取分享查询对象
func (r *userShareRepository) userShare(uid int64) *query.Query {
	key := r.GetKey(uid)
	return r.dao.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "UserShare")
	}, key+"#userShare", key)
}

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
		share.ID = m.ID // 回填生成的 ID
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
	return r.GetByRes(ctx, uid, "note", note.ID)
}

func (r *userShareRepository) GetByRes(ctx context.Context, uid int64, resType string, resID int64) (*domain.UserShare, error) {
	us := r.userShare(uid).UserShare
	m, err := us.WithContext(ctx).Where(us.ResType.Eq(resType), us.ResID.Eq(resID), us.Status.Eq(1)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

func (r *userShareRepository) UpdateStatus(ctx context.Context, uid int64, id int64, status int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		us := r.userShare(uid).UserShare
		_, err := us.WithContext(ctx).Where(us.ID.Eq(id)).Update(us.Status, status)
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

	// 验证排序方向
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	orderClause := field + " " + sortOrder
	var ms []*model.UserShare
	q := us.WithContext(ctx).Where(us.UID.Eq(uid), us.Status.Eq(1))
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
	return us.WithContext(ctx).Where(us.UID.Eq(uid), us.Status.Eq(1)).Count()
}

var _ domain.UserShareRepository = (*userShareRepository)(nil)
