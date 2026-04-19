// Package mocks provides testify/mock implementations for domain Repository interfaces.
// Package mocks 提供 domain Repository 接口的 testify/mock 实现。
package mocks

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a testify mock for domain.UserRepository.
// MockUserRepository 是 domain.UserRepository 的 testify mock 实现。
type MockUserRepository struct {
	mock.Mock
}

// GetByUID retrieves a user by UID.
// GetByUID 根据 UID 获取用户。
func (m *MockUserRepository) GetByUID(ctx context.Context, uid int64) (*domain.User, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// GetByEmail retrieves a user by email.
// GetByEmail 根据邮箱获取用户。
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// GetByUsername retrieves a user by username.
// GetByUsername 根据用户名获取用户。
func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// Create creates a new user.
// Create 创建新用户。
func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// UpdatePassword updates the user's password.
// UpdatePassword 更新用户密码。
func (m *MockUserRepository) UpdatePassword(ctx context.Context, password string, uid int64) error {
	args := m.Called(ctx, password, uid)
	return args.Error(0)
}

// GetAllUIDs retrieves all user UIDs.
// GetAllUIDs 获取所有用户的 UID 列表。
func (m *MockUserRepository) GetAllUIDs(ctx context.Context) ([]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

// Compile-time check: MockUserRepository must implement domain.UserRepository.
// 编译时检查：MockUserRepository 必须实现 domain.UserRepository 接口。
var _ domain.UserRepository = (*MockUserRepository)(nil)
