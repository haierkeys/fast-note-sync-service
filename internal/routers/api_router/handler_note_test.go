package api_router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/mocks"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNoteHandler_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNoteSvc := new(mocks.NoteServiceMock)
	appObj := &app.App{
		Services: &app.Services{
			NoteService: mockNoteSvc,
		},
	}
	// Note: We might need to initialize more Infra/Services if used
	// By default, h.App.Logger() might be used.
	
	h := &NoteHandler{
		Handler: &Handler{
			App: appObj,
		},
	}

	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		// Setup URL and parameters
		c.Request, _ = http.NewRequest("GET", "/api/note?vault=test&pathHash=hash123", nil)
		c.Set("UID", int64(1)) // Mock authentication middleware

		mockNoteSvc.On("WithClient", app.WebClientName, "").Return(mockNoteSvc)
		mockNoteSvc.On("Get", mock.Anything, int64(1), mock.AnythingOfType("*dto.NoteGetRequest")).Return(&dto.NoteDTO{
			ID:       1,
			Path:     "note1.md",
			PathHash: "hash123",
			Content:  "hello",
		}, nil)

		// NoteHandler might use other services like FileService
		// If h.App.FileService is used, we need to mock it too.
		
		h.Get(c)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var res pkgapp.Res
		err := json.Unmarshal(w.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Equal(t, 0, res.Code) // Success code is usually 0
		
		mockNoteSvc.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/api/note", nil)
		// No UID set

		h.Get(c)

		assert.Equal(t, http.StatusOK, w.Code) // Project might return 200 with error code in body
		var res pkgapp.Res
		json.Unmarshal(w.Body.Bytes(), &res)
		assert.NotEqual(t, 0, res.Code)
	})
}
