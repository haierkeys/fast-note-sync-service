// Package service implements the business logic layer
// Package service 实现业务逻辑层
package service

import (
	"context"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	domainmocks "github.com/haierkeys/fast-note-sync-service/internal/domain/mocks"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// mockTokenManager is a minimal TokenManager stub for UserService tests.
// mockTokenManager 是用于 UserService 测试的最小 TokenManager stub。
type mockTokenManager struct{}

func (m *mockTokenManager) Generate(uid int64, nickname, ip string) (string, error) {
	return "test-token", nil
}
func (m *mockTokenManager) Parse(token string) (*pkgapp.UserEntity, error) {
	return &pkgapp.UserEntity{UID: 1}, nil
}
func (m *mockTokenManager) ShareGenerate(shareID int64, uid int64, resources map[string][]string) (string, error) {
	return "share-token", nil
}
func (m *mockTokenManager) ShareParse(token string) (*pkgapp.ShareEntity, error) {
	return nil, nil
}
func (m *mockTokenManager) Validate(token string) error { return nil }
func (m *mockTokenManager) GetSecretKey() string        { return "test-key" }

// newUserSvc creates a userService with mocked dependencies for testing.
// newUserSvc 创建带 mock 依赖的 userService 用于测试。
func newUserSvc(repo domain.UserRepository, registerEnabled bool) UserService {
	return NewUserService(repo, &mockTokenManager{}, zap.NewNop(), &ServiceConfig{
		User: UserServiceConfig{RegisterIsEnable: registerEnabled},
	})
}

// --- Register ---

// TestUserService_Register_Success verifies successful user registration.
// TestUserService_Register_Success 验证正常用户注册流程。
func TestUserService_Register_Success(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	params := &dto.UserCreateRequest{
		Email:           "test@example.com",
		Username:        "testuser",
		Password:        "password123",
		ConfirmPassword: "password123",
	}

	// Email and username both not found (available)
	// 邮箱和用户名均未注册（可用）
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").
		Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("GetByUsername", mock.Anything, "testuser").
		Return(nil, gorm.ErrRecordNotFound)

	created := &domain.User{UID: 1, Email: "test@example.com", Username: "testuser"}
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
		Return(created, nil)

	svc := newUserSvc(mockRepo, true)
	result, err := svc.Register(context.Background(), params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-token", result.Token)
	mockRepo.AssertExpectations(t)
}

// TestUserService_Register_Disabled verifies error when registration is disabled.
// TestUserService_Register_Disabled 验证注册功能关闭时返回错误。
func TestUserService_Register_Disabled(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	svc := newUserSvc(mockRepo, false)
	_, err := svc.Register(context.Background(), &dto.UserCreateRequest{
		Email:           "a@b.com",
		Username:        "user1",
		Password:        "pass",
		ConfirmPassword: "pass",
	})

	assert.ErrorIs(t, err, code.ErrorUserRegisterIsDisable)
	mockRepo.AssertExpectations(t) // no repo calls expected // 期望没有 Repository 调用
}

// TestUserService_Register_PasswordMismatch verifies error when passwords do not match.
// TestUserService_Register_PasswordMismatch 验证密码不一致时返回错误。
func TestUserService_Register_PasswordMismatch(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	svc := newUserSvc(mockRepo, true)
	_, err := svc.Register(context.Background(), &dto.UserCreateRequest{
		Email:           "a@b.com",
		Username:        "validuser",
		Password:        "pass1",
		ConfirmPassword: "pass2",
	})

	assert.ErrorIs(t, err, code.ErrorUserPasswordNotMatch)
	mockRepo.AssertExpectations(t)
}

// TestUserService_Register_EmailExists verifies error when email is already registered.
// TestUserService_Register_EmailExists 验证邮箱已存在时返回错误。
func TestUserService_Register_EmailExists(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	mockRepo.On("GetByEmail", mock.Anything, "dup@example.com").
		Return(&domain.User{UID: 99, Email: "dup@example.com"}, nil)

	svc := newUserSvc(mockRepo, true)
	_, err := svc.Register(context.Background(), &dto.UserCreateRequest{
		Email:           "dup@example.com",
		Username:        "newuser",
		Password:        "password123",
		ConfirmPassword: "password123",
	})

	assert.ErrorIs(t, err, code.ErrorUserEmailAlreadyExists)
	mockRepo.AssertExpectations(t)
}

// TestUserService_Register_UsernameExists verifies error when username is already taken.
// TestUserService_Register_UsernameExists 验证用户名已存在时返回错误。
func TestUserService_Register_UsernameExists(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	// Email is available, but username is taken
	// 邮箱可用，但用户名已被占用
	mockRepo.On("GetByEmail", mock.Anything, "new@example.com").
		Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("GetByUsername", mock.Anything, "takenuser").
		Return(&domain.User{UID: 99, Username: "takenuser"}, nil)

	svc := newUserSvc(mockRepo, true)
	_, err := svc.Register(context.Background(), &dto.UserCreateRequest{
		Email:           "new@example.com",
		Username:        "takenuser",
		Password:        "password123",
		ConfirmPassword: "password123",
	})

	assert.ErrorIs(t, err, code.ErrorUserAlreadyExists)
	mockRepo.AssertExpectations(t)
}

// --- Login ---

// TestUserService_Login_ByEmail_Success verifies successful login using email.
// TestUserService_Login_ByEmail_Success 验证通过邮箱登录成功。
func TestUserService_Login_ByEmail_Success(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	// Pre-hashed password for "password123"
	// "password123" 的预计算哈希密码（使用真实 bcrypt hash用于测试）
	// We use a real hash here to make util.CheckPasswordHash pass
	// 此处使用真实 hash 以通过 util.CheckPasswordHash 验证
	hashedPwd := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi" // "password" from bcrypt

	user := &domain.User{
		UID:      1,
		Email:    "test@example.com",
		Username: "testuser",
		Password: hashedPwd,
	}
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").
		Return(user, nil)

	svc := newUserSvc(mockRepo, true)
	result, err := svc.Login(context.Background(), &dto.UserLoginRequest{
		Credentials: "test@example.com",
		Password:    "password",
	}, "127.0.0.1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-token", result.Token)
	mockRepo.AssertExpectations(t)
}

// TestUserService_Login_WrongPassword verifies error when password is incorrect.
// TestUserService_Login_WrongPassword 验证密码错误时返回错误。
func TestUserService_Login_WrongPassword(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	user := &domain.User{
		UID:      1,
		Email:    "test@example.com",
		Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
	}
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").
		Return(user, nil)

	svc := newUserSvc(mockRepo, true)
	_, err := svc.Login(context.Background(), &dto.UserLoginRequest{
		Credentials: "test@example.com",
		Password:    "wrong-password",
	}, "127.0.0.1")

	assert.ErrorIs(t, err, code.ErrorUserLoginPasswordFailed)
	mockRepo.AssertExpectations(t)
}

// --- GetInfo ---

// TestUserService_GetInfo_Success verifies successful user info retrieval.
// TestUserService_GetInfo_Success 验证正常获取用户信息。
func TestUserService_GetInfo_Success(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	user := &domain.User{UID: 1, Email: "a@b.com", Username: "user1"}
	mockRepo.On("GetByUID", mock.Anything, int64(1)).
		Return(user, nil)

	svc := newUserSvc(mockRepo, true)
	result, err := svc.GetInfo(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.UID)
	mockRepo.AssertExpectations(t)
}

// TestUserService_GetInfo_NotFound verifies nil return when user does not exist.
// TestUserService_GetInfo_NotFound 验证用户不存在时返回 nil。
func TestUserService_GetInfo_NotFound(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	mockRepo.On("GetByUID", mock.Anything, int64(99)).
		Return(nil, gorm.ErrRecordNotFound)

	svc := newUserSvc(mockRepo, true)
	result, err := svc.GetInfo(context.Background(), 99)

	assert.NoError(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

// --- ChangePassword ---

// TestUserService_ChangePassword_Success verifies successful password change.
// TestUserService_ChangePassword_Success 验证正常修改密码流程。
func TestUserService_ChangePassword_Success(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	user := &domain.User{
		UID:      1,
		Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
	}
	mockRepo.On("GetByUID", mock.Anything, int64(1)).Return(user, nil)
	mockRepo.On("UpdatePassword", mock.Anything, mock.AnythingOfType("string"), int64(1)).Return(nil)

	svc := newUserSvc(mockRepo, true)
	err := svc.ChangePassword(context.Background(), 1, &dto.UserChangePasswordRequest{
		OldPassword:     "password",
		Password:        "newpass123",
		ConfirmPassword: "newpass123",
	})

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
