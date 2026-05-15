package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMiddlewareTokenService struct {
	activeToken *domain.AuthToken
	activeErr   error
}

func (s *fakeMiddlewareTokenService) Create(ctx context.Context, uid int64, params *dto.TokenIssueRequest) (*dto.TokenCreateResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *fakeMiddlewareTokenService) CreateForLogin(ctx context.Context, uid int64, clientType, ip, userAgent string) (*domain.AuthToken, string, error) {
	return nil, "", errors.New("not implemented")
}

func (s *fakeMiddlewareTokenService) ListByUser(ctx context.Context, uid int64) ([]*dto.TokenResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *fakeMiddlewareTokenService) Update(ctx context.Context, uid int64, tokenID int64, params *dto.TokenUpdateRequest) error {
	return errors.New("not implemented")
}

func (s *fakeMiddlewareTokenService) Revoke(ctx context.Context, uid int64, tokenID int64) error {
	return errors.New("not implemented")
}

func (s *fakeMiddlewareTokenService) Rotate(ctx context.Context, uid int64, tokenID int64) (*dto.TokenCreateResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *fakeMiddlewareTokenService) GetActiveToken(ctx context.Context, uid int64, tokenID int64) (*domain.AuthToken, error) {
	return s.activeToken, s.activeErr
}

func (s *fakeMiddlewareTokenService) RecordAccessLog(ctx context.Context, log *domain.AuthTokenLog) error {
	return nil
}

func (s *fakeMiddlewareTokenService) ListLogs(ctx context.Context, uid, tokenID int64, page, pageSize int) ([]*dto.TokenLogResponse, int64, error) {
	return nil, 0, errors.New("not implemented")
}

func (s *fakeMiddlewareTokenService) UpdateLastUsedAt(ctx context.Context, tokenID int64) error {
	return errors.New("not implemented")
}

func (s *fakeMiddlewareTokenService) SetSyncHandler(handler func(uid int64, tokenID int64, scope string, kick bool)) {
}

func newMiddlewareJWT(t *testing.T, secretKey, nonce string) string {
	t.Helper()
	tokenManager := app.NewTokenManager(app.TokenConfig{
		SecretKey: secretKey,
		Expiry:    time.Hour,
	})
	token, err := tokenManager.Generate(1, "", "", 2, nonce)
	require.NoError(t, err)
	return token
}

func runUserAuthMiddleware(t *testing.T, tokenService *fakeMiddlewareTokenService, token string) app.Res {
	t.Helper()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(UserAuthTokenWithConfig("test-secret", tokenService))
	router.GET("/api/note/list", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": code.Success.Code(), "status": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/note/list?path=test.md", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("x-client", "ObsidianPlugin")
	req.Header.Set("User-Agent", "Obsidian")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	var res app.Res
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &res))
	return res
}

func TestUserAuthTokenWithConfig_AllowsValidManualTokenScope(t *testing.T) {
	token := newMiddlewareJWT(t, "test-secret", "nonce-ok")
	res := runUserAuthMiddleware(t, &fakeMiddlewareTokenService{activeToken: &domain.AuthToken{
		ID:          2,
		UID:         1,
		TokenString: "nonce-ok",
		Status:      1,
		Scope:       "p:rest,ws c:ObsidianPlugin f:note_rw,file_rw,config_rw",
		IssueType:   2,
		ExpiredAt:   time.Now().Add(time.Hour),
	}}, token)

	assert.Equal(t, code.Success.Code(), res.Code)
}

func TestUserAuthTokenWithConfig_PropagatesInvalidStatefulToken(t *testing.T) {
	token := newMiddlewareJWT(t, "test-secret", "stale-nonce")
	res := runUserAuthMiddleware(t, &fakeMiddlewareTokenService{
		activeErr: code.ErrorInvalidUserAuthToken.WithDetails("Token has been revoked or no longer exists"),
	}, token)

	assert.Equal(t, code.ErrorInvalidUserAuthToken.Code(), res.Code)
	assert.Contains(t, res.Details, "revoked")
}

func TestUserAuthTokenWithConfig_PropagatesExpiredToken(t *testing.T) {
	token := newMiddlewareJWT(t, "test-secret", "expired-nonce")
	res := runUserAuthMiddleware(t, &fakeMiddlewareTokenService{
		activeErr: code.ErrorTokenExpired,
	}, token)

	assert.Equal(t, code.ErrorTokenExpired.Code(), res.Code)
}

func TestUserAuthTokenWithConfig_RejectsNonceMismatch(t *testing.T) {
	token := newMiddlewareJWT(t, "test-secret", "old-nonce")
	res := runUserAuthMiddleware(t, &fakeMiddlewareTokenService{activeToken: &domain.AuthToken{
		ID:          2,
		UID:         1,
		TokenString: "new-nonce",
		Status:      1,
		Scope:       "p:rest,ws c:ObsidianPlugin f:note_rw,file_rw,config_rw",
		IssueType:   2,
		ExpiredAt:   time.Now().Add(time.Hour),
	}}, token)

	assert.Equal(t, code.ErrorInvalidUserAuthToken.Code(), res.Code)
	assert.Contains(t, res.Details, "rotated")
}

func TestUserAuthTokenWithConfig_RejectsScopeRestrictedToken(t *testing.T) {
	token := newMiddlewareJWT(t, "test-secret", "nonce-ok")
	res := runUserAuthMiddleware(t, &fakeMiddlewareTokenService{activeToken: &domain.AuthToken{
		ID:          2,
		UID:         1,
		TokenString: "nonce-ok",
		Status:      1,
		Scope:       "p:ws c:ObsidianPlugin f:note_rw",
		IssueType:   2,
		ExpiredAt:   time.Now().Add(time.Hour),
	}}, token)

	assert.Equal(t, code.ErrorAuthTokenScopeRestricted.Code(), res.Code)
	assert.Contains(t, res.Details, "Permission denied")
}
