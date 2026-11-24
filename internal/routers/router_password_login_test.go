package routers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/stretchr/testify/require"
)

type passwordLoginRouteTokens struct {
	service.TokenService
	created bool
}

func (s *passwordLoginRouteTokens) GetActiveToken(context.Context, int64, int64) (*domain.AuthToken, error) {
	return &domain.AuthToken{ID: 1, UID: 1, IssueType: 1, ClientType: "webgui", Scope: "p:* c:* f:*"}, nil
}
func (s *passwordLoginRouteTokens) RecordAccessLog(context.Context, *domain.AuthTokenLog) error {
	return nil
}
func (s *passwordLoginRouteTokens) Create(_ context.Context, uid int64, params *dto.TokenIssueRequest) (*dto.TokenCreateResponse, error) {
	s.created = true
	return &dto.TokenCreateResponse{TokenString: "manual-plugin-token"}, nil
}
func (s *passwordLoginRouteTokens) ListByUser(context.Context, int64) ([]*dto.TokenResponse, error) {
	return []*dto.TokenResponse{}, nil
}
func (s *passwordLoginRouteTokens) GetRecentClients(context.Context, int64, time.Duration) (map[int64][]string, error) {
	return nil, nil
}

func TestPasswordLoginDisabledActualAPIRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := &passwordLoginRouteTokens{}
	a := app.NewTestApp(&app.Services{TokenService: tokens})
	a.Repositories = &app.Repositories{}
	require.NoError(t, defaults.Set(a.Config()))
	disabled := false
	a.Config().User.PasswordLoginIsEnable = &disabled
	a.Config().User.RegisterIsEnable = true
	a.Config().Security.AuthTokenKey = "test-key"
	wss := pkgapp.NewWebsocketServer(pkgapp.WSConfig{}, a)
	r := gin.New()
	locale := en.New()
	registerAPIRoutes(r, a, wss, ut.New(locale, locale))
	token, err := pkgapp.NewTokenManager(pkgapp.TokenConfig{SecretKey: "test-key"}).Generate(1, "user", "", 1, "")
	require.NoError(t, err)
	for _, tc := range []struct {
		method, path, body string
		want               int
	}{
		{"POST", "/api/user/login", "not-json", code.ErrorUserPasswordLoginIsDisable.Code()},
		{"POST", "/api/user/register", "not-json", code.ErrorUserPasswordLoginIsDisable.Code()},
		{"POST", "/api/user/change_password", "not-json", code.ErrorUserPasswordLoginIsDisable.Code()},
		{"POST", "/api/token", `{"clientType":"obsidian","scope":"rest,ws","expiredDays":30}`, code.Success.Code()},
		{"GET", "/api/tokens", "", code.Success.Code()},
	} {
		t.Run(tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("X-Client", "WebGUI")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			r.ServeHTTP(w, req)
			require.Equal(t, 200, w.Code)
			var res struct {
				Code   int  `json:"code"`
				Status bool `json:"status"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
			require.Equal(t, tc.want, res.Code, w.Body.String())
			require.Equal(t, tc.want == code.Success.Code(), res.Status)
		})
	}
	require.True(t, tokens.created)
}
