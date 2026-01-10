package domain

import "time"

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
