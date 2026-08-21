package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

func TestNoFoundReturnsHTTPNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.NoRoute(NoFound())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/missing-route", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404, body=%s", w.Code, w.Body.String())
	}

	var body app.Res
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if body.Code != code.ErrorNotFoundAPI.Code() {
		t.Fatalf("code = %d, want %d", body.Code, code.ErrorNotFoundAPI.Code())
	}
	if body.Status {
		t.Fatal("status = true, want false")
	}
}
