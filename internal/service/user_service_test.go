// Package service implements the business logic layer
// Package service 实现业务逻辑层
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	domainmocks "github.com/haierkeys/fast-note-sync-service/internal/domain/mocks"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
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

func (m *mockTokenManager) Generate(uid int64, nickname, ip string, tokenID int64, nonce string) (string, error) {
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

type mockUserTokenService struct{}

func (m *mockUserTokenService) Create(ctx context.Context, uid int64, params *dto.TokenIssueRequest) (*dto.TokenCreateResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserTokenService) CreateForLogin(ctx context.Context, uid int64, clientType, ip, userAgent string) (*domain.AuthToken, string, error) {
	return &domain.AuthToken{ID: 1, UID: uid, Status: 1}, "test-token", nil
}

func (m *mockUserTokenService) RotateForLogin(ctx context.Context, uid int64, tokenID int64, ip, userAgent string) (*domain.AuthToken, string, error) {
	if tokenID == 999 {
		return nil, "", errors.New("mock rotate error")
	}
	return &domain.AuthToken{ID: tokenID, UID: uid, Status: 1, ClientType: "webgui"}, "rotated-token", nil
}

func (m *mockUserTokenService) ListByUser(ctx context.Context, uid int64) ([]*dto.TokenResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserTokenService) Update(ctx context.Context, uid int64, tokenID int64, params *dto.TokenUpdateRequest) error {
	return errors.New("not implemented")
}

func (m *mockUserTokenService) Revoke(ctx context.Context, uid int64, tokenID int64) error {
	return errors.New("not implemented")
}

func (m *mockUserTokenService) Rotate(ctx context.Context, uid int64, tokenID int64) (*dto.TokenCreateResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserTokenService) GetActiveToken(ctx context.Context, uid int64, tokenID int64) (*domain.AuthToken, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserTokenService) RecordAccessLog(ctx context.Context, log *domain.AuthTokenLog) error {
	return errors.New("not implemented")
}

func (m *mockUserTokenService) ListLogs(ctx context.Context, uid, tokenID int64, page, pageSize int) ([]*dto.TokenLogResponse, int64, error) {
	return nil, 0, errors.New("not implemented")
}

func (m *mockUserTokenService) UpdateLastUsedAt(ctx context.Context, tokenID int64) error {
	return errors.New("not implemented")
}

func (m *mockUserTokenService) SetSyncHandler(handler func(uid int64, tokenID int64, scope string, kick bool)) {
}

func (m *mockUserTokenService) GetRecentClients(ctx context.Context, uid int64, duration time.Duration) (map[int64][]string, error) {
	return nil, nil
}

func (m *mockUserTokenService) CleanExpired(ctx context.Context, uid int64, issueType int) error {
	return nil
}

// newUserSvc creates a userService with mocked dependencies for testing.
// newUserSvc 创建带 mock 依赖的 userService 用于测试。
func newUserSvc(repo domain.UserRepository, registerEnabled bool) UserService {
	return NewUserService(repo, &mockTokenManager{}, &mockUserTokenService{}, zap.NewNop(), &ServiceConfig{
		User: UserServiceConfig{RegisterIsEnable: registerEnabled, AdminUID: 1},
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
	result, err := svc.Register(context.Background(), params, "127.0.0.1", "WebGui", "test-agent")

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
	}, "127.0.0.1", "WebGui", "test-agent")

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
	}, "127.0.0.1", "WebGui", "test-agent")

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
	}, "127.0.0.1", "WebGui", "test-agent")

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
	}, "127.0.0.1", "WebGui", "test-agent")

	assert.ErrorIs(t, err, code.ErrorUserAlreadyExists)
	mockRepo.AssertExpectations(t)
}

// --- Create ---

// TestUserService_Create_Success verifies successful create user.
func TestUserService_Create_Success(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	params := &dto.UserCreateRequest{
		Email:           "test@example.com",
		Username:        "testuser",
		Password:        "password123",
		ConfirmPassword: "password123",
	}

	// Email and username both not found (available)
	// 邮箱和用户名均未注册（可用）
	mockRepo.On("GetByEmail", mock.Anything, params.Email).
		Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("GetByUsername", mock.Anything, params.Username).
		Return(nil, gorm.ErrRecordNotFound)

	created := &domain.User{UID: 1, Email: params.Email, Username: params.Username}
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
		Return(created, nil)

	svc := newUserSvc(mockRepo, true)
	result, err := svc.Create(context.Background(), params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRepo.AssertExpectations(t)
}

// TestUserService_Create_PasswordMismatch verifies error when passwords do not match.
func TestUserService_Create_PasswordMismatch(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	svc := newUserSvc(mockRepo, true)
	_, err := svc.Create(context.Background(), &dto.UserCreateRequest{
		Email:           "a@b.com",
		Username:        "validuser",
		Password:        "pass1",
		ConfirmPassword: "pass2",
	})

	assert.ErrorIs(t, err, code.ErrorUserPasswordNotMatch)
	mockRepo.AssertExpectations(t)
}

// TestUserService_Create_EmailExists verifies error if a user already exists with this email.
func TestUserService_Create_EmailExists(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	mockRepo.On("GetByEmail", mock.Anything, "dup@example.com").
		Return(&domain.User{UID: 99, Email: "dup@example.com"}, nil)

	svc := newUserSvc(mockRepo, true)
	_, err := svc.Create(context.Background(), &dto.UserCreateRequest{
		Email:           "dup@example.com",
		Username:        "newuser",
		Password:        "password123",
		ConfirmPassword: "password123",
	})

	assert.ErrorIs(t, err, code.ErrorUserEmailAlreadyExists)
	mockRepo.AssertExpectations(t)
}

// TestUserService_Create_UsernameExists verifies error when username is already taken.
func TestUserService_Create_UsernameExists(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	// Email is available, but username is taken
	// 邮箱可用，但用户名已被占用
	mockRepo.On("GetByEmail", mock.Anything, "new@example.com").
		Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("GetByUsername", mock.Anything, "takenuser").
		Return(&domain.User{UID: 99, Username: "takenuser"}, nil)

	svc := newUserSvc(mockRepo, true)
	_, err := svc.Create(context.Background(), &dto.UserCreateRequest{
		Email:           "new@example.com",
		Username:        "takenuser",
		Password:        "password123",
		ConfirmPassword: "password123",
	})

	assert.ErrorIs(t, err, code.ErrorUserAlreadyExists)
	mockRepo.AssertExpectations(t)
}

// --- Update ---

// TestUserService_Update_Success verifies successful update user.
func TestUserService_Update_Success(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	params := &dto.UserUpdateRequest{
		UID:      1,
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	}

	// Updated user exists
	updated := &domain.User{UID: params.UID, Email: params.Email, Username: params.Username}

	mockRepo.On("GetByUID", mock.Anything, params.UID, mock.Anything).
		Return(updated, nil)

	// Email and username both not found (available)
	// 邮箱和用户名均未注册（可用）
	mockRepo.On("GetByEmail", mock.Anything, params.Email).
		Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("GetByUsername", mock.Anything, params.Username).
		Return(nil, gorm.ErrRecordNotFound)

	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).
		Return(updated, nil)

	svc := newUserSvc(mockRepo, true)
	err := svc.Update(context.Background(), params)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestUserService_Update_Wrong_Uid verifies when user not exists.
func TestUserService_Update_Wrong_Uid(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	params := &dto.UserUpdateRequest{
		UID:      1,
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	}

	mockRepo.On("GetByUID", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, gorm.ErrRecordNotFound)

	svc := newUserSvc(mockRepo, true)
	err := svc.Update(context.Background(), params)

	assert.ErrorIs(t, err, code.ErrorUserNotFound)
	mockRepo.AssertExpectations(t)
}

// TestUserService_Update_EmailExists verifies error if a user already exists with this email.
func TestUserService_Update_EmailExists(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	params := &dto.UserUpdateRequest{
		UID:      1,
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	}

	// Updated user exists
	updated := &domain.User{UID: params.UID, Email: params.Email, Username: params.Username}

	mockRepo.On("GetByUID", mock.Anything, params.UID, mock.Anything).
		Return(updated, nil)

	// The email is already in use by an existing user.
	mockRepo.On("GetByEmail", mock.Anything, params.Email).
		Return(&domain.User{UID: 99, Email: params.Email}, nil)

	svc := newUserSvc(mockRepo, true)
	err := svc.Update(context.Background(), params)

	assert.ErrorIs(t, err, code.ErrorUserEmailAlreadyExists)
	mockRepo.AssertExpectations(t)
}

// TestUserService_Update_UsernameExists verifies error if a user already exists with this username.
func TestUserService_Update_UsernameExists(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	params := &dto.UserUpdateRequest{
		UID:      1,
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	}

	// Updated user exists
	updated := &domain.User{UID: params.UID, Email: params.Email, Username: params.Username}

	mockRepo.On("GetByUID", mock.Anything, params.UID, mock.Anything).
		Return(updated, nil)

	mockRepo.On("GetByEmail", mock.Anything, params.Email).
		Return(nil, gorm.ErrRecordNotFound)

	// The username is already in use by an existing user.
	mockRepo.On("GetByUsername", mock.Anything, params.Username).
		Return(&domain.User{UID: 99, Username: params.Username}, nil)

	svc := newUserSvc(mockRepo, true)
	err := svc.Update(context.Background(), params)

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
	}, "127.0.0.1", "WebGui", "test-agent")

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
	}, "127.0.0.1", "WebGui", "test-agent")

	assert.ErrorIs(t, err, code.ErrorUserLoginPasswordFailed)
	mockRepo.AssertExpectations(t)
}

// TestUserService_Login_Rotate_Success verifies token rotation during login.
// TestUserService_Login_Rotate_Success 验证登录时令牌成功轮转。
func TestUserService_Login_Rotate_Success(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)
	hashedPwd := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi" // "password"

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
		TokenID:     123,
	}, "127.0.0.1", "webgui", "test-agent")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "rotated-token", result.Token)
	assert.Equal(t, int64(123), result.TokenID)
	mockRepo.AssertExpectations(t)
}

// TestUserService_Login_Rotate_Fallback verifies token creation fallback when rotation fails.
// TestUserService_Login_Rotate_Fallback 验证轮转失败时降级创建新令牌。
func TestUserService_Login_Rotate_Fallback(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)
	hashedPwd := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi" // "password"

	user := &domain.User{
		UID:      1,
		Email:    "test@example.com",
		Username: "testuser",
		Password: hashedPwd,
	}
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").
		Return(user, nil)

	svc := newUserSvc(mockRepo, true)
	// mockUserTokenService 遇到 TokenID=999 时会模拟报错，并降级
	result, err := svc.Login(context.Background(), &dto.UserLoginRequest{
		Credentials: "test@example.com",
		Password:    "password",
		TokenID:     999,
	}, "127.0.0.1", "webgui", "test-agent")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-token", result.Token) // Fallback creates new token returning "test-token"
	assert.Equal(t, int64(1), result.TokenID)   // Mock created token has ID: 1
	mockRepo.AssertExpectations(t)
}

// --- GetInfo ---

// TestUserService_GetInfo_Success verifies successful user info retrieval.
// TestUserService_GetInfo_Success 验证正常获取用户信息。
func TestUserService_GetInfo_Success(t *testing.T) {
	mockRepo := new(domainmocks.MockUserRepository)

	user := &domain.User{UID: 1, Email: "a@b.com", Username: "user1"}
	mockRepo.On("GetByUID", mock.Anything, int64(1), mock.Anything).
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

	mockRepo.On("GetByUID", mock.Anything, int64(99), mock.Anything).
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
	mockRepo.On("GetByUID", mock.Anything, int64(1), mock.Anything).Return(user, nil)
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

// --- GetList ---

// TestUserService_GetList verifies returning users list with pagination
func TestUserService_GetList(t *testing.T) {
	// Domain data
	mockUsers := []*domain.User{
		{UID: 1, Email: "user1@example.com", Username: "user1"},
		{UID: 2, Email: "user2@example.com", Username: "user2"},
	}

	// DTOs
	expectedDTOs := []*dto.UserDTO{
		{UID: 1, Email: "user1@example.com", Username: "user1"},
		{UID: 2, Email: "user2@example.com", Username: "user2"},
	}

	dbError := errors.New("database connection failed")

	// cases
	tests := []struct {
		name           string
		pager          *app.Pager
		mockSetup      func(m *domainmocks.MockUserRepository)
		expectedResult []*dto.UserDTO
		expectedTotal  int64
		expectedErr    error
	}{
		{
			name: "Success - First page",
			pager: &app.Pager{
				Page:     1,
				PageSize: 10,
			},
			mockSetup: func(m *domainmocks.MockUserRepository) {
				// expect: offset = 0, limit = 10
				m.On("GetList", mock.Anything, 0, 10).
					Return(mockUsers, 25, nil) // total in db = 25
			},
			expectedResult: expectedDTOs,
			expectedTotal:  25,
			expectedErr:    nil,
		},
		{
			name: "Success - second page",
			pager: &app.Pager{
				Page:     2,
				PageSize: 10,
			},
			mockSetup: func(m *domainmocks.MockUserRepository) {
				// expect: offset = 10, limit = 10
				m.On("GetList", mock.Anything, 10, 10).
					Return(mockUsers, 25, nil)
			},
			expectedResult: expectedDTOs,
			expectedTotal:  25,
			expectedErr:    nil,
		},
		{
			name: "Failure - error db",
			pager: &app.Pager{
				Page:     1,
				PageSize: 10,
			},
			mockSetup: func(m *domainmocks.MockUserRepository) {
				m.On("GetList", mock.Anything, 0, 10).
					Return(nil, 0, dbError)
			},
			expectedResult: nil,
			expectedTotal:  0,
			expectedErr:    code.ErrorDBQuery,
		},
	}

	// run all test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(domainmocks.MockUserRepository)
			tt.mockSetup(mockRepo)
			svc := newUserSvc(mockRepo, true)

			result, total, err := svc.GetList(context.Background(), tt.pager)

			// check error
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			// check total and result
			assert.Equal(t, tt.expectedTotal, total)
			assert.Equal(t, tt.expectedResult, result)

			mockRepo.AssertExpectations(t)
		})
	}
}

// --- IsRegisterEnabled ---

// TestUserService_IsRegisterEnabled verifies the logic for allowing/disallowing registration.
// TestUserService_IsRegisterEnabled 验证是否允许注册的逻辑。
func TestUserService_IsRegisterEnabled(t *testing.T) {
	t.Run("ConfigDisabled", func(t *testing.T) {
		mockRepo := new(domainmocks.MockUserRepository)
		svc := NewUserService(mockRepo, &mockTokenManager{}, &mockUserTokenService{}, zap.NewNop(), &ServiceConfig{
			User: UserServiceConfig{RegisterIsEnable: false, AdminUID: 0},
		})
		assert.False(t, svc.IsRegisterEnabled(context.Background()))
	})

	t.Run("AdminUIDSet_Enabled", func(t *testing.T) {
		mockRepo := new(domainmocks.MockUserRepository)
		svc := NewUserService(mockRepo, &mockTokenManager{}, &mockUserTokenService{}, zap.NewNop(), &ServiceConfig{
			User: UserServiceConfig{RegisterIsEnable: true, AdminUID: 1},
		})
		assert.True(t, svc.IsRegisterEnabled(context.Background()))
	})

	t.Run("AdminUIDZero_NoUsers", func(t *testing.T) {
		mockRepo := new(domainmocks.MockUserRepository)
		mockRepo.On("GetAllUIDs", mock.Anything).Return([]int64{}, nil)
		svc := NewUserService(mockRepo, &mockTokenManager{}, &mockUserTokenService{}, zap.NewNop(), &ServiceConfig{
			User: UserServiceConfig{RegisterIsEnable: true, AdminUID: 0},
		})
		assert.True(t, svc.IsRegisterEnabled(context.Background()))
		mockRepo.AssertExpectations(t)
	})

	t.Run("AdminUIDZero_WithUsers", func(t *testing.T) {
		mockRepo := new(domainmocks.MockUserRepository)
		mockRepo.On("GetAllUIDs", mock.Anything).Return([]int64{1}, nil)
		svc := NewUserService(mockRepo, &mockTokenManager{}, &mockUserTokenService{}, zap.NewNop(), &ServiceConfig{
			User: UserServiceConfig{RegisterIsEnable: true, AdminUID: 0},
		})
		assert.False(t, svc.IsRegisterEnabled(context.Background()))
		mockRepo.AssertExpectations(t)
	})
}
