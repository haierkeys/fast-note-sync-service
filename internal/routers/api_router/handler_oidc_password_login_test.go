package api_router

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	domainmocks "github.com/haierkeys/fast-note-sync-service/internal/domain/mocks"
	internaloidc "github.com/haierkeys/fast-note-sync-service/internal/oidc"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type passwordDisabledOIDCProvider struct{ fakeOIDCProvider }

func (p *passwordDisabledOIDCProvider) Exchange(context.Context, string, string, string) (*internaloidc.Claims, error) {
	return &internaloidc.Claims{Subject: "new-subject", Email: "new@example.invalid", Username: "new-user"}, nil
}

type passwordDisabledIdentityRepo struct {
	domain.OIDCIdentityRepository
	created *domain.OIDCIdentity
}

func (r *passwordDisabledIdentityRepo) GetByIssuerSubject(context.Context, string, string) (*domain.OIDCIdentity, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *passwordDisabledIdentityRepo) Create(_ context.Context, identity *domain.OIDCIdentity) (*domain.OIDCIdentity, error) {
	r.created = identity
	return identity, nil
}

type passwordDisabledLoginIssuer struct{ service.TokenService }

func (s *passwordDisabledLoginIssuer) CreateForLogin(context.Context, int64, string, string, string) (*domain.AuthToken, string, error) {
	return &domain.AuthToken{ID: 12}, "oidc-login-token", nil
}

func TestOIDCCallbackCreatesUserWithPasswordLoginAndLocalRegistrationDisabled(t *testing.T) {
	h := newTestOIDCHandler()
	disabled := false
	h.App.Config().User.PasswordLoginIsEnable = &disabled
	h.App.Config().User.RegisterIsEnable = false
	h.App.Config().OIDC.Providers[0].AutoRegister = true
	repo := new(domainmocks.MockUserRepository)
	repo.On("GetByEmail", mock.Anything, "new@example.invalid").Return(nil, gorm.ErrRecordNotFound)
	repo.On("GetByUsername", mock.Anything, "new_user").Return(nil, gorm.ErrRecordNotFound)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == "new@example.invalid" && u.Username == "new_user"
	})).Return(&domain.User{UID: 42, Username: "new_user"}, nil)
	identities := &passwordDisabledIdentityRepo{}
	h.App.OIDCService = service.NewOIDCService(repo, identities, &passwordDisabledLoginIssuer{})
	h.providerFactory = func(context.Context, internaloidc.ProviderConfig) (oidcProvider, error) {
		return &passwordDisabledOIDCProvider{}, nil
	}
	h.stateStore.Save(internaloidc.State{ProviderID: "dex", State: "state", Nonce: "nonce", CodeVerifier: "verifier", RedirectTo: "/", CreatedAt: time.Now()})
	c, w := newOIDCTestContext("GET", "/api/user/auth/oidc/callback/dex?state=state&code=valid-code")
	c.Params = gin.Params{{Key: "providerID", Value: "dex"}}
	h.Callback(c)
	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), "oidc-login-token")
	require.Contains(t, w.Body.String(), `"uid":"42"`)
	require.NotNil(t, identities.created)
	require.Equal(t, int64(42), identities.created.UID)
	repo.AssertExpectations(t)
}
