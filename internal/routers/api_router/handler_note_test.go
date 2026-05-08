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

// newNoteTestContext creates a gin.Context suitable for NoteHandler tests.
func newNoteTestContext(method, url, body string, uid int64) (*gin.Context, *httptest.ResponseRecorder) {
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

// newTestNoteHandler creates a NoteHandler with mock services.
func newTestNoteHandler(noteSvc *svcmocks.MockNoteService, fileSvc *svcmocks.MockFileService) *NoteHandler {
	testApp := app.NewTestApp(&app.Services{
		NoteService: noteSvc,
		FileService: fileSvc,
	})
	// WSS is used in NoteHandler for broadcasting
	wss := pkgapp.NewWebsocketServer(pkgapp.WSConfig{}, testApp)
	return NewNoteHandler(testApp, wss)
}

// TestNoteHandler_Get_Success verifies successful note fetch
func TestNoteHandler_Get_Success(t *testing.T) {
	mockNoteSvc := new(svcmocks.MockNoteService)
	mockNoteSvc.On("WithClient", "Web", "").Return(mockNoteSvc)
	mockFileSvc := new(svcmocks.MockFileService)
	mockFileSvc.On("WithClient", "Web", "").Return(mockFileSvc)

	mockData := &dto.NoteDTO{
		ID:       1,
		Path:     "test.md",
		PathHash: "hash123",
		Content:  "content",
	}

	mockNoteSvc.On("Get", mock.Anything, int64(1), mock.AnythingOfType("*dto.NoteGetRequest")).
		Return(mockData, nil)

	mockFileSvc.On("ResolveEmbedLinks", mock.Anything, int64(1), "main", "test.md", "content").
		Return(map[string]string{}, nil)

	handler := newTestNoteHandler(mockNoteSvc, mockFileSvc)
	c, w := newNoteTestContext("GET", "/api/note", "", 1)

	// Simulate binding query params
	c.Request.URL.RawQuery = "vault=main&path=test.md"
	handler.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockNoteSvc.AssertExpectations(t)
}

// TestNoteHandler_List_Success verifies successful note list fetch
func TestNoteHandler_List_Success(t *testing.T) {
	mockNoteSvc := new(svcmocks.MockNoteService)
	mockNoteSvc.On("WithClient", "Web", "").Return(mockNoteSvc)
	listData := []*dto.NoteNoContentDTO{
		{ID: 1, Path: "test1.md"},
		{ID: 2, Path: "test2.md"},
	}

	mockNoteSvc.On("List", mock.Anything, int64(1), mock.AnythingOfType("*dto.NoteListRequest"), mock.AnythingOfType("*app.Pager")).
		Return(listData, 2, nil)

	handler := newTestNoteHandler(mockNoteSvc, nil)
	c, w := newNoteTestContext("GET", "/api/notes", "", 1)
	c.Request.URL.RawQuery = "vault=main&page=1&size=10"

	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockNoteSvc.AssertExpectations(t)
}

// TestNoteHandler_CreateOrUpdate_Success verifies successful note creation/update
func TestNoteHandler_CreateOrUpdate_Success(t *testing.T) {
	mockNoteSvc := new(svcmocks.MockNoteService)
	mockNoteSvc.On("WithClient", "Web", "").Return(mockNoteSvc)

	mockNoteSvc.On("UpdateCheck", mock.Anything, int64(1), mock.AnythingOfType("*dto.NoteUpdateCheckRequest")).
		Return("", (*dto.NoteDTO)(nil), nil)

	createdNote := &dto.NoteDTO{ID: 2, Path: "create.md"}
	mockNoteSvc.On("ModifyOrCreate", mock.Anything, int64(1), mock.AnythingOfType("*dto.NoteModifyOrCreateRequest"), false).
		Return(true, createdNote, nil)

	handler := newTestNoteHandler(mockNoteSvc, nil)
	body := `{"vault":"main", "path":"create.md", "content":"hello"}`
	c, w := newNoteTestContext("POST", "/api/note", body, 1)

	handler.CreateOrUpdate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockNoteSvc.AssertExpectations(t)
}

// TestNoteHandler_Delete_Success verifies successful note deletion
func TestNoteHandler_Delete_Success(t *testing.T) {
	mockNoteSvc := new(svcmocks.MockNoteService)
	mockNoteSvc.On("WithClient", "Web", "").Return(mockNoteSvc)

	existingNote := &dto.NoteDTO{ID: 3, Path: "del.md", Action: ""}
	mockNoteSvc.On("Get", mock.Anything, int64(1), mock.AnythingOfType("*dto.NoteGetRequest")).
		Return(existingNote, nil)

	mockNoteSvc.On("Delete", mock.Anything, int64(1), mock.AnythingOfType("*dto.NoteDeleteRequest")).
		Return(existingNote, nil)

	handler := newTestNoteHandler(mockNoteSvc, nil)
	c, w := newNoteTestContext("DELETE", "/api/note", "", 1)
	c.Request.URL.RawQuery = "vault=main&path=del.md"

	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockNoteSvc.AssertExpectations(t)
}
