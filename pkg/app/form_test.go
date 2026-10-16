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

// TestBindAndValidRejectsUnparseableBody guards kvpac-suxu2: a request that carries a
// body whose Content-Type the binder cannot parse (text/plain, text/html, or none) used
// to pass validation by binding from the URL query alone, letting handlers persist empty
// content while returning Success (silent data loss). Such requests must be rejected.
// TestBindAndValidRejectsUnparseableBody 守护 kvpac-suxu2：带 body 但 Content-Type
// 不可解析（text/plain、text/html 或未声明）的请求，过去会仅凭 URL query 通过校验，
// 使处理函数把内容写成空值却返回 Success（静默数据丢失）。此类请求必须被拒绝。
func TestBindAndValidRejectsUnparseableBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type noteReq struct {
		Vault   string `json:"vault" form:"vault" binding:"required"`
		Path    string `json:"path" form:"path" binding:"required"`
		Content string `json:"content" form:"content"`
	}

	newCtx := func(method, target, body, contentType string) *gin.Context {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		r := httptest.NewRequest(method, target, nil)
		if body != "" {
			r = httptest.NewRequest(method, target, strings.NewReader(body))
		}
		if contentType != "" {
			r.Header.Set("Content-Type", contentType)
		}
		c.Request = r
		return c
	}

	t.Run("JSON body still binds and validates", func(t *testing.T) {
		c := newCtx("POST", "/api/note", `{"vault":"v","path":"p","content":"hello"}`, "application/json")
		req := &noteReq{}
		ok, errs := BindAndValid(c, req)
		assert.True(t, ok)
		assert.Empty(t, errs)
		assert.Equal(t, "hello", req.Content)
	})

	t.Run("form-urlencoded body still binds", func(t *testing.T) {
		c := newCtx("POST", "/api/note", "vault=v&path=p&content=hello", "application/x-www-form-urlencoded")
		req := &noteReq{}
		ok, _ := BindAndValid(c, req)
		assert.True(t, ok)
		assert.Equal(t, "hello", req.Content)
	})

	t.Run("text/plain raw body is rejected", func(t *testing.T) {
		c := newCtx("POST", "/api/note?vault=v&path=p", "RAW CONTENT", "text/plain; charset=utf-8")
		req := &noteReq{}
		ok, errs := BindAndValid(c, req)
		assert.False(t, ok, "text/plain body must not pass validation")
		assert.NotEmpty(t, errs)
		assert.Equal(t, "Content-Type", errs[0].Field())
	})

	t.Run("text/html body is rejected", func(t *testing.T) {
		c := newCtx("POST", "/api/note?vault=v&path=p", "<p>x</p>", "text/html")
		req := &noteReq{}
		ok, _ := BindAndValid(c, req)
		assert.False(t, ok)
	})

	t.Run("undeclared Content-Type with body is rejected", func(t *testing.T) {
		c := newCtx("POST", "/api/note?vault=v&path=p", "RAW CONTENT", "")
		req := &noteReq{}
		ok, _ := BindAndValid(c, req)
		assert.False(t, ok, "body without Content-Type must not pass validation")
	})

	t.Run("bodyless request keeps previous behaviour", func(t *testing.T) {
		c := newCtx("POST", "/api/vault?vault=v", "", "")
		req := &noteReq{Vault: "v", Path: "p"}
		ok, _ := BindAndValid(c, req)
		assert.True(t, ok, "query-only requests must keep working")
	})
}
