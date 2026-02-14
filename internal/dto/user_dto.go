// Package dto Defines data transfer objects (request parameters and response structs)
// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// UserCreateRequest User registration request parameters
// 用户注册请求参数
type UserCreateRequest struct {
	Email           string `json:"email" form:"email" binding:"required,email"`               // User email // 用户邮件
	Username        string `json:"username" form:"username" binding:"required"`               // User name // 用户名
	Password        string `json:"password" form:"password" binding:"required"`               // User password // 用户密码
	ConfirmPassword string `json:"confirmPassword" form:"confirmPassword" binding:"required"` // Confirm password // 校验密码
}

// UserLoginRequest User login request parameters
// 用户登录请求参数
type UserLoginRequest struct {
	Credentials string `form:"credentials" binding:"required"` // Username or Email // 登录凭证（用户名或邮件）
	Password    string `form:"password" binding:"required"`    // Password // 密码
}

// UserRegisterSendEmailRequest Request parameters for sending registration email
// 发送注册邮件请求参数
type UserRegisterSendEmailRequest struct {
	Email string `json:"email" form:"email" binding:"required,email"` // User email // 用户邮件
}

// UserChangePasswordRequest Request parameters for changing password
// 修改密码请求参数
type UserChangePasswordRequest struct {
	OldPassword     string `json:"oldPassword" form:"oldPassword" binding:"required"`         // Old password // 旧密码
	Password        string `json:"password" form:"password" binding:"required"`               // New password // 新密码
	ConfirmPassword string `json:"confirmPassword" form:"confirmPassword" binding:"required"` // Confirm password // 校验密码
}

// ---------------- DTO / Response ----------------

// UserDTO User data transfer object
// UserDTO 用户数据传输对象
type UserDTO struct {
	UID       int64      `json:"uid"`       // User ID (primary key) // 用户唯一标识（主键）
	Email     string     `json:"email"`     // Email address // 邮件地址
	Username  string     `json:"username"`  // Username // 用户名
	Token     string     `json:"token"`     // Authentication Token // 认证 Token
	Avatar    string     `json:"avatar"`    // Avatar URL or handle // 头像路径或名称
	UpdatedAt timex.Time `json:"updatedAt"` // Last updated time // 最后更新时间
	CreatedAt timex.Time `json:"createdAt"` // Account created time // 账号创建时间
}
