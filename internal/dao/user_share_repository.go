package dao

import (
	"context"
	"encoding/json"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

// userShareRepository 实现 domain.UserShareRepository 接口
type userShareRepository struct {
	dao *Dao
}

// NewUserShareRepository 创建 UserShareRepository 实例
func NewUserShareRepository(dao *Dao) domain.UserShareRepository {
	return &userShareRepository{dao: dao}
}

// userShare 获取分享查询对象
func (r *userShareRepository) userShare() *query.Query {
	return r.dao.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "UserShare")
	}, "user_share#user_share")
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
		Resources:    res,
		Status:       m.Status,
		ViewCount:    m.ViewCount,
		LastViewedAt: m.LastViewedAt,
		ExpiresAt:    m.ExpiresAt,
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
		Res:          string(resBytes),
		Status:       d.Status,
		ViewCount:    d.ViewCount,
		LastViewedAt: d.LastViewedAt,
		ExpiresAt:    d.ExpiresAt,
		CreatedAt:    timex.Time(d.CreatedAt),
		UpdatedAt:    timex.Time(d.UpdatedAt),
	}
}

func (r *userShareRepository) Create(ctx context.Context, share *domain.UserShare) error {
	us := r.userShare().UserShare
	m := r.toModel(share)
	if err := us.WithContext(ctx).Create(m); err != nil {
		return err
	}
	share.ID = m.ID // 回填生成的 ID
	return nil
}

func (r *userShareRepository) GetByID(ctx context.Context, id int64) (*domain.UserShare, error) {
	us := r.userShare().UserShare
	m, err := us.WithContext(ctx).Where(us.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

func (r *userShareRepository) UpdateStatus(ctx context.Context, id int64, status int64) error {
	us := r.userShare().UserShare
	_, err := us.WithContext(ctx).Where(us.ID.Eq(id)).Update(us.Status, status)
	return err
}

func (r *userShareRepository) UpdateViewStats(ctx context.Context, id int64, viewCountIncr int64, lastViewedAt time.Time) error {
	us := r.userShare().UserShare
	_, err := us.WithContext(ctx).Where(us.ID.Eq(id)).Updates(map[string]interface{}{
		"view_count":     gorm.Expr("view_count + ?", viewCountIncr),
		"last_viewed_at": lastViewedAt,
	})
	return err
}

func (r *userShareRepository) ListByUID(ctx context.Context, uid int64) ([]*domain.UserShare, error) {
	us := r.userShare().UserShare
	ms, err := us.WithContext(ctx).Where(us.UID.Eq(uid)).Find()
	if err != nil {
		return nil, err
	}
	var ds []*domain.UserShare
	for _, m := range ms {
		ds = append(ds, r.toDomain(m))
	}
	return ds, nil
}

var _ domain.UserShareRepository = (*userShareRepository)(nil)
