package api_router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestExpvar_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/debug/vars", nil)
	c.Request = req

	Expvar(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	
	// expvar should at least contain "cmdline" and "memstats" by default in Go
	assert.Contains(t, w.Body.String(), "cmdline")
	assert.Contains(t, w.Body.String(), "memstats")
}
