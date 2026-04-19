package mcp_router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleSSE_HEAD(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create a dummy app container or use a real one if possible
	// Here we just need a handler, but NewMCPHandler requires app and websocket server.
	// For simplicity, we can mock the behavior or just test if the handler sets the header.
	
	r := gin.New()
	
	// We'll manually create a handler state that doesn't crash
	h := &MCPHandler{} 

	r.Match([]string{http.MethodGet, http.MethodHead}, "/sse", h.HandleSSE)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodHead, "/sse", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", w.Header().Get("Connection"))
}
