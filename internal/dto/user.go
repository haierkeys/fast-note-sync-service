// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// UserCreateRequest 用户注册请求参数
type UserCreateRequest struct {
	Email           string `json:"email" form:"email" binding:"required,email"`
	Username        string `json:"username" form:"username" binding:"required"`
	Password        string `json:"password" form:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" form:"confirmPassword" binding:"required"`
}

// UserLoginRequest 用户登录请求参数
type UserLoginRequest struct {
	Credentials string `form:"credentials" binding:"required"`
	Password    string `form:"password" binding:"required"`
}

// UserRegisterSendEmailRequest 发送注册邮件请求参数
type UserRegisterSendEmailRequest struct {
	Email string `json:"email" form:"email" binding:"required,email"`
}

// UserChangePasswordRequest 修改密码请求参数
type UserChangePasswordRequest struct {
	OldPassword     string `json:"oldPassword" form:"oldPassword" binding:"required"`
	Password        string `json:"password" form:"password" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" form:"confirmPassword" binding:"required"`
}

// UserDTO 用户数据传输对象
type UserDTO struct {
	UID       int64      `json:"uid"`
	Email     string     `json:"email"`
	Username  string     `json:"username"`
	Token     string     `json:"token"`
	Avatar    string     `json:"avatar"`
	UpdatedAt timex.Time `json:"updatedAt"`
	CreatedAt timex.Time `json:"createdAt"`
}
