package api_router

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	internaloidc "github.com/haierkeys/fast-note-sync-service/internal/oidc"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	apperrors "github.com/haierkeys/fast-note-sync-service/pkg/errors"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
)

type OIDCHandler struct {
	*Handler
	stateStore      *internaloidc.StateStore
	providerFactory oidcProviderFactory
}

type oidcProvider interface {
	AuthCodeURL(state, nonce, codeVerifier string) string
	Exchange(ctx context.Context, code, codeVerifier, nonce string) (*internaloidc.Claims, error)
}

type oidcProviderFactory func(ctx context.Context, cfg internaloidc.ProviderConfig) (oidcProvider, error)

type OIDCConfigResponse struct {
	Enabled     bool   `json:"enabled"`
	DisplayName string `json:"displayName"`
	StartURL    string `json:"startUrl"`
}

func NewOIDCHandler(a *app.App) *OIDCHandler {
	return &OIDCHandler{
		Handler:    NewHandler(a),
		stateStore: internaloidc.NewStateStore(5 * time.Minute),
		providerFactory: func(ctx context.Context, cfg internaloidc.ProviderConfig) (oidcProvider, error) {
			return internaloidc.NewProvider(ctx, cfg)
		},
	}
}

func (h *OIDCHandler) Config(c *gin.Context) {
	cfg := h.App.Config().OIDC
	response := pkgapp.NewResponse(c)
	response.ToResponse(code.Success.WithData(OIDCConfigResponse{
		Enabled:     cfg.Enabled,
		DisplayName: cfg.DisplayName,
		StartURL:    "/api/user/auth/oidc/start",
	}))
}

func (h *OIDCHandler) Start(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config().OIDC
	if !cfg.Enabled {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("oidc is disabled"))
		return
	}

	provider, err := h.provider(c.Request.Context())
	if err != nil {
		h.App.Logger().Error("OIDCHandler.Start.Provider", zap.Error(err))
		response.ToResponse(code.ErrorInvalidParams.WithDetails("oidc provider discovery failed"))
		return
	}

	state := internaloidc.State{
		State:        util.GetRandomString(32),
		Nonce:        util.GetRandomString(32),
		CodeVerifier: util.GetRandomString(64),
		RedirectTo:   normalizeOIDCRedirect(c.Query("redirectTo")),
		CreatedAt:    time.Now(),
	}
	h.stateStore.Save(state)
	c.Redirect(http.StatusFound, provider.AuthCodeURL(state.State, state.Nonce, state.CodeVerifier))
}

func (h *OIDCHandler) Callback(c *gin.Context) {
	stateValue := c.Query("state")
	state, ok := h.stateStore.Consume(stateValue)
	if !ok {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(renderOIDCCallbackErrorHTML("OIDC state is invalid or expired")))
		return
	}
	if errMsg := strings.TrimSpace(c.Query("error")); errMsg != "" {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(renderOIDCCallbackErrorHTML(errMsg)))
		return
	}

	provider, err := h.provider(c.Request.Context())
	if err != nil {
		h.App.Logger().Error("OIDCHandler.Callback.Provider", zap.Error(err))
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(renderOIDCCallbackErrorHTML("OIDC provider discovery failed")))
		return
	}

	claims, err := provider.Exchange(c.Request.Context(), c.Query("code"), state.CodeVerifier, state.Nonce)
	if err != nil {
		h.App.Logger().Error("OIDCHandler.Callback.Exchange", zap.Error(err))
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(renderOIDCCallbackErrorHTML("OIDC token exchange failed")))
		return
	}

	user, err := h.App.OIDCService.Authenticate(c.Request.Context(), *claims, c.ClientIP(), "WebGUI", c.GetHeader("User-Agent"))
	if err != nil {
		h.App.Logger().Error("OIDCHandler.Callback.Authenticate", zap.Error(err))
		apperrors.ErrorResponse(c, err)
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(renderOIDCCallbackSuccessHTML(user, state.RedirectTo)))
}

func (h *OIDCHandler) provider(ctx context.Context) (oidcProvider, error) {
	cfg := h.App.Config().OIDC
	return h.providerFactory(ctx, internaloidc.ProviderConfig{
		Issuer:       cfg.Issuer,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
	})
}

func normalizeOIDCRedirect(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "//") || strings.Contains(value, "://") {
		return "/"
	}
	if !strings.HasPrefix(value, "/") {
		return "/"
	}
	return value
}

func renderOIDCCallbackSuccessHTML(user *dto.UserDTO, redirectTo string) string {
	payload := map[string]interface{}{
		"token":    user.Token,
		"username": user.Username,
		"uid":      fmt.Sprintf("%d", user.UID),
		"avatar":   user.Avatar,
		"email":    user.Email,
		"tokenId":  fmt.Sprintf("%d", user.TokenID),
		"user":     "true",
	}
	raw, _ := json.Marshal(payload)
	return fmt.Sprintf(`<!doctype html>
<html>
<head><meta charset="utf-8"><title>OIDC Login</title></head>
<body>
<script>
const auth = %s;
for (const [key, value] of Object.entries(auth)) {
  if (value !== undefined && value !== null) localStorage.setItem(key, String(value));
}
window.location.replace(%q);
</script>
</body>
</html>`, raw, normalizeOIDCRedirect(redirectTo))
}

func renderOIDCCallbackErrorHTML(message string) string {
	return fmt.Sprintf(`<!doctype html>
<html>
<head><meta charset="utf-8"><title>OIDC Login Failed</title></head>
<body>OIDC login failed: %s</body>
</html>`, html.EscapeString(message))
}
