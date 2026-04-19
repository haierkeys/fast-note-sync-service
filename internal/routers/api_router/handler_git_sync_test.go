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

func newGitSyncTestContext(method, url, body string, uid int64) (*gin.Context, *httptest.ResponseRecorder) {
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

func newTestGitSyncHandler(gitSvc *svcmocks.MockGitSyncService) *GitSyncHandler {
	testApp := app.NewTestApp(&app.Services{
		GitSyncService: gitSvc,
	})
	return NewGitSyncHandler(testApp)
}

// TestGitSyncHandler_GetConfigs_Success verifies successful retrieval of git configs
func TestGitSyncHandler_GetConfigs_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockGitSyncService)

	mockData := []*dto.GitSyncConfigDTO{
		{ID: 1, RepoURL: "https://github.com/user/repo.git"},
	}

	mockSvc.On("GetConfigs", mock.Anything, int64(1)).Return(mockData, nil)

	handler := newTestGitSyncHandler(mockSvc)
	c, w := newGitSyncTestContext("GET", "/api/git-sync/configs", "", 1)

	handler.GetConfigs(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestGitSyncHandler_UpdateConfig_Success verifies successful git sync update
func TestGitSyncHandler_UpdateConfig_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockGitSyncService)

	mockData := &dto.GitSyncConfigDTO{
		ID:      1,
		RepoURL: "https://github.com/user/repo.git",
	}

	mockSvc.On("UpdateConfig", mock.Anything, int64(1), mock.AnythingOfType("*dto.GitSyncConfigRequest")).
		Return(mockData, nil)

	handler := newTestGitSyncHandler(mockSvc)
	body := `{"id":1, "repoUrl":"https://github.com/user/repo.git"}`
	c, w := newGitSyncTestContext("POST", "/api/git-sync/config", body, 1)

	handler.UpdateConfig(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.SuccessUpdate.Code())
	mockSvc.AssertExpectations(t)
}

// TestGitSyncHandler_DeleteConfig_Success verifies successful deletion
func TestGitSyncHandler_DeleteConfig_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockGitSyncService)

	mockSvc.On("DeleteConfig", mock.Anything, int64(1), int64(1)).Return(nil)

	handler := newTestGitSyncHandler(mockSvc)
	body := `{"id":1}`
	c, w := newGitSyncTestContext("DELETE", "/api/git-sync/config", body, 1)

	handler.DeleteConfig(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestGitSyncHandler_Validate_Success verifies git sync param validation
func TestGitSyncHandler_Validate_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockGitSyncService)

	mockSvc.On("Validate", mock.Anything, mock.AnythingOfType("*dto.GitSyncValidateRequest")).Return(nil)

	handler := newTestGitSyncHandler(mockSvc)
	body := `{"repoUrl":"https://github.com/user/repo.git"}`
	c, w := newGitSyncTestContext("POST", "/api/git-sync/validate", body, 1)

	handler.Validate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestGitSyncHandler_Execute_Success verifies manual trigger
func TestGitSyncHandler_Execute_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockGitSyncService)

	mockSvc.On("ExecuteSync", mock.Anything, int64(1), int64(1)).Return(nil)

	handler := newTestGitSyncHandler(mockSvc)
	body := `{"id":1}`
	c, w := newGitSyncTestContext("POST", "/api/git-sync/config/execute", body, 1)

	handler.Execute(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestGitSyncHandler_CleanWorkspace_Success verifies workspace clean
func TestGitSyncHandler_CleanWorkspace_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockGitSyncService)

	mockSvc.On("CleanWorkspace", mock.Anything, int64(1), int64(1)).Return(nil)

	handler := newTestGitSyncHandler(mockSvc)
	body := `{"configId":1}`
	c, w := newGitSyncTestContext("DELETE", "/api/git-sync/config/clean", body, 1)

	handler.CleanWorkspace(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestGitSyncHandler_ListHistory_Success verifies getting history
func TestGitSyncHandler_ListHistory_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockGitSyncService)

	mockData := []*dto.GitSyncHistoryDTO{
		{ID: 10, ConfigID: 1},
	}

	mockSvc.On("ListHistory", mock.Anything, int64(1), int64(1), mock.AnythingOfType("*app.Pager")).
		Return(mockData, int64(1), nil)

	handler := newTestGitSyncHandler(mockSvc)
	c, w := newGitSyncTestContext("GET", "/api/git-sync/histories?configId=1", "", 1)

	handler.GetHistories(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}
