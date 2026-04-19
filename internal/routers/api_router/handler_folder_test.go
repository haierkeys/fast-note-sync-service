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

// newFolderTestContext creates a gin.Context suitable for FolderHandler tests.
func newFolderTestContext(method, url, body string, uid int64) (*gin.Context, *httptest.ResponseRecorder) {
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

func newTestFolderHandler(folderSvc *svcmocks.MockFolderService) *FolderHandler {
	testApp := app.NewTestApp(&app.Services{
		FolderService: folderSvc,
	})
	return NewFolderHandler(testApp)
}

// TestFolderHandler_Get_Success verifies successful folder fetch
func TestFolderHandler_Get_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockFolderService)
	
	mockData := &dto.FolderDTO{
		ID:       1,
		Path:     "folder1",
		PathHash: "h1",
	}

	mockSvc.On("Get", mock.Anything, int64(1), mock.AnythingOfType("*dto.FolderGetRequest")).
		Return(mockData, nil)

	handler := newTestFolderHandler(mockSvc)
	c, w := newFolderTestContext("GET", "/api/folder", "", 1)
	c.Request.URL.RawQuery = "vault=main&path=folder1"
	
	handler.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestFolderHandler_List_Success verifies successful folder list fetch
func TestFolderHandler_List_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockFolderService)
	
	listData := []*dto.FolderDTO{
		{ID: 1, Path: "f1"},
		{ID: 2, Path: "f2"},
	}

	mockSvc.On("List", mock.Anything, int64(1), mock.AnythingOfType("*dto.FolderListRequest")).
		Return(listData, nil)

	handler := newTestFolderHandler(mockSvc)
	c, w := newFolderTestContext("GET", "/api/folders", "", 1)
	c.Request.URL.RawQuery = "vault=main"
	
	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestFolderHandler_Create_Success verifies successful folder creation
func TestFolderHandler_Create_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockFolderService)
	
	createdFolder := &dto.FolderDTO{ID: 3, Path: "f3"}
	mockSvc.On("UpdateOrCreate", mock.Anything, int64(1), mock.AnythingOfType("*dto.FolderCreateRequest")).
		Return(createdFolder, nil)

	handler := newTestFolderHandler(mockSvc)
	body := `{"vault":"main", "path":"f3"}`
	c, w := newFolderTestContext("POST", "/api/folder", body, 1)

	handler.Create(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestFolderHandler_Delete_Success verifies successful folder deletion
func TestFolderHandler_Delete_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockFolderService)
	
	mockSvc.On("Delete", mock.Anything, int64(1), mock.AnythingOfType("*dto.FolderDeleteRequest")).
		Return((*dto.FolderDTO)(nil), nil)

	handler := newTestFolderHandler(mockSvc)
	body := `{"vault":"main", "path":"f3"}`
	c, w := newFolderTestContext("DELETE", "/api/folder", body, 1)

	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}
