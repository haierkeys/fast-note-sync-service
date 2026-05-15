package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type stubAuthTokenRepository struct {
	getByIDToken *domain.AuthToken
	getByIDErr   error
}

func (r *stubAuthTokenRepository) Create(ctx context.Context, token *domain.AuthToken) (*domain.AuthToken, error) {
	return nil, errors.New("not implemented")
}

func (r *stubAuthTokenRepository) GetByID(ctx context.Context, id int64) (*domain.AuthToken, error) {
	return r.getByIDToken, r.getByIDErr
}

func (r *stubAuthTokenRepository) GetByTokenString(ctx context.Context, tokenString string) (*domain.AuthToken, error) {
	return nil, errors.New("not implemented")
}

func (r *stubAuthTokenRepository) ListByUID(ctx context.Context, uid int64) ([]*domain.AuthToken, error) {
	return nil, errors.New("not implemented")
}

func (r *stubAuthTokenRepository) Update(ctx context.Context, token *domain.AuthToken) error {
	return errors.New("not implemented")
}

func (r *stubAuthTokenRepository) UpdateScope(ctx context.Context, id int64, scope string) error {
	return errors.New("not implemented")
}

func (r *stubAuthTokenRepository) UpdateLastUsedAt(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (r *stubAuthTokenRepository) Revoke(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}

func (r *stubAuthTokenRepository) RevokeAllByUID(ctx context.Context, uid int64) error {
	return errors.New("not implemented")
}

func (r *stubAuthTokenRepository) UpdateTokenString(ctx context.Context, id int64, tokenString string) error {
	return errors.New("not implemented")
}

func tokenErrorCode(t *testing.T, err error) int {
	t.Helper()
	require.Error(t, err)
	appErr, ok := err.(*code.Code)
	require.True(t, ok, "expected *code.Code, got %T", err)
	return appErr.Code()
}

func TestTokenService_GetActiveToken_RecordNotFoundIsSessionError(t *testing.T) {
	svc := NewTokenService(
		&stubAuthTokenRepository{getByIDErr: gorm.ErrRecordNotFound},
		nil,
		app.NewTokenManager(app.TokenConfig{SecretKey: "test-secret"}),
		nil,
	)

	token, err := svc.GetActiveToken(context.Background(), 1, 2)

	assert.Nil(t, token)
	assert.Equal(t, code.ErrorInvalidUserAuthToken.Code(), tokenErrorCode(t, err))
}

func TestTokenService_GetActiveToken_RevokedIsSessionError(t *testing.T) {
	svc := NewTokenService(
		&stubAuthTokenRepository{getByIDToken: &domain.AuthToken{
			ID:        2,
			UID:       1,
			Status:    0,
			ExpiredAt: time.Now().Add(time.Hour),
		}},
		nil,
		app.NewTokenManager(app.TokenConfig{SecretKey: "test-secret"}),
		nil,
	)

	token, err := svc.GetActiveToken(context.Background(), 1, 2)

	assert.Nil(t, token)
	assert.Equal(t, code.ErrorInvalidUserAuthToken.Code(), tokenErrorCode(t, err))
}

func TestTokenService_GetActiveToken_WrongUserIsSessionError(t *testing.T) {
	svc := NewTokenService(
		&stubAuthTokenRepository{getByIDToken: &domain.AuthToken{
			ID:        2,
			UID:       99,
			Status:    1,
			ExpiredAt: time.Now().Add(time.Hour),
		}},
		nil,
		app.NewTokenManager(app.TokenConfig{SecretKey: "test-secret"}),
		nil,
	)

	token, err := svc.GetActiveToken(context.Background(), 1, 2)

	assert.Nil(t, token)
	assert.Equal(t, code.ErrorInvalidUserAuthToken.Code(), tokenErrorCode(t, err))
}

func TestTokenService_GetActiveToken_ExpiredRemainsExpired(t *testing.T) {
	svc := NewTokenService(
		&stubAuthTokenRepository{getByIDToken: &domain.AuthToken{
			ID:        2,
			UID:       1,
			Status:    1,
			ExpiredAt: time.Now().Add(-time.Hour),
		}},
		nil,
		app.NewTokenManager(app.TokenConfig{SecretKey: "test-secret"}),
		nil,
	)

	token, err := svc.GetActiveToken(context.Background(), 1, 2)

	assert.Nil(t, token)
	assert.Equal(t, code.ErrorTokenExpired.Code(), tokenErrorCode(t, err))
}

func TestTokenService_GetActiveToken_DBFailureRemainsDBError(t *testing.T) {
	svc := NewTokenService(
		&stubAuthTokenRepository{getByIDErr: errors.New("database offline")},
		nil,
		app.NewTokenManager(app.TokenConfig{SecretKey: "test-secret"}),
		nil,
	)

	token, err := svc.GetActiveToken(context.Background(), 1, 2)

	assert.Nil(t, token)
	assert.Equal(t, code.ErrorDBQuery.Code(), tokenErrorCode(t, err))
}

func TestTokenService_GetActiveToken_Valid(t *testing.T) {
	want := &domain.AuthToken{
		ID:        2,
		UID:       1,
		Status:    1,
		ExpiredAt: time.Now().Add(time.Hour),
	}
	svc := NewTokenService(
		&stubAuthTokenRepository{getByIDToken: want},
		nil,
		app.NewTokenManager(app.TokenConfig{SecretKey: "test-secret"}),
		nil,
	)

	got, err := svc.GetActiveToken(context.Background(), 1, 2)

	assert.NoError(t, err)
	assert.Same(t, want, got)
}
