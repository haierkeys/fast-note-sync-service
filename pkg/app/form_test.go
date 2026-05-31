package app

import (
	"bytes"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestRequestParam verifies the correctness of RequestParam extraction across query parameters, post forms, and JSON bodies.
// TestRequestParam 验证 RequestParam 从查询参数、表单及 JSON Body 中提取参数的正确性。
func TestRequestParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Query parameter", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/test?vault=myvault&id=123", nil)

		assert.Equal(t, "myvault", RequestParam(c, "vault"))
		assert.Equal(t, "123", RequestParam(c, "id"))
		assert.Equal(t, "", RequestParam(c, "nonexistent"))
	})

	t.Run("PostForm parameter", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := "vault=formvault&status=true"
		c.Request = httptest.NewRequest("POST", "/api/test", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		assert.Equal(t, "formvault", RequestParam(c, "vault"))
		assert.Equal(t, "true", RequestParam(c, "status"))
	})

	t.Run("JSON Body parameter with stream restoration", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		jsonBody := `{"vault":"jsonvault","id":456,"active":true}`
		c.Request = httptest.NewRequest("POST", "/api/test", bytes.NewBufferString(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// 1. Verify parameter extraction
		assert.Equal(t, "jsonvault", RequestParam(c, "vault"))
		assert.Equal(t, "456", RequestParam(c, "id"))
		assert.Equal(t, "true", RequestParam(c, "active"))

		// 2. Verify that c.Request.Body is fully restored and can be read again downstream
		// 验证 c.Request.Body 已完全复原，下游仍可二次读取
		readBytes, err := io.ReadAll(c.Request.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, jsonBody, string(readBytes))
	})

	t.Run("Empty and invalid content", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/test", nil)
		assert.Equal(t, "", RequestParam(c, "vault"))

		c.Request = httptest.NewRequest("POST", "/api/test", bytes.NewBufferString("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")
		assert.Equal(t, "", RequestParam(c, "vault"))
	})
}
