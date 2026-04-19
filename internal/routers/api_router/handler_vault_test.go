// Package api_router provides HTTP API router handlers
// Package api_router 提供 HTTP API 路由处理器
package api_router

import (
	"encoding/json"
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

func init() {
	// Switch gin to test mode to suppress debug output.
	// 切换到测试模式，抑制 gin 的调试输出。
	gin.SetMode(gin.TestMode)
}

// newVaultTestContext creates a gin.Context suitable for VaultHandler tests.
// newVaultTestContext 创建适合 VaultHandler 测试的 gin.Context。
//
// method: HTTP method (GET, POST, DELETE)
// url: request URL
// body: request body JSON string (empty string for no body)
// uid: authenticated user ID (0 means unauthenticated)
func newVaultTestContext(method, url, body string, uid int64) (*gin.Context, *httptest.ResponseRecorder) {
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

	// Inject authenticated user ID into context (simulates JWT middleware).
	// 将认证用户 ID 注入 context（模拟 JWT 中间件的行为）。
	if uid > 0 {
		c.Set("user_token", &pkgapp.UserEntity{UID: uid})
	}

	return c, w
}

// decodeRes decodes the pkgapp.Res response body.
// decodeRes 解码 pkgapp.Res 响应体。
func decodeRes(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err, "response body should be valid JSON")
	return result
}

// assertResponseCode checks the business code in the JSON response.
// assertResponseCode 检查 JSON 响应中的业务码。
func assertResponseCode(t *testing.T, w *httptest.ResponseRecorder, wantCode int) {
	t.Helper()
	result := decodeRes(t, w)
	gotCode, ok := result["code"].(float64)
	assert.True(t, ok, "response should have a code field")
	assert.Equal(t, float64(wantCode), gotCode, "unexpected business code")
}

// newVaultHandler creates a VaultHandler with the given mock service.
// newVaultHandler 创建使用指定 mock service 的 VaultHandler。
func newVaultHandler(mockSvc *svcmocks.MockVaultService) *VaultHandler {
	testApp := app.NewTestApp(&app.Services{
		VaultService: mockSvc,
	})
	return NewVaultHandler(testApp)
}

// --- List ---

// TestVaultHandler_List_Success verifies successful vault list response.
// TestVaultHandler_List_Success 验证正常获取 Vault 列表的响应。
func TestVaultHandler_List_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockVaultService)
	mockSvc.On("List", mock.Anything, int64(1)).
		Return([]*dto.VaultDTO{
			{ID: 1, Name: "Vault-A"},
			{ID: 2, Name: "Vault-B"},
		}, nil)

	handler := newVaultHandler(mockSvc)
	c, w := newVaultTestContext("GET", "/api/vault", "", 1)
	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestVaultHandler_List_NoUID verifies that missing UID returns auth error.
// TestVaultHandler_List_NoUID 验证未携带 UID 时返回认证错误。
func TestVaultHandler_List_NoUID(t *testing.T) {
	mockSvc := new(svcmocks.MockVaultService)

	handler := newVaultHandler(mockSvc)
	// uid=0 means no authentication injected
	// uid=0 表示未注入认证信息
	c, w := newVaultTestContext("GET", "/api/vault", "", 0)
	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.ErrorNotUserAuthToken.Code())
	mockSvc.AssertExpectations(t) // no service call expected // 期望没有 service 调用
}

// --- Create (CreateOrUpdate with ID=0) ---

// TestVaultHandler_CreateOrUpdate_Create_Success verifies successful vault creation via POST.
// TestVaultHandler_CreateOrUpdate_Create_Success 验证 POST 创建 Vault 成功。
func TestVaultHandler_CreateOrUpdate_Create_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockVaultService)
	mockSvc.On("Create", mock.Anything, int64(1), "TestVault").
		Return(&dto.VaultDTO{ID: 10, Name: "TestVault"}, nil)

	handler := newVaultHandler(mockSvc)
	body := `{"vault": "TestVault", "id": 0}`
	c, w := newVaultTestContext("POST", "/api/vault", body, 1)
	handler.CreateOrUpdate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.SuccessCreate.Code())
	mockSvc.AssertExpectations(t)
}

// TestVaultHandler_CreateOrUpdate_Update_Success verifies successful vault update via POST.
// TestVaultHandler_CreateOrUpdate_Update_Success 验证 POST 更新 Vault 成功。
func TestVaultHandler_CreateOrUpdate_Update_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockVaultService)
	mockSvc.On("Update", mock.Anything, int64(1), int64(5), "UpdatedVault").
		Return(&dto.VaultDTO{ID: 5, Name: "UpdatedVault"}, nil)

	handler := newVaultHandler(mockSvc)
	body := `{"vault": "UpdatedVault", "id": 5}`
	c, w := newVaultTestContext("POST", "/api/vault", body, 1)
	handler.CreateOrUpdate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.SuccessUpdate.Code())
	mockSvc.AssertExpectations(t)
}

// TestVaultHandler_CreateOrUpdate_NoUID verifies auth error when UID is missing.
// TestVaultHandler_CreateOrUpdate_NoUID 验证缺少 UID 时返回认证错误。
func TestVaultHandler_CreateOrUpdate_NoUID(t *testing.T) {
	mockSvc := new(svcmocks.MockVaultService)

	handler := newVaultHandler(mockSvc)
	body := `{"vault": "TestVault"}`
	c, w := newVaultTestContext("POST", "/api/vault", body, 0)
	handler.CreateOrUpdate(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.ErrorNotUserAuthToken.Code())
	mockSvc.AssertExpectations(t)
}

// --- Get ---

// TestVaultHandler_Get_Success verifies successful vault retrieval.
// TestVaultHandler_Get_Success 验证正常获取 Vault 详情。
func TestVaultHandler_Get_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockVaultService)
	mockSvc.On("Get", mock.Anything, int64(1), int64(3)).
		Return(&dto.VaultDTO{ID: 3, Name: "MyVault"}, nil)

	handler := newVaultHandler(mockSvc)
	c, w := newVaultTestContext("GET", "/api/vault/get?id=3", "", 1)
	// Manually set query param since gin test context doesn't parse URL
	// 手动设置查询参数（gin 测试 context 不自动解析 URL query）
	c.Request.URL.RawQuery = "id=3"
	handler.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// --- Delete ---

// TestVaultHandler_Delete_Success verifies successful vault deletion.
// TestVaultHandler_Delete_Success 验证正常删除 Vault。
func TestVaultHandler_Delete_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockVaultService)
	mockSvc.On("Delete", mock.Anything, int64(1), int64(4)).
		Return(nil)

	handler := newVaultHandler(mockSvc)
	c, w := newVaultTestContext("DELETE", "/api/vault?id=4", "", 1)
	c.Request.URL.RawQuery = "id=4"
	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.SuccessDelete.Code())
	mockSvc.AssertExpectations(t)
}

// TestVaultHandler_Delete_NoUID verifies auth error when UID is missing.
// TestVaultHandler_Delete_NoUID 验证缺少 UID 时返回认证错误。
func TestVaultHandler_Delete_NoUID(t *testing.T) {
	mockSvc := new(svcmocks.MockVaultService)

	handler := newVaultHandler(mockSvc)
	c, w := newVaultTestContext("DELETE", "/api/vault?id=4", "", 0)
	c.Request.URL.RawQuery = "id=4"
	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.ErrorNotUserAuthToken.Code())
	mockSvc.AssertExpectations(t)
}
