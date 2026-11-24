package api_router

import (
	"encoding/json"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPasswordLoginDisabledHandlersRejectBeforeBinding(t *testing.T) {
	disabled := false
	testApp := app.NewTestApp(&app.Services{}) // No DB, token service or UserService may be called.
	testApp.Config().User.PasswordLoginIsEnable = &disabled
	h := NewUserHandler(testApp)
	for path, handler := range map[string]gin.HandlerFunc{
		"/api/user/login":           h.Login,
		"/api/user/register":        h.Register,
		"/api/user/change_password": h.UserChangePassword,
	} {
		for _, body := range []string{"", "not-json", `{}`, `{"credentials":"user","password":"wrong"}`} {
			t.Run(path+body, func(t *testing.T) {
				c, w := newUserTestContext("POST", path, body, 1)
				handler(c)
				require.Equal(t, 200, w.Code)
				assertResponseCode(t, w, code.ErrorUserPasswordLoginIsDisable.Code())
				var res struct {
					Status bool `json:"status"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
				require.False(t, res.Status)
			})
		}
	}
}

func TestPasswordLoginPublicCapabilities(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		testApp := app.NewTestApp(&app.Services{})
		cfg := testApp.Config()
		cfg.User.RegisterIsEnable = true
		cfg.User.AdminUID = 1
		if !enabled {
			cfg.User.PasswordLoginIsEnable = &enabled
		}
		cfg.Security.AuthTokenKey = "DO-NOT-EXPOSE-TOKEN-KEY"
		cfg.OIDC.ClientSecret = "DO-NOT-EXPOSE-OIDC-SECRET"
		testApp.UserService = service.NewUserService(nil, nil, nil, zap.NewNop(), &service.ServiceConfig{
			User: service.UserServiceConfig{RegisterIsEnable: true, AdminUID: 1, PasswordLoginIsEnable: cfg.User.PasswordLoginIsEnable},
		})
		c, w := newAdminTestContext("GET", "/api/webgui/config", "", 0)
		NewAdminControlHandler(testApp, nil).Config(c)
		var res struct {
			Data map[string]interface{} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
		require.Equal(t, enabled, res.Data["passwordLoginIsEnable"])
		require.Equal(t, enabled, res.Data["registerIsEnable"])
		require.Len(t, res.Data, 4)
		require.NotContains(t, w.Body.String(), "DO-NOT-EXPOSE")
	}
}
