// Package api_router provides HTTP API router handlers
// Package api_router 提供 HTTP API 路由处理器
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

// newUserTestContext creates a gin.Context suitable for UserHandler tests.
// newUserTestContext 创建适合 UserHandler 测试的 gin.Context。
func newUserTestContext(method, url, body string, uid int64) (*gin.Context, *httptest.ResponseRecorder) {
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

	// Inject authenticated user into context when UID is provided.
	// 当 UID 非零时，将认证用户注入 context。
	if uid > 0 {
		c.Set("user_token", &pkgapp.UserEntity{UID: uid})
	}

	return c, w
}

// newUserHandler creates a UserHandler with the given mock service.
// newUserHandler 创建使用指定 mock service 的 UserHandler。
func newUserHandler(mockSvc *svcmocks.MockUserService) *UserHandler {
	testApp := app.NewTestApp(&app.Services{
		UserService: mockSvc,
	})
	return NewUserHandler(testApp)
}

// --- Register ---

// TestUserHandler_Register_Success verifies successful registration response.
// TestUserHandler_Register_Success 验证注册成功时的响应。
func TestUserHandler_Register_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockUserService)
	mockSvc.On("Register", mock.Anything, mock.AnythingOfType("*dto.UserCreateRequest")).
		Return(&dto.UserDTO{UID: 1, Username: "testuser", Token: "test-token"}, nil)

	handler := newUserHandler(mockSvc)
	body := `{"email":"test@example.com","username":"testuser","password":"pass123","confirmPassword":"pass123"}`
	c, w := newUserTestContext("POST", "/api/user/register", body, 0)
	handler.Register(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestUserHandler_Register_ServiceError verifies that service errors are propagated.
// TestUserHandler_Register_ServiceError 验证 service 错误被正确传递到响应。
func TestUserHandler_Register_ServiceError(t *testing.T) {
	mockSvc := new(svcmocks.MockUserService)
	mockSvc.On("Register", mock.Anything, mock.AnythingOfType("*dto.UserCreateRequest")).
		Return(nil, code.ErrorUserRegisterIsDisable)

	handler := newUserHandler(mockSvc)
	body := `{"email":"test@example.com","username":"testuser","password":"pass123","confirmPassword":"pass123"}`
	c, w := newUserTestContext("POST", "/api/user/register", body, 0)
	handler.Register(c)

	assert.Equal(t, http.StatusOK, w.Code)
	// Service-returned error code should be reflected in response
	// service 返回的错误码应反映在响应中
	assertResponseCode(t, w, code.ErrorUserRegisterIsDisable.Code())
	mockSvc.AssertExpectations(t)
}

// --- Login ---

// TestUserHandler_Login_Success verifies successful login response.
// TestUserHandler_Login_Success 验证登录成功时的响应。
func TestUserHandler_Login_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockUserService)
	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("*dto.UserLoginRequest"), mock.AnythingOfType("string")).
		Return(&dto.UserDTO{UID: 1, Token: "test-token"}, nil)

	handler := newUserHandler(mockSvc)
	body := `{"credentials":"test@example.com","password":"pass123"}`
	c, w := newUserTestContext("POST", "/api/user/login", body, 0)
	handler.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestUserHandler_Login_Failure verifies error response when credentials are wrong.
// TestUserHandler_Login_Failure 验证凭证错误时返回错误响应。
func TestUserHandler_Login_Failure(t *testing.T) {
	mockSvc := new(svcmocks.MockUserService)
	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("*dto.UserLoginRequest"), mock.AnythingOfType("string")).
		Return(nil, code.ErrorUserLoginPasswordFailed)

	handler := newUserHandler(mockSvc)
	body := `{"credentials":"test@example.com","password":"wrong"}`
	c, w := newUserTestContext("POST", "/api/user/login", body, 0)
	handler.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.ErrorUserLoginPasswordFailed.Code())
	mockSvc.AssertExpectations(t)
}

// --- UserInfo ---

// TestUserHandler_UserInfo_Success verifies successful user info retrieval.
// TestUserHandler_UserInfo_Success 验证正常获取用户信息的响应。
func TestUserHandler_UserInfo_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockUserService)
	mockSvc.On("GetInfo", mock.Anything, int64(1)).
		Return(&dto.UserDTO{UID: 1, Email: "a@b.com", Username: "user1"}, nil)

	handler := newUserHandler(mockSvc)
	c, w := newUserTestContext("GET", "/api/user/info", "", 1)
	handler.UserInfo(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.Success.Code())
	mockSvc.AssertExpectations(t)
}

// TestUserHandler_UserInfo_NoUID verifies auth error when UID is missing.
// TestUserHandler_UserInfo_NoUID 验证缺少 UID 时返回认证错误。
func TestUserHandler_UserInfo_NoUID(t *testing.T) {
	mockSvc := new(svcmocks.MockUserService)

	handler := newUserHandler(mockSvc)
	c, w := newUserTestContext("GET", "/api/user/info", "", 0)
	handler.UserInfo(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.ErrorNotUserAuthToken.Code())
	mockSvc.AssertExpectations(t) // no service call expected // 期望没有 service 调用
}

// --- ChangePassword ---

// TestUserHandler_ChangePassword_Success verifies successful password change response.
// TestUserHandler_ChangePassword_Success 验证修改密码成功时的响应。
func TestUserHandler_ChangePassword_Success(t *testing.T) {
	mockSvc := new(svcmocks.MockUserService)
	mockSvc.On("ChangePassword", mock.Anything, int64(1), mock.AnythingOfType("*dto.UserChangePasswordRequest")).
		Return(nil)

	handler := newUserHandler(mockSvc)
	body := `{"oldPassword":"old123","password":"new123","confirmPassword":"new123"}`
	c, w := newUserTestContext("POST", "/api/user/change_password", body, 1)
	handler.UserChangePassword(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.SuccessPasswordUpdate.Code())
	mockSvc.AssertExpectations(t)
}

// TestUserHandler_ChangePassword_NoUID verifies auth error when UID is missing.
// TestUserHandler_ChangePassword_NoUID 验证缺少 UID 时返回认证错误。
func TestUserHandler_ChangePassword_NoUID(t *testing.T) {
	mockSvc := new(svcmocks.MockUserService)

	handler := newUserHandler(mockSvc)
	body := `{"oldPassword":"old123","password":"new123","confirmPassword":"new123"}`
	c, w := newUserTestContext("POST", "/api/user/change_password", body, 0)
	handler.UserChangePassword(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assertResponseCode(t, w, code.ErrorNotUserAuthToken.Code())
	mockSvc.AssertExpectations(t)
}
