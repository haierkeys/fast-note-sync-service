// Package mocks provides testify/mock implementations for service layer interfaces.
// Package mocks 提供 service 层接口的 testify/mock 实现，供路由层测试使用。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a testify mock for service.UserService.
// MockUserService 是 service.UserService 的 testify mock 实现。
type MockUserService struct {
	mock.Mock
}

// Register handles user registration.
// Register 处理用户注册。
func (m *MockUserService) Register(ctx context.Context, params *dto.UserCreateRequest, clientIP string, clientType string, userAgent string) (*dto.UserDTO, error) {
	args := m.Called(ctx, params, clientIP, clientType, userAgent)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserDTO), args.Error(1)
}

// Create user
func (m *MockUserService) Create(ctx context.Context, params *dto.UserCreateRequest) (*dto.UserDTO, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserDTO), args.Error(1)
}

// Update user
func (m *MockUserService) Update(ctx context.Context, params *dto.UserUpdateRequest) error {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return args.Error(1)
	}
	return args.Error(1)
}

// Login handles user login.
// Login 处理用户登录。
func (m *MockUserService) Login(ctx context.Context, params *dto.UserLoginRequest, clientIP string, clientType string, userAgent string) (*dto.UserDTO, error) {
	args := m.Called(ctx, params, clientIP, clientType, userAgent)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserDTO), args.Error(1)
}

// ChangePassword changes user password.
// ChangePassword 修改用户密码。
func (m *MockUserService) ChangePassword(ctx context.Context, uid int64, params *dto.UserChangePasswordRequest) error {
	args := m.Called(ctx, uid, params)
	return args.Error(0)
}

// GetInfo retrieves user information.
// GetInfo 获取用户信息。
func (m *MockUserService) GetInfo(ctx context.Context, uid int64) (*dto.UserDTO, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserDTO), args.Error(1)
}

// GetAllUIDs retrieves all user UIDs.
// GetAllUIDs 获取所有用户的 UID 列表。
func (m *MockUserService) GetAllUIDs(ctx context.Context) ([]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

// GetAll retrieves all users info
func (m *MockUserService) GetAll(ctx context.Context) ([]*dto.UserDTO, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.UserDTO), args.Error(1)
}

// GetList retrieves users with pagination // GetList 分页获取用户列表
func (m *MockUserService) GetList(ctx context.Context, pager *pkgapp.Pager) ([]*dto.UserDTO, int64, error) {
	args := m.Called(ctx, pager)
	if args.Get(0) == nil {
		return nil, int64(args.Int(1)), args.Error(2)
	}
	return args.Get(0).([]*dto.UserDTO), int64(args.Int(1)), args.Error(2)
}

// IsRegisterEnabled checks if registration is allowed.
// IsRegisterEnabled 检查是否允许注册。
func (m *MockUserService) IsRegisterEnabled(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

// Compile-time check: MockUserService must implement service.UserService.
// 编译时检查：MockUserService 必须实现 service.UserService 接口。
var _ service.UserService = (*MockUserService)(nil)
