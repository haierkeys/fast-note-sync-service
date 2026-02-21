package dao

import (
	"context"
	"strconv"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

type gitSyncRepository struct {
	dao             *Dao
	customPrefixKey string
}

// NewGitSyncRepository 创建 GitSyncRepository 实例
func NewGitSyncRepository(dao *Dao) domain.GitSyncRepository {
	return &gitSyncRepository{dao: dao, customPrefixKey: "user_"}
}

func (r *gitSyncRepository) GetKey(uid int64) string {
	return r.customPrefixKey + strconv.FormatInt(uid, 10)
}

func (r *gitSyncRepository) gitSync(uid int64) *query.Query {
	return r.dao.UseQueryWithOnceFunc(func(g *gorm.DB) {
		_ = model.AutoMigrate(g, "GitSyncConfig")
	}, r.GetKey(uid)+"#git_sync", r.GetKey(uid))
}

func (r *gitSyncRepository) toDomain(m *model.GitSyncConfig) *domain.GitSyncConfig {
	if m == nil {
		return nil
	}
	var lastSyncTime *time.Time
	if !m.LastSyncTime.IsZero() {
		t := m.LastSyncTime
		lastSyncTime = &t
	}
	return &domain.GitSyncConfig{
		ID:           m.ID,
		UID:          m.UID,
		VaultID:      m.VaultID,
		RepoURL:      m.RepoURL,
		Username:     m.Username,
		Password:     m.Password,
		Branch:       m.Branch,
		IsEnabled:    m.IsEnabled == 1,
		Delay:        m.Delay,
		LastSyncTime: lastSyncTime,
		LastStatus:   m.LastStatus,
		LastMessage:  m.LastMessage,
		CreatedAt:    time.Time(m.CreatedAt),
		UpdatedAt:    time.Time(m.UpdatedAt),
	}
}

func (r *gitSyncRepository) toModel(d *domain.GitSyncConfig) *model.GitSyncConfig {
	if d == nil {
		return nil
	}
	isEnabled := int64(0)
	if d.IsEnabled {
		isEnabled = 1
	}
	var lastSyncTime time.Time
	if d.LastSyncTime != nil {
		lastSyncTime = *d.LastSyncTime
	}
	return &model.GitSyncConfig{
		ID:           d.ID,
		UID:          d.UID,
		VaultID:      d.VaultID,
		RepoURL:      d.RepoURL,
		Username:     d.Username,
		Password:     d.Password,
		Branch:       d.Branch,
		IsEnabled:    isEnabled,
		Delay:        d.Delay,
		LastSyncTime: lastSyncTime,
		LastStatus:   d.LastStatus,
		LastMessage:  d.LastMessage,
		CreatedAt:    timex.Time(d.CreatedAt),
		UpdatedAt:    timex.Time(d.UpdatedAt),
	}
}

func (r *gitSyncRepository) GetByID(ctx context.Context, id, uid int64) (*domain.GitSyncConfig, error) {
	q := r.gitSync(uid).GitSyncConfig
	m, err := q.WithContext(ctx).Where(q.ID.Eq(id), q.UID.Eq(uid)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.toDomain(m), nil
}

func (r *gitSyncRepository) GetByVaultID(ctx context.Context, vaultID, uid int64) (*domain.GitSyncConfig, error) {
	q := r.gitSync(uid).GitSyncConfig
	m, err := q.WithContext(ctx).Where(q.VaultID.Eq(vaultID), q.UID.Eq(uid)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return r.toDomain(m), nil
}

func (r *gitSyncRepository) Save(ctx context.Context, config *domain.GitSyncConfig, uid int64) (*domain.GitSyncConfig, error) {
	var result *domain.GitSyncConfig
	err := r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		q := r.gitSync(uid).GitSyncConfig
		m := r.toModel(config)
		m.UID = uid

		if config.ID > 0 {
			old, err := q.WithContext(ctx).Where(q.ID.Eq(config.ID), q.UID.Eq(uid)).First()
			if err != nil {
				return err
			}
			m.CreatedAt = old.CreatedAt
			m.UpdatedAt = timex.Now()
			if err := q.WithContext(ctx).Save(m); err != nil {
				return err
			}
		} else {
			m.CreatedAt = timex.Now()
			m.UpdatedAt = timex.Now()
			if err := q.WithContext(ctx).Create(m); err != nil {
				return err
			}
		}
		result = r.toDomain(m)
		return nil
	})
	return result, err
}

func (r *gitSyncRepository) Delete(ctx context.Context, id, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		q := r.gitSync(uid).GitSyncConfig
		_, err := q.WithContext(ctx).Where(q.ID.Eq(id), q.UID.Eq(uid)).Delete()
		return err
	})
}

func (r *gitSyncRepository) List(ctx context.Context, uid int64) ([]*domain.GitSyncConfig, error) {
	q := r.gitSync(uid).GitSyncConfig
	ms, err := q.WithContext(ctx).Where(q.UID.Eq(uid)).Order(q.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	var res []*domain.GitSyncConfig
	for _, m := range ms {
		res = append(res, r.toDomain(m))
	}
	return res, nil
}

func (r *gitSyncRepository) ListEnabled(ctx context.Context) ([]*domain.GitSyncConfig, error) {
	uids, err := r.dao.GetAllUserUIDs()
	if err != nil {
		return nil, err
	}
	var all []*domain.GitSyncConfig
	for _, uid := range uids {
		q := r.gitSync(uid).GitSyncConfig
		ms, err := q.WithContext(ctx).Where(q.UID.Eq(uid), q.IsEnabled.Eq(1)).Find()
		if err != nil {
			continue
		}
		for _, m := range ms {
			all = append(all, r.toDomain(m))
		}
	}
	return all, nil
}

// 确保 gitSyncRepository 实现了 domain.GitSyncRepository 接口
var _ domain.GitSyncRepository = (*gitSyncRepository)(nil)
