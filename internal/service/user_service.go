// Package service 实现业务逻辑层
package service

import (
	"context"
	"errors"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserService 定义用户业务服务接口
type UserService interface {
	// Register 用户注册
	Register(ctx context.Context, params *dto.UserCreateRequest) (*dto.UserDTO, error)

	// Login 用户登录
	Login(ctx context.Context, params *dto.UserLoginRequest, clientIP string) (*dto.UserDTO, error)

	// ChangePassword 修改密码
	ChangePassword(ctx context.Context, uid int64, params *dto.UserChangePasswordRequest) error

	// GetInfo 获取用户信息
	GetInfo(ctx context.Context, uid int64) (*dto.UserDTO, error)

	// GetAllUIDs 获取所有用户的 UID
	GetAllUIDs(ctx context.Context) ([]int64, error)
}

// userService 实现 UserService 接口
type userService struct {
	userRepo     domain.UserRepository
	tokenManager app.TokenManager
	logger       *zap.Logger
	config       *ServiceConfig
}

// NewUserService 创建 UserService 实例
func NewUserService(userRepo domain.UserRepository, tokenManager app.TokenManager, logger *zap.Logger, config *ServiceConfig) UserService {
	return &userService{
		userRepo:     userRepo,
		tokenManager: tokenManager,
		logger:       logger,
		config:       config,
	}
}

// domainToDTO 将领域模型转换为 DTO
func (s *userService) domainToDTO(user *domain.User) *dto.UserDTO {
	if user == nil {
		return nil
	}
	return &dto.UserDTO{
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
func (s *userService) Register(ctx context.Context, params *dto.UserCreateRequest) (*dto.UserDTO, error) {
	// 检查注册是否启用
	if s.config == nil || !s.config.User.RegisterIsEnable {
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
		return nil, code.ErrorUserRegister.WithDetails(err.Error())
	}

	// 生成 Token
	token, err := s.tokenManager.Generate(user.UID, "", "")
	if err != nil {
		return nil, code.ErrorTokenGenerate.WithDetails(err.Error())
	}

	dto := s.domainToDTO(user)
	dto.Token = token
	return dto, nil
}

// Login 用户登录
func (s *userService) Login(ctx context.Context, params *dto.UserLoginRequest, clientIP string) (*dto.UserDTO, error) {
	var user *domain.User
	var err error

	// 根据凭证类型查找用户
	if util.IsValidEmail(params.Credentials) {
		user, err = s.userRepo.GetByEmail(ctx, params.Credentials)
		if err != nil {
			// 安全考虑：不暴露用户是否存在，统一返回用户名或密码错误
			return nil, code.ErrorUserLoginPasswordFailed
		}
	} else {
		user, err = s.userRepo.GetByUsername(ctx, params.Credentials)
		if err != nil {
			// 安全考虑：不暴露用户是否存在，统一返回用户名或密码错误
			return nil, code.ErrorUserLoginPasswordFailed
		}
	}

	// 验证密码
	if !util.CheckPasswordHash(user.Password, params.Password) {
		return nil, code.ErrorUserLoginPasswordFailed
	}

	// 生成 Token
	token, err := s.tokenManager.Generate(user.UID, user.Username, clientIP)
	if err != nil {
		return nil, code.ErrorTokenGenerate.WithDetails(err.Error())
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
func (s *userService) GetInfo(ctx context.Context, uid int64) (*dto.UserDTO, error) {
	user, err := s.userRepo.GetByUID(ctx, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		if s.logger != nil {
			s.logger.Error("UserService.GetInfo failed",
				zap.Int64("uid", uid),
				zap.Error(err),
			)
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
