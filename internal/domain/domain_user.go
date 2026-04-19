package domain

import (
	"context"
	"time"
)


// User 用户领域模型
type User struct {
	UID       int64
	Email     string
	Username  string
	Password  string
	Salt      string
	Token     string
	Avatar    string
	IsDeleted bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

// HasEmail 判断用户是否有邮箱
func (u *User) HasEmail() bool {
	return u.Email != ""
}

// HasAvatar 判断用户是否有头像
func (u *User) HasAvatar() bool {
	return u.Avatar != ""
}

// IsActive 判断用户是否活跃（未删除）
func (u *User) IsActive() bool {
	return !u.IsDeleted
}

// UserRepository 用户仓储接口
type UserRepository interface {
	// GetByUID 根据UID获取用户
	GetByUID(ctx context.Context, uid int64) (*User, error)

	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*User, error)

	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*User, error)

	// Create 创建用户
	Create(ctx context.Context, user *User) (*User, error)

	// UpdatePassword 更新用户密码
	UpdatePassword(ctx context.Context, password string, uid int64) error

	// GetAllUIDs 获取所有用户UID
	GetAllUIDs(ctx context.Context) ([]int64, error)
}

