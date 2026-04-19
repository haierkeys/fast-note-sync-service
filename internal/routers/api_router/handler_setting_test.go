package api_router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	svcmocks "github.com/haierkeys/fast-note-sync-service/internal/service/mocks"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newSettingTestContext(method, url, body string, uid int64) (*gin.Context, *httptest.ResponseRecorder) {
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

func newTestSettingHandler(settingSvc *svcmocks.MockSettingService) *SettingHandler {
	testApp := app.NewTestApp(&app.Services{
		SettingService: settingSvc,
	})
	wss := pkgapp.NewWebsocketServer(pkgapp.WSConfig{}, testApp)
	return NewSettingHandler(testApp, wss)
}

func TestSettingHandler_Get_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockSettingService)
	mockData := &dto.SettingDTO{Path: "theme", Content: "dark"}

	mockSvc.On("Get", mock.Anything, int64(1), mock.AnythingOfType("*dto.SettingGetRequest")).
		Return(mockData, nil)

	handler := newTestSettingHandler(mockSvc)
	c, w := newSettingTestContext("GET", "/api/setting?path=theme&vault=main", "", 1)

	handler.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

func TestSettingHandler_List_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockSettingService)
	mockData := []*dto.SettingDTO{
		{Path: "theme", Content: "dark"},
	}

	mockSvc.On("List", mock.Anything, int64(1), mock.AnythingOfType("*dto.SettingListRequest"), mock.AnythingOfType("*app.Pager")).
		Return(mockData, int64(1), nil)

	handler := newTestSettingHandler(mockSvc)
	c, w := newSettingTestContext("GET", "/api/settings?vault=main", "", 1)

	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

func TestSettingHandler_CreateOrUpdate_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockSettingService)
	mockData := &dto.SettingDTO{Path: "theme", Content: "light"}

	mockSvc.On("ModifyOrCreate", mock.Anything, int64(1), mock.AnythingOfType("*dto.SettingModifyOrCreateRequest"), false).
		Return(true, mockData, nil)

	handler := newTestSettingHandler(mockSvc)
	body := `{"vault":"main", "path":"theme", "content":"light"}`
	c, w := newSettingTestContext("POST", "/api/setting", body, 1)

	handler.CreateOrUpdate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

func TestSettingHandler_Delete_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockSettingService)
	mockData := &dto.SettingDTO{Path: "theme"}

	mockSvc.On("Delete", mock.Anything, int64(1), mock.AnythingOfType("*dto.SettingDeleteRequest")).
		Return(mockData, nil)

	handler := newTestSettingHandler(mockSvc)
	body := `{"vault":"main", "path":"theme"}`
	c, w := newSettingTestContext("DELETE", "/api/setting", body, 1)

	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

func TestSettingHandler_Rename_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockSettingService)
	mockData := &dto.SettingDTO{Path: "new_theme", Content: "dark"}

	mockSvc.On("Rename", mock.Anything, int64(1), mock.AnythingOfType("*dto.SettingRenameRequest")).
		Return(mockData, nil)

	handler := newTestSettingHandler(mockSvc)
	body := `{"vault":"main", "oldPath":"theme", "newPath":"new_theme"}`
	c, w := newSettingTestContext("POST", "/api/setting/rename", body, 1)

	handler.Rename(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}
