package dao

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
)

func init() {
	RegisterModel(ModelConfig{
		Name:     "AuthToken",
		IsMainDB: true,
	})
	RegisterModel(ModelConfig{
		Name:     "AuthTokenLog",
		IsMainDB: true,
	})
}

// authTokenRepository implements domain.AuthTokenRepository interface
// authTokenRepository 实现 domain.AuthTokenRepository 接口
type authTokenRepository struct {
	dao *Dao
}

// NewAuthTokenRepository creates AuthTokenRepository instance
// NewAuthTokenRepository 创建 AuthTokenRepository 实例
func NewAuthTokenRepository(dao *Dao) domain.AuthTokenRepository {
	return &authTokenRepository{dao: dao}
}

func (r *authTokenRepository) GetKey(uid int64) string {
	return ""
}

// authToken gets the auth token query object
// authToken 获取认证令牌查询对象
func (r *authTokenRepository) authToken() *query.Query {
	return r.dao.QueryWithOnceInit(func(g *gorm.DB) {
		model.AutoMigrate(g, "AuthToken")
	}, "user#auth_token")
}

// toDomain converts database model to domain model
// toDomain 将数据库模型转换为领域模型
func (r *authTokenRepository) toDomain(m *model.AuthToken) *domain.AuthToken {
	if m == nil {
		return nil
	}
	return &domain.AuthToken{
		ID:          int64(m.ID),
		UID:         int64(m.UID),
		TokenString: m.TokenString,
		Scope:       m.Scope,
		ClientType:  m.ClientType,
		BoundIP:     m.BoundIP,
		UserAgent:   m.UserAgent,
		Status:      int64(m.Status),
		ExpiredAt:   m.ExpiredAt,
		IssueType:   int(m.IssueType),
		LastUsedAt:  m.LastUsedAt,
		CreatedAt:   time.Time(m.CreatedAt),
		UpdatedAt:   time.Time(m.UpdatedAt),
	}
}

func (r *authTokenRepository) Create(ctx context.Context, token *domain.AuthToken) (*domain.AuthToken, error) {
	u := r.authToken().AuthToken
	m := &model.AuthToken{
		UID:         token.UID,
		TokenString: token.TokenString,
		Scope:       token.Scope,
		ClientType:  token.ClientType,
		BoundIP:     token.BoundIP,
		UserAgent:   token.UserAgent,
		Status:      token.Status,
		ExpiredAt:   token.ExpiredAt,
		IssueType:   int64(token.IssueType),
		LastUsedAt:  token.LastUsedAt,
		CreatedAt:   timex.Now(),
		UpdatedAt:   timex.Now(),
	}

	err := u.WithContext(ctx).Create(m)
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

func (r *authTokenRepository) GetByID(ctx context.Context, id int64) (*domain.AuthToken, error) {
	u := r.authToken().AuthToken
	m, err := u.WithContext(ctx).Where(u.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

func (r *authTokenRepository) GetByTokenString(ctx context.Context, tokenString string) (*domain.AuthToken, error) {
	u := r.authToken().AuthToken
	m, err := u.WithContext(ctx).Where(u.TokenString.Eq(tokenString)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

func (r *authTokenRepository) ListByUID(ctx context.Context, uid int64) ([]*domain.AuthToken, error) {
	u := r.authToken().AuthToken
	models, err := u.WithContext(ctx).Where(u.UID.Eq(uid), u.Status.Eq(1)).Find()
	if err != nil {
		return nil, err
	}

	var res []*domain.AuthToken
	for _, m := range models {
		res = append(res, r.toDomain(m))
	}
	return res, nil
}

func (r *authTokenRepository) UpdateScope(ctx context.Context, id int64, scope string) error {
	u := r.authToken().AuthToken
	_, err := u.WithContext(ctx).Where(u.ID.Eq(id)).UpdateSimple(
		u.Scope.Value(scope),
		u.UpdatedAt.Value(timex.Now()),
	)
	return err
}

func (r *authTokenRepository) Revoke(ctx context.Context, id int64) error {
	u := r.authToken().AuthToken
	_, err := u.WithContext(ctx).Where(u.ID.Eq(id)).UpdateSimple(
		u.Status.Value(0),
		u.UpdatedAt.Value(timex.Now()),
	)
	return err
}

func (r *authTokenRepository) RevokeAllByUID(ctx context.Context, uid int64) error {
	u := r.authToken().AuthToken
	_, err := u.WithContext(ctx).Where(u.UID.Eq(uid)).UpdateSimple(
		u.Status.Value(0),
		u.UpdatedAt.Value(timex.Now()),
	)
	return err
}

func (r *authTokenRepository) UpdateLastUsedAt(ctx context.Context, id int64) error {
	u := r.authToken().AuthToken
	_, err := u.WithContext(ctx).Where(u.ID.Eq(id)).UpdateSimple(
		u.LastUsedAt.Value(time.Now()),
	)
	return err
}

type authTokenLogRepository struct {
	dao *Dao
}

func (r *authTokenLogRepository) GetKey(uid int64) string {
	return ""
}

func NewAuthTokenLogRepository(dao *Dao) domain.AuthTokenLogRepository {
	return &authTokenLogRepository{dao: dao}
}

// authTokenLog gets the auth token log query object
// authTokenLog 获取认证令牌日志查询对象
func (r *authTokenLogRepository) authTokenLog() *query.Query {
	return r.dao.QueryWithOnceInit(func(g *gorm.DB) {
		model.AutoMigrate(g, "AuthTokenLog")
	}, "user#auth_token_log")
}

func (r *authTokenLogRepository) Create(ctx context.Context, log *domain.AuthTokenLog) error {
	u := r.authTokenLog().AuthTokenLog
	m := &model.AuthTokenLog{
		TokenID:       log.TokenID,
		UID:           log.UID,
		Protocol:      log.Protocol,
		Client:        log.Client,
		ClientName:    log.ClientName,
		ClientVersion: log.ClientVersion,
		IP:            log.IP,
		Ua:            log.UA,
		StatusCode:    log.StatusCode,
		CreatedAt:     timex.Now(),
	}
	return u.WithContext(ctx).Omit(u.ID).Create(m)
}

func (r *authTokenLogRepository) ListByTokenID(ctx context.Context, tokenID int64, page, pageSize int) ([]*domain.AuthTokenLog, int64, error) {
	u := r.authTokenLog().AuthTokenLog
	q := u.WithContext(ctx).Where(u.TokenID.Eq(tokenID))

	count, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	models, err := q.Order(u.CreatedAt.Desc()).Limit(pageSize).Offset(offset).Find()
	if err != nil {
		return nil, 0, err
	}

	var res []*domain.AuthTokenLog
	for _, m := range models {
		res = append(res, r.toDomain(m))
	}
	return res, count, nil
}

func (r *authTokenLogRepository) toDomain(m *model.AuthTokenLog) *domain.AuthTokenLog {
	return &domain.AuthTokenLog{
		ID:            m.ID,
		TokenID:       m.TokenID,
		UID:           m.UID,
		Protocol:      m.Protocol,
		Client:        m.Client,
		ClientName:    m.ClientName,
		ClientVersion: m.ClientVersion,
		IP:            m.IP,
		UA:            m.Ua,
		StatusCode:    m.StatusCode,
		CreatedAt:     time.Time(m.CreatedAt),
	}
}
