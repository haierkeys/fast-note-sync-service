package api_router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/stretchr/testify/assert"
)

func newVersionTestContext(url string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", url, nil)
	c.Request = req
	return c, w
}

func TestVersionHandler_ServerVersion_Success(t *testing.T) {
	testApp := app.NewTestApp(nil)
	handler := NewVersionHandler(testApp)
	c, w := newVersionTestContext("/api/version")

	handler.ServerVersion(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	
	// Verify that the response contains the version from app.Version
	var resp struct {
		Data struct {
			Version string `json:"version"`
		} `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, app.Version, resp.Data.Version)
}

func TestVersionHandler_Support_Success(t *testing.T) {
	testApp := app.NewTestApp(nil)
	
	// Inject some support records
	records := []pkgapp.SupportRecord{
		{Name: "User1", Amount: "10", Time: "2023-01-01"},
	}
	testApp.UpdateSupportRecords("en", records)

	handler := NewVersionHandler(testApp)
	c, w := newVersionTestContext("/api/support?lang=en&page=1&pageSize=10")

	handler.Support(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	
	assert.Contains(t, w.Body.String(), "User1")
}
