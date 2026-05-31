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

func newStorageTestContext(method, url, body string, uid int64) (*gin.Context, *httptest.ResponseRecorder) {
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

func newTestStorageHandler(storageSvc *svcmocks.MockStorageService) *StorageHandler {
	testApp := app.NewTestApp(&app.Services{
		StorageService: storageSvc,
	})
	return NewStorageHandler(testApp)
}

func TestStorageHandler_CreateOrUpdate_Create_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockStorageService)
	mockData := &dto.StorageDTO{ID: 1, Type: "localfs"}

	mockSvc.On("CreateOrUpdate", mock.Anything, int64(1), int64(0), mock.AnythingOfType("*dto.StoragePostRequest")).
		Return(mockData, nil)

	handler := newTestStorageHandler(mockSvc)
	body := `{"type":"localfs", "accessUrlPrefix":"http://cdn"}`
	c, w := newStorageTestContext("POST", "/api/storage", body, 1)

	handler.CreateOrUpdate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.SuccessCreate.Code())
	mockSvc.AssertExpectations(t)
}

func TestStorageHandler_CreateOrUpdate_Update_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockStorageService)
	mockData := &dto.StorageDTO{ID: 1, Type: "localfs"}

	mockSvc.On("CreateOrUpdate", mock.Anything, int64(1), int64(1), mock.AnythingOfType("*dto.StoragePostRequest")).
		Return(mockData, nil)

	handler := newTestStorageHandler(mockSvc)
	body := `{"id":1, "type":"localfs", "accessUrlPrefix":"http://cdn"}`
	c, w := newStorageTestContext("POST", "/api/storage", body, 1)

	handler.CreateOrUpdate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.SuccessUpdate.Code())
	mockSvc.AssertExpectations(t)
}

func TestStorageHandler_List_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockStorageService)
	mockData := []*dto.StorageDTO{
		{ID: 1, Type: "localfs"},
	}

	mockSvc.On("List", mock.Anything, int64(1)).Return(mockData, nil)

	handler := newTestStorageHandler(mockSvc)
	c, w := newStorageTestContext("GET", "/api/storage", "", 1)

	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

func TestStorageHandler_Delete_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockStorageService)
	mockSvc.On("Delete", mock.Anything, int64(1), int64(1)).Return(nil)

	handler := newTestStorageHandler(mockSvc)
	c, w := newStorageTestContext("DELETE", "/api/storage?id=1", "", 1)

	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.SuccessDelete.Code())
	mockSvc.AssertExpectations(t)
}

func TestStorageHandler_EnabledTypes_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockStorageService)
	mockData := []string{"localfs", "s3"}

	mockSvc.On("GetEnabledTypes").Return(mockData, nil)

	handler := newTestStorageHandler(mockSvc)
	c, w := newStorageTestContext("GET", "/api/storage/enabled_types", "", 1)

	handler.EnabledTypes(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

func TestStorageHandler_Validate_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockStorageService)

	mockSvc.On("Validate", mock.Anything, mock.AnythingOfType("*dto.StoragePostRequest")).Return(nil)

	handler := newTestStorageHandler(mockSvc)
	body := `{"type":"localfs", "accessUrlPrefix":"http://cdn"}`
	c, w := newStorageTestContext("POST", "/api/storage/validate", body, 1)

	handler.Validate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}
