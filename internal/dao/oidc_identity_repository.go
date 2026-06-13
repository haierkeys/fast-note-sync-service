package dao

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

type oidcIdentityRepository struct {
	dao *Dao
}

func NewOIDCIdentityRepository(dao *Dao) domain.OIDCIdentityRepository {
	return &oidcIdentityRepository{dao: dao}
}

func init() {
	RegisterModel(ModelConfig{
		Name:     "UserOIDCIdentity",
		IsMainDB: true,
	})
}

func (r *oidcIdentityRepository) db() *gorm.DB {
	db := r.dao.ResolveDB()
	r.dao.QueryWithOnceInit(func(g *gorm.DB) {
		model.AutoMigrate(g, "UserOIDCIdentity")
	}, "oidc_identity#user_oidc_identity")
	return db
}

func (r *oidcIdentityRepository) toDomain(m *model.UserOIDCIdentity) *domain.OIDCIdentity {
	if m == nil {
		return nil
	}
	return &domain.OIDCIdentity{
		ID:        m.ID,
		UID:       m.UID,
		Issuer:    m.Issuer,
		Subject:   m.Subject,
		Email:     m.Email,
		Username:  m.Username,
		CreatedAt: time.Time(m.CreatedAt),
		UpdatedAt: time.Time(m.UpdatedAt),
	}
}

func (r *oidcIdentityRepository) toModel(identity *domain.OIDCIdentity) *model.UserOIDCIdentity {
	if identity == nil {
		return nil
	}
	return &model.UserOIDCIdentity{
		ID:        identity.ID,
		UID:       identity.UID,
		Issuer:    identity.Issuer,
		Subject:   identity.Subject,
		Email:     identity.Email,
		Username:  identity.Username,
		CreatedAt: timex.Time(identity.CreatedAt),
		UpdatedAt: timex.Time(identity.UpdatedAt),
	}
}

func (r *oidcIdentityRepository) GetByIssuerSubject(ctx context.Context, issuer, subject string) (*domain.OIDCIdentity, error) {
	var m model.UserOIDCIdentity
	if err := r.db().WithContext(ctx).Where("issuer = ? AND subject = ?", issuer, subject).First(&m).Error; err != nil {
		return nil, err
	}
	return r.toDomain(&m), nil
}

func (r *oidcIdentityRepository) Create(ctx context.Context, identity *domain.OIDCIdentity) (*domain.OIDCIdentity, error) {
	m := r.toModel(identity)
	m.CreatedAt = timex.Now()
	m.UpdatedAt = timex.Now()
	if err := r.db().WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

var _ domain.OIDCIdentityRepository = (*oidcIdentityRepository)(nil)
