package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestClientIPTrustsOnlyConfiguredProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newRouterEngine([]string{"192.0.2.10/32", "127.0.0.1/32"})
	r.GET("/probe", func(c *gin.Context) { c.String(http.StatusOK, c.ClientIP()) })
	for _, test := range []struct{ remote, forwarded, expected string }{
		{"192.0.2.1:1234", "198.51.100.2", "192.0.2.1"},
		{"192.0.2.10:1234", "198.51.100.2", "198.51.100.2"},
		{"192.0.2.11:1234", "198.51.100.2", "192.0.2.11"},
	} {
		request := httptest.NewRequest(http.MethodGet, "/probe", nil)
		request.RemoteAddr = test.remote
		request.Header.Set("X-Forwarded-For", test.forwarded)
		response := httptest.NewRecorder()
		r.ServeHTTP(response, request)
		if response.Body.String() != test.expected {
			t.Fatalf("remote %s: got %s, want %s", test.remote, response.Body.String(), test.expected)
		}
	}
}
