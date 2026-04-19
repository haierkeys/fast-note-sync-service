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

// newNoteHistoryTestContext creates a gin.Context suitable for NoteHistoryHandler tests.
func newNoteHistoryTestContext(method, url, body string, uid int64) (*gin.Context, *httptest.ResponseRecorder) {
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

func newTestNoteHistoryHandler(historySvc *svcmocks.MockNoteHistoryService) *NoteHistoryHandler {
	testApp := app.NewTestApp(&app.Services{
		NoteHistoryService: historySvc,
	})
	wss := pkgapp.NewWebsocketServer(pkgapp.WSConfig{}, testApp)
	return NewNoteHistoryHandler(testApp, wss)
}

// TestNoteHistoryHandler_Get_Success verifies successful history fetch
func TestNoteHistoryHandler_Get_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockNoteHistoryService)
	
	mockData := &dto.NoteHistoryDTO{
		ID:       10,
		NoteID:   1,
		Path:     "note1.md",
		Content:  "old content",
	}

	mockSvc.On("Get", mock.Anything, int64(1), int64(10)).Return(mockData, nil)

	handler := newTestNoteHistoryHandler(mockSvc)
	c, w := newNoteHistoryTestContext("GET", "/api/note/history?id=10", "", 1)
	
	handler.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestNoteHistoryHandler_List_Success verifies successful history list fetch
func TestNoteHistoryHandler_List_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockNoteHistoryService)
	
	listData := []*dto.NoteHistoryNoContentDTO{
		{ID: 10, NoteID: 1, Path: "note1.md"},
		{ID: 11, NoteID: 1, Path: "note1.md"},
	}

	mockSvc.On("List", mock.Anything, int64(1), mock.AnythingOfType("*dto.NoteHistoryListRequest"), mock.AnythingOfType("*app.Pager")).
		Return(listData, int64(2), nil)

	handler := newTestNoteHistoryHandler(mockSvc)
	c, w := newNoteHistoryTestContext("GET", "/api/note/histories?vault=main&path=note1.md", "", 1)
	
	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestNoteHistoryHandler_Restore_Success verifies successful restore from history
func TestNoteHistoryHandler_Restore_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockNoteHistoryService)
	
	restoredNote := &dto.NoteDTO{ID: 1, Path: "note1.md", Action: ""}
	mockSvc.On("RestoreFromHistory", mock.Anything, int64(1), int64(10)).
		Return(restoredNote, nil)

	handler := newTestNoteHistoryHandler(mockSvc)
	body := `{"vault":"main", "historyId":10}`
	c, w := newNoteHistoryTestContext("PUT", "/api/note/history/restore", body, 1)

	handler.Restore(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}
