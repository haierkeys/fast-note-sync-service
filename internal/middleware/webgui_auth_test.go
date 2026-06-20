package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildWebGUITestRouter creates a test router with RequireWebGUI middleware,
// optionally pre-seeding token context values to simulate authenticated routes.
// buildWebGUITestRouter 创建一个带 RequireWebGUI 中间件的测试路由，
// 可选预置 Token 上下文值以模拟已认证路由
func buildWebGUITestRouter(issueType *int, clientType *string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Seed middleware: simulates UserAuthTokenWithConfig injecting token attributes
	// 种子中间件：模拟 UserAuthTokenWithConfig 注入 Token 属性
	if issueType != nil {
		router.Use(func(c *gin.Context) {
			c.Set("token_issue_type", *issueType)
			if clientType != nil {
				c.Set("token_client_type", *clientType)
			} else {
				c.Set("token_client_type", "")
			}
			c.Next()
		})
	}

	router.Use(RequireWebGUI())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": code.Success.Code(), "status": true})
	})
	return router
}

// doWebGUIRequest sends a GET /test request with optional x-client header and returns the parsed response.
// doWebGUIRequest 发送带可选 x-client 请求头的 GET /test 请求，返回解析后的响应
func doWebGUIRequest(t *testing.T, router *gin.Engine, xClient string) app.Res {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	if xClient != "" {
		req.Header.Set("x-client", xClient)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var res app.Res
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	return res
}

// ------- No-auth route tests (no token context) -------

// TestRequireWebGUI_NoAuthRoute_WebGUIHeader_Allowed verifies that unauthenticated routes
// are allowed when x-client: webgui header is present.
// 验证免认证路由在携带 x-client: webgui 时可正常通过
func TestRequireWebGUI_NoAuthRoute_WebGUIHeader_Allowed(t *testing.T) {
	router := buildWebGUITestRouter(nil, nil)
	res := doWebGUIRequest(t, router, "webgui")
	assert.Equal(t, code.Success.Code(), res.Code)
}

// TestRequireWebGUI_NoAuthRoute_NonWebGUIHeader_Blocked verifies that non-webgui client headers are rejected.
// 验证非 webgui 客户端头被拦截
func TestRequireWebGUI_NoAuthRoute_NonWebGUIHeader_Blocked(t *testing.T) {
	router := buildWebGUITestRouter(nil, nil)
	res := doWebGUIRequest(t, router, "obsidian")
	assert.Equal(t, code.ErrorAuthTokenClientRestricted.Code(), res.Code)
}

// TestRequireWebGUI_NoAuthRoute_NoHeader_Blocked verifies that missing x-client header is rejected.
// 验证缺少 x-client 请求头时被拦截
func TestRequireWebGUI_NoAuthRoute_NoHeader_Blocked(t *testing.T) {
	router := buildWebGUITestRouter(nil, nil)
	res := doWebGUIRequest(t, router, "")
	assert.Equal(t, code.ErrorAuthTokenClientRestricted.Code(), res.Code)
}

// ------- Authenticated route tests (token context present) -------

// TestRequireWebGUI_AuthRoute_LoginToken_WebGUIClient_Allowed verifies the happy path:
// Login Token + ClientType=webgui + x-client: webgui header all pass.
// 验证正常使用场景：Login Token + webgui ClientType + webgui Header 全部通过
func TestRequireWebGUI_AuthRoute_LoginToken_WebGUIClient_Allowed(t *testing.T) {
	issueType := 1
	clientType := "webgui"
	router := buildWebGUITestRouter(&issueType, &clientType)
	res := doWebGUIRequest(t, router, "webgui")
	assert.Equal(t, code.Success.Code(), res.Code)
}

// TestRequireWebGUI_AuthRoute_ManualToken_Blocked verifies that manual API tokens (IssueType=2)
// are rejected even with correct header and clientType.
// 验证手动 API 令牌（IssueType=2）即使 Header 和 ClientType 都正确也被拦截
func TestRequireWebGUI_AuthRoute_ManualToken_Blocked(t *testing.T) {
	issueType := 2
	clientType := "webgui"
	router := buildWebGUITestRouter(&issueType, &clientType)
	res := doWebGUIRequest(t, router, "webgui")
	assert.Equal(t, code.ErrorAuthTokenClientRestricted.Code(), res.Code)
	assert.Contains(t, res.Details, "Manual API tokens")
}

// TestRequireWebGUI_AuthRoute_LoginToken_NonWebGUIClientType_Blocked verifies that a Login Token
// with a non-webgui ClientType is rejected even with a correct x-client header.
// 验证 Login Token 但 ClientType 不是 webgui 时被拦截（即使 Header 正确）
func TestRequireWebGUI_AuthRoute_LoginToken_NonWebGUIClientType_Blocked(t *testing.T) {
	issueType := 1
	clientType := "obsidian"
	router := buildWebGUITestRouter(&issueType, &clientType)
	res := doWebGUIRequest(t, router, "webgui")
	assert.Equal(t, code.ErrorAuthTokenClientRestricted.Code(), res.Code)
	assert.Contains(t, res.Details, "client type does not match")
}

// TestRequireWebGUI_AuthRoute_SpoofedHeader_ManualToken_Blocked is the CORE security test.
// It verifies that spoofing x-client: webgui with a manual API token (IssueType=2) is blocked.
// This is the exact attack vector described in Issue #378.
// 核心安全测试：验证伪造 x-client: webgui + 手动 API 令牌（IssueType=2）被拦截。
// 这正是 Issue #378 描述的攻击向量
func TestRequireWebGUI_AuthRoute_SpoofedHeader_ManualToken_Blocked(t *testing.T) {
	issueType := 2
	clientType := "webgui" // Token was created as webgui-type // Token 签发时绑定了 webgui 类型
	router := buildWebGUITestRouter(&issueType, &clientType)
	// Attacker spoofs x-client: webgui header // 攻击者伪造 x-client: webgui 请求头
	res := doWebGUIRequest(t, router, "webgui")
	assert.Equal(t, code.ErrorAuthTokenClientRestricted.Code(), res.Code)
	assert.Contains(t, res.Details, "Manual API tokens")
}

// TestRequireWebGUI_AuthRoute_CaseInsensitiveWebGUI_Allowed verifies case-insensitive x-client matching.
// 验证 x-client Header 大小写不敏感匹配
func TestRequireWebGUI_AuthRoute_CaseInsensitiveWebGUI_Allowed(t *testing.T) {
	issueType := 1
	clientType := "webgui"
	router := buildWebGUITestRouter(&issueType, &clientType)
	res := doWebGUIRequest(t, router, "WebGui")
	assert.Equal(t, code.Success.Code(), res.Code)
}
