package api_router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service/mocks"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newAdminTestContext(method, url, body string, uid int64) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, url, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	if uid > 0 {
		c.Set("user_token", &pkgapp.UserEntity{UID: uid})
	}
	return c, w
}

func newTestAdminHandler() (*AdminControlHandler, *app.App, *mocks.MockUserService) {
	mockUserSvc := new(mocks.MockUserService)
	svcs := &app.Services{
		UserService: mockUserSvc,
	}
	testApp := app.NewTestApp(svcs)
	// Set mock config values
	cfg := testApp.Config()
	cfg.User.AdminUID = 1
	cfg.WebGUI.FontSet = "Inter"

	wss := pkgapp.NewWebsocketServer(pkgapp.WSConfig{}, testApp)
	return NewAdminControlHandler(testApp, wss), testApp, mockUserSvc
}

func TestAdminControlHandler_Config_Success(t *testing.T) {
	handler, _, mockUserSvc := newTestAdminHandler()
	c, w := newAdminTestContext("GET", "/api/webgui/config", "", 0)

	mockUserSvc.On("IsRegisterEnabled", mock.Anything).Return(true)

	handler.Config(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())

	var resp struct {
		Data dto.AdminWebGUIConfig `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Inter", resp.Data.FontSet)
	assert.True(t, resp.Data.RegisterIsEnable)

	mockUserSvc.AssertExpectations(t)
}

func TestAdminControlHandler_Config_RegisterDisabled(t *testing.T) {
	handler, _, mockUserSvc := newTestAdminHandler()
	c, w := newAdminTestContext("GET", "/api/webgui/config", "", 0)

	mockUserSvc.On("IsRegisterEnabled", mock.Anything).Return(false)

	handler.Config(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())

	var resp struct {
		Data dto.AdminWebGUIConfig `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.False(t, resp.Data.RegisterIsEnable)

	mockUserSvc.AssertExpectations(t)
}

func TestAdminControlHandler_GetConfig_Success(t *testing.T) {
	handler, _, _ := newTestAdminHandler()
	c, w := newAdminTestContext("GET", "/api/admin/config", "", 1) // UID 1 is admin

	handler.GetConfig(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
}

func TestAdminControlHandler_GetConfig_Forbidden(t *testing.T) {
	handler, _, _ := newTestAdminHandler()
	c, w := newAdminTestContext("GET", "/api/admin/config", "", 2) // UID 2 is not admin

	handler.GetConfig(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.ErrorUserIsNotAdmin.Code())
}

func TestAdminControlHandler_GetSystemInfo_Success(t *testing.T) {
	handler, _, _ := newTestAdminHandler()
	c, w := newAdminTestContext("GET", "/api/admin/system/info", "", 1)

	handler.GetSystemInfo(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	assert.Contains(t, w.Body.String(), `"uptime"`)
}

func TestAdminControlHandler_GC_Success(t *testing.T) {
	handler, _, _ := newTestAdminHandler()
	c, w := newAdminTestContext("GET", "/api/admin/gc", "", 1)

	handler.GC(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	assert.Contains(t, w.Body.String(), "Manual GC completed")
}
