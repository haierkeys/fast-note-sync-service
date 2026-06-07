package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	internaloauth "github.com/haierkeys/fast-note-sync-service/internal/oauth"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errFakeInvalidToken      = errors.New("invalid token")
	errFakeInsufficientScope = internaloauth.ErrInsufficientScope
)

type fakeMCPOAuthVerifier struct {
	claims     *MCPOAuthClaims
	err        error
	calls      int
	clientType string
}

func (v *fakeMCPOAuthVerifier) VerifyBearerToken(ctx context.Context, token string, clientType string) (*MCPOAuthClaims, error) {
	v.calls++
	v.clientType = clientType
	return v.claims, v.err
}

func runMCPMixedAuthMiddleware(t *testing.T, cfg config.OAuthConfig, tokenService *fakeMiddlewareTokenService, verifier MCPOAuthVerifier, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(MCPMixedAuthWithConfig("test-secret", tokenService, MCPMixedAuthConfig{
		OAuthEnabled:         cfg.Enabled,
		Resource:             cfg.Resource,
		AllowStaticFNSToken:  cfg.AllowStaticFNSToken,
		Verifier:             verifier,
		DefaultClient:        cfg.DefaultClient,
		DefaultClientName:    cfg.DefaultClientName,
		DefaultClientVersion: cfg.DefaultClientVersion,
	}))
	router.Any("/api/mcp", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"uid":    pkgapp.GetUID(c),
			"scope":  c.GetString("scope"),
			"client": c.GetHeader("X-Client"),
		})
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func validMCPStaticTokenService() *fakeMiddlewareTokenService {
	return &fakeMiddlewareTokenService{activeToken: &domain.AuthToken{
		ID:          2,
		UID:         1,
		TokenString: "nonce-ok",
		Status:      1,
		Scope:       "p:mcp c:ObsidianPlugin f:*",
		IssueType:   2,
		ExpiredAt:   time.Now().Add(time.Hour),
	}}
}

func TestMCPMixedAuthWithConfig_DisabledUsesUserAuthTokenCompatibility(t *testing.T) {
	token := newMiddlewareJWT(t, "test-secret", "nonce-ok")
	req := httptest.NewRequest(http.MethodGet, "/api/mcp?token="+token, nil)
	req.Header.Set("X-Client", "ObsidianPlugin")
	req.Header.Set("User-Agent", "Obsidian")

	recorder := runMCPMixedAuthMiddleware(t, config.OAuthConfig{}, validMCPStaticTokenService(), nil, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"uid":1`)
	assert.Contains(t, recorder.Body.String(), `"scope":"p:mcp c:ObsidianPlugin f:*"`)
}

func TestMCPMixedAuthWithConfig_EnabledAllowsOldStaticToken(t *testing.T) {
	token := newMiddlewareJWT(t, "test-secret", "nonce-ok")
	verifier := &fakeMCPOAuthVerifier{err: errFakeInvalidToken}
	req := httptest.NewRequest(http.MethodPost, "/api/mcp", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Client", "ObsidianPlugin")
	req.Header.Set("User-Agent", "Obsidian")

	recorder := runMCPMixedAuthMiddleware(t, config.OAuthConfig{
		Enabled:             true,
		Resource:            "https://example.test/api/mcp",
		AllowStaticFNSToken: true,
	}, validMCPStaticTokenService(), verifier, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, 0, verifier.calls)
	assert.Contains(t, recorder.Body.String(), `"uid":1`)
}

func TestMCPMixedAuthWithConfig_MissingOAuthTokenReturnsBearerChallenge(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/mcp", nil)

	recorder := runMCPMixedAuthMiddleware(t, config.OAuthConfig{
		Enabled:  true,
		Resource: "https://example.test/api/mcp",
	}, validMCPStaticTokenService(), nil, req)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Contains(t, recorder.Header().Get("WWW-Authenticate"), `Bearer`)
	assert.Contains(t, recorder.Header().Get("WWW-Authenticate"), `resource="https://example.test/api/mcp"`)
	assert.Contains(t, recorder.Header().Get("WWW-Authenticate"), `resource_metadata="https://example.test/.well-known/oauth-protected-resource/api/mcp"`)
}

func TestMCPMixedAuthWithConfig_InsufficientOAuthScopeReturnsForbidden(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/mcp", nil)
	req.Header.Set("Authorization", "Bearer oauth-token")

	recorder := runMCPMixedAuthMiddleware(t, config.OAuthConfig{
		Enabled:       true,
		Resource:      "https://example.test/api/mcp",
		DefaultClient: "ChatGPT",
	}, validMCPStaticTokenService(), &fakeMCPOAuthVerifier{err: errFakeInsufficientScope}, req)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	assert.Contains(t, recorder.Header().Get("WWW-Authenticate"), `error="insufficient_scope"`)
}

func TestMCPMixedAuthWithConfig_OAuthSuccessSetsUserScopeAndClientDefaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/mcp", nil)
	req.Header.Set("Authorization", "Bearer oauth-token")

	recorder := runMCPMixedAuthMiddleware(t, config.OAuthConfig{
		Enabled:           true,
		Resource:          "https://example.test/api/mcp",
		DefaultClient:     "ChatGPT",
		DefaultClientName: "ChatGPT",
	}, validMCPStaticTokenService(), &fakeMCPOAuthVerifier{claims: &MCPOAuthClaims{
		UID:   7,
		Scope: "p:mcp c:ChatGPT f:*",
	}}, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"uid":7`)
	assert.Contains(t, recorder.Body.String(), `"scope":"p:mcp c:ChatGPT f:*"`)
	assert.Contains(t, recorder.Body.String(), `"client":"ChatGPT"`)
}

func TestRequiredOAuthScopes_DefaultFNSScopeDisablesRequiredScopes(t *testing.T) {
	got := requiredOAuthScopes(config.OAuthConfig{
		RequiredScopes:  []string{"notes:read"},
		DefaultFNSScope: "p:mcp c:* f:*",
	})

	assert.Empty(t, got)
}

func TestRequiredOAuthScopes_UsesConfiguredScopesWithoutDefaultFNSScope(t *testing.T) {
	got := requiredOAuthScopes(config.OAuthConfig{
		RequiredScopes: []string{"notes:read"},
	})

	assert.Equal(t, []string{"notes:read"}, got)
}
