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

func newBackupTestContext(method, url, body string, uid int64) (*gin.Context, *httptest.ResponseRecorder) {
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

func newTestBackupHandler(backupSvc *svcmocks.MockBackupService) *BackupHandler {
	testApp := app.NewTestApp(&app.Services{
		BackupService: backupSvc,
	})
	return NewBackupHandler(testApp)
}

// TestBackupHandler_GetConfigs_Success verifies successful retrieval of backup configs
func TestBackupHandler_GetConfigs_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockBackupService)

	mockData := []*dto.BackupConfigDTO{
		{ID: 1, Vault: "main"},
	}

	mockSvc.On("GetConfigs", mock.Anything, int64(1)).Return(mockData, nil)

	handler := newTestBackupHandler(mockSvc)
	c, w := newBackupTestContext("GET", "/api/backup/configs", "", 1)

	handler.GetConfigs(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestBackupHandler_UpdateConfig_Success verifies successful backup config update
func TestBackupHandler_UpdateConfig_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockBackupService)

	mockData := &dto.BackupConfigDTO{
		ID:    1,
		Vault: "main",
	}

	mockSvc.On("UpdateConfig", mock.Anything, int64(1), mock.AnythingOfType("*dto.BackupConfigRequest")).
		Return(mockData, nil)

	handler := newTestBackupHandler(mockSvc)
	body := `{"id":1, "vault":"main", "type":"sync", "storageIds":"[1]", "cronStrategy":"daily"}`
	c, w := newBackupTestContext("POST", "/api/backup/config", body, 1)

	handler.UpdateConfig(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.SuccessUpdate.Code())
	mockSvc.AssertExpectations(t)
}

// TestBackupHandler_DeleteConfig_Success verifies successful backup config deletion
func TestBackupHandler_DeleteConfig_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockBackupService)

	mockSvc.On("DeleteConfig", mock.Anything, int64(1), int64(1)).Return(nil)

	handler := newTestBackupHandler(mockSvc)
	// Delete request with body since c.ShouldBind parses body instead of query
	body := `{"id":1}`
	c, w := newBackupTestContext("DELETE", "/api/backup/config", body, 1)

	handler.DeleteConfig(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestBackupHandler_ListHistory_Success verifies successful history fetch
func TestBackupHandler_ListHistory_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockBackupService)

	mockData := []*dto.BackupHistoryDTO{
		{ID: 10, ConfigID: 1},
	}

	mockSvc.On("ListHistory", mock.Anything, int64(1), int64(1), mock.AnythingOfType("*app.Pager")).
		Return(mockData, int64(1), nil)

	handler := newTestBackupHandler(mockSvc)
	c, w := newBackupTestContext("GET", "/api/backup/histories?configId=1", "", 1)

	handler.ListHistory(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestBackupHandler_Execute_Success verifies manual trigger backup execution
func TestBackupHandler_Execute_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockBackupService)

	mockSvc.On("ExecuteUserBackup", mock.Anything, int64(1), int64(1)).Return(nil)

	handler := newTestBackupHandler(mockSvc)
	body := `{"id":1}`
	c, w := newBackupTestContext("POST", "/api/backup/execute", body, 1)

	handler.Execute(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}
