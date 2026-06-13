package api_router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	appconfig "github.com/haierkeys/fast-note-sync-service/internal/config"
	internaloidc "github.com/haierkeys/fast-note-sync-service/internal/oidc"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/stretchr/testify/assert"
)

func newOIDCTestContext(method, url string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, url, nil)
	return c, w
}

func newTestOIDCHandler() *OIDCHandler {
	testApp := app.NewTestApp(&app.Services{})
	cfg := testApp.Config()
	cfg.OIDC = appconfig.OIDCConfig{
		Enabled: true,
		Providers: []appconfig.OIDCProviderConfig{
			{
				ID:           "dex",
				DisplayName:  "Login with Dex",
				Issuer:       "https://dex.example.com",
				ClientID:     "fns-webgui",
				ClientSecret: "secret",
				RedirectURL:  "https://fns.example.com/api/user/auth/oidc/callback/dex",
			},
			{
				ID:           "keycloak",
				DisplayName:  "Login with Keycloak",
				Issuer:       "https://keycloak.example.com/realms/fns",
				ClientID:     "fns-webgui",
				ClientSecret: "secret",
				RedirectURL:  "https://fns.example.com/api/user/auth/oidc/callback/keycloak",
			},
		},
	}
	cfg.OIDC.Normalize()

	handler := NewOIDCHandler(testApp)
	handler.providerFactory = func(ctx context.Context, cfg internaloidc.ProviderConfig) (oidcProvider, error) {
		return &fakeOIDCProvider{issuer: cfg.Issuer}, nil
	}
	return handler
}

func TestOIDCHandlerConfigReturnsMultipleProviders(t *testing.T) {
	handler := newTestOIDCHandler()
	c, w := newOIDCTestContext("GET", "/api/user/auth/oidc/config")

	handler.Config(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())

	var resp struct {
		Data OIDCConfigResponse `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Data.Enabled)
	assert.Equal(t, "Login with Dex", resp.Data.DisplayName)
	assert.Equal(t, "/api/user/auth/oidc/start/dex", resp.Data.StartURL)
	assert.Equal(t, []OIDCProviderConfigResponse{
		{ID: "dex", DisplayName: "Login with Dex", StartURL: "/api/user/auth/oidc/start/dex"},
		{ID: "keycloak", DisplayName: "Login with Keycloak", StartURL: "/api/user/auth/oidc/start/keycloak"},
	}, resp.Data.Providers)
}

func TestOIDCHandlerStartRedirectsToSelectedProvider(t *testing.T) {
	handler := newTestOIDCHandler()
	c, w := newOIDCTestContext("GET", "/api/user/auth/oidc/start/keycloak?redirectTo=/webgui/")
	c.Params = gin.Params{{Key: "providerID", Value: "keycloak"}}

	handler.Start(c)

	assert.Equal(t, http.StatusFound, w.Code)
	location := w.Header().Get("Location")
	assert.True(t, strings.HasPrefix(location, "https://keycloak.example.com/realms/fns/auth?"))
	assert.Contains(t, location, "redirectTo=%2Fwebgui%2F")
}

type fakeOIDCProvider struct {
	issuer string
}

func (p *fakeOIDCProvider) AuthCodeURL(state, nonce, codeVerifier string) string {
	return p.issuer + "/auth?state=" + state + "&redirectTo=%2Fwebgui%2F"
}

func (p *fakeOIDCProvider) Exchange(ctx context.Context, code, codeVerifier, nonce string) (*internaloidc.Claims, error) {
	return &internaloidc.Claims{}, nil
}
