// Package service 实现业务逻辑层
package service

import (
	"context"
	"errors"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"gorm.io/gorm"
)

// UserService 定义用户业务服务接口
type UserService interface {
	// Register 用户注册
	Register(ctx context.Context, params *dto.UserCreateRequest) (*UserDTO, error)

	// Login 用户登录
	Login(ctx context.Context, params *dto.UserLoginRequest, clientIP string) (*UserDTO, error)

	// ChangePassword 修改密码
	ChangePassword(ctx context.Context, uid int64, params *dto.UserChangePasswordRequest) error

	// GetInfo 获取用户信息
	GetInfo(ctx context.Context, uid int64) (*UserDTO, error)

	// GetAllUIDs 获取所有用户的 UID
	GetAllUIDs(ctx context.Context) ([]int64, error)
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

// userService 实现 UserService 接口
type userService struct {
	userRepo     domain.UserRepository
	tokenManager app.TokenManager
}

// NewUserService 创建 UserService 实例
func NewUserService(userRepo domain.UserRepository, tokenManager app.TokenManager) UserService {
	return &userService{
		userRepo:     userRepo,
		tokenManager: tokenManager,
	}
}

// domainToDTO 将领域模型转换为 DTO
func (s *userService) domainToDTO(user *domain.User) *UserDTO {
	if user == nil {
		return nil
	}
	return &UserDTO{
		UID:       user.UID,
		Email:     user.Email,
		Username:  user.Username,
		Token:     user.Token,
		Avatar:    user.Avatar,
		UpdatedAt: timex.Time(user.UpdatedAt),
		CreatedAt: timex.Time(user.CreatedAt),
	}
}

// Register 用户注册
func (s *userService) Register(ctx context.Context, params *dto.UserCreateRequest) (*UserDTO, error) {
	// 检查注册是否启用
	if !global.Config.User.RegisterIsEnable {
		return nil, code.ErrorUserRegisterIsDisable
	}

	// 验证用户名格式
	if !util.IsValidUsername(params.Username) {
		return nil, code.ErrorUserUsernameNotValid
	}

	// 验证密码一致性
	if params.Password != params.ConfirmPassword {
		return nil, code.ErrorUserPasswordNotMatch
	}

	// 检查邮箱是否已存在
	emailUser, err := s.userRepo.GetByEmail(ctx, params.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, code.ErrorDBQuery
	}
	if emailUser != nil {
		return nil, code.ErrorUserEmailAlreadyExists
	}

	// 检查用户名是否已存在
	nameUser, err := s.userRepo.GetByUsername(ctx, params.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, code.ErrorDBQuery
	}
	if nameUser != nil {
		return nil, code.ErrorUserAlreadyExists
	}

	// 生成密码哈希
	password, err := util.GeneratePasswordHash(params.Password)
	if err != nil {
		return nil, code.ErrorPasswordNotValid
	}

	// 创建用户
	newUser := &domain.User{
		Username: params.Username,
		Email:    params.Email,
		Password: password,
	}

	user, err := s.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, err
	}

	// 生成 Token
	token, err := s.tokenManager.Generate(user.UID, "", "")
	if err != nil {
		return nil, err
	}

	dto := s.domainToDTO(user)
	dto.Token = token
	return dto, nil
}

// Login 用户登录
func (s *userService) Login(ctx context.Context, params *dto.UserLoginRequest, clientIP string) (*UserDTO, error) {
	var user *domain.User
	var err error

	// 根据凭证类型查找用户
	if util.IsValidEmail(params.Credentials) {
		user, err = s.userRepo.GetByEmail(ctx, params.Credentials)
		if err != nil {
			return nil, code.ErrorUserNotFound
		}
	} else {
		user, err = s.userRepo.GetByUsername(ctx, params.Credentials)
		if err != nil {
			return nil, code.ErrorUserNotFound
		}
	}

	// 验证密码
	if !util.CheckPasswordHash(user.Password, params.Password) {
		return nil, code.ErrorUserLoginPasswordFailed
	}

	// 生成 Token
	token, err := s.tokenManager.Generate(user.UID, user.Username, clientIP)
	if err != nil {
		return nil, err
	}

	dto := s.domainToDTO(user)
	dto.Token = token
	return dto, nil
}

// ChangePassword 修改密码
func (s *userService) ChangePassword(ctx context.Context, uid int64, params *dto.UserChangePasswordRequest) error {
	// 验证密码一致性
	if params.Password != params.ConfirmPassword {
		return code.ErrorUserPasswordNotMatch
	}

	// 获取用户
	user, err := s.userRepo.GetByUID(ctx, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.ErrorUserNotFound
		}
		return code.ErrorDBQuery
	}

	// 验证旧密码
	if !util.CheckPasswordHash(user.Password, params.OldPassword) {
		return code.ErrorUserOldPasswordFailed
	}

	// 生成新密码哈希
	password, err := util.GeneratePasswordHash(params.Password)
	if err != nil {
		return code.ErrorPasswordNotValid
	}

	// 更新密码
	return s.userRepo.UpdatePassword(ctx, password, uid)
}

// GetInfo 获取用户信息
func (s *userService) GetInfo(ctx context.Context, uid int64) (*UserDTO, error) {
	user, err := s.userRepo.GetByUID(ctx, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, code.ErrorDBQuery
	}
	return s.domainToDTO(user), nil
}

// GetAllUIDs 获取所有用户的 UID
func (s *userService) GetAllUIDs(ctx context.Context) ([]int64, error) {
	return s.userRepo.GetAllUIDs(ctx)
}

// 确保 userService 实现了 UserService 接口
var _ UserService = (*userService)(nil)
