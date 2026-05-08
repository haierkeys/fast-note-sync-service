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

// newFileTestContext creates a gin.Context suitable for FileHandler tests.
func newFileTestContext(method, url, body string, uid int64) (*gin.Context, *httptest.ResponseRecorder) {
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

func newTestFileHandler(fileSvc *svcmocks.MockFileService) *FileHandler {
	testApp := app.NewTestApp(&app.Services{
		FileService: fileSvc,
	})
	wss := pkgapp.NewWebsocketServer(pkgapp.WSConfig{}, testApp)
	return NewFileHandler(testApp, wss)
}

// TestFileHandler_Get_Success verifies successful file fetch
func TestFileHandler_Get_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockFileService)
	mockSvc.On("WithClient", "Web", "").Return(mockSvc)

	mockData := &dto.FileDTO{
		ID:       1,
		Path:     "file1.jpg",
		PathHash: "h1",
	}

	mockSvc.On("Get", mock.Anything, int64(1), mock.AnythingOfType("*dto.FileGetRequest")).
		Return(mockData, nil)

	handler := newTestFileHandler(mockSvc)
	c, w := newFileTestContext("GET", "/api/file/info?vault=main&path=file1.jpg", "", 1)

	handler.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestFileHandler_List_Success verifies successful file list fetch
func TestFileHandler_List_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockFileService)
	mockSvc.On("WithClient", "Web", "").Return(mockSvc)

	listData := []*dto.FileDTO{
		{ID: 1, Path: "f1.jpg"},
		{ID: 2, Path: "f2.png"},
	}

	mockSvc.On("List", mock.Anything, int64(1), mock.AnythingOfType("*dto.FileListRequest"), mock.AnythingOfType("*app.Pager")).
		Return(listData, 2, nil)

	handler := newTestFileHandler(mockSvc)
	c, w := newFileTestContext("GET", "/api/files?vault=main&page=1", "", 1)

	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestFileHandler_Delete_Success verifies successful file deletion
func TestFileHandler_Delete_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockFileService)
	mockSvc.On("WithClient", "Web", "").Return(mockSvc)

	deletedFile := &dto.FileDTO{ID: 3, Path: "f3.zip"}
	mockSvc.On("Delete", mock.Anything, int64(1), mock.AnythingOfType("*dto.FileDeleteRequest")).
		Return(deletedFile, nil)

	handler := newTestFileHandler(mockSvc)
	body := `{"vault":"main", "path":"f3.zip", "pathHash":"f3_hash"}`
	c, w := newFileTestContext("DELETE", "/api/file", body, 1)

	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}
