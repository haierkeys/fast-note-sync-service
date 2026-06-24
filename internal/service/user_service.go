// Package service implements the business logic layer
// Package service 实现业务逻辑层
package service

import (
	"context"
	"errors"
	"strings"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserService defines the user business service interface
// UserService 定义用户业务服务接口
type UserService interface {
	// Register user registration
	// Register 用户注册
	Register(ctx context.Context, params *dto.UserCreateRequest, clientIP string, clientType string, userAgent string) (*dto.UserDTO, error)

	// Create user
	Create(ctx context.Context, params *dto.UserCreateRequest) (*dto.UserDTO, error)

	// Update user
	Update(ctx context.Context, params *dto.UserUpdateRequest) error

	// Login user login
	// Login 用户登录
	Login(ctx context.Context, params *dto.UserLoginRequest, clientIP string, clientType string, userAgent string) (*dto.UserDTO, error)

	// ChangePassword change user password
	// ChangePassword 修改密码
	ChangePassword(ctx context.Context, uid int64, params *dto.UserChangePasswordRequest) error

	// GetInfo retrieves user information
	// GetInfo 获取用户信息
	GetInfo(ctx context.Context, uid int64) (*dto.UserDTO, error)

	// GetAllUIDs retrieves all user UIDs
	// GetAllUIDs 获取所有用户的 UID
	GetAllUIDs(ctx context.Context) ([]int64, error)

	// GetList retrieves users with pagination // GetList 分页获取用户列表
	GetList(ctx context.Context, pager *app.Pager) ([]*dto.UserDTO, int64, error)

	// IsRegisterEnabled checks if registration is allowed
	// IsRegisterEnabled 检查是否允许注册
	IsRegisterEnabled(ctx context.Context) bool
}

// userService implementation of UserService interface
// userService 实现 UserService 接口
type userService struct {
	userRepo     domain.UserRepository // User repository // 用户仓库
	tokenManager app.TokenManager      // Token manager // Token 管理器
	tokenService TokenService          // Token service // Token 服务
	logger       *zap.Logger           // Logger // 日志器
	config       *ServiceConfig        // Service configuration // 服务配置
}

// NewUserService creates UserService instance
// NewUserService 创建 UserService 实例
func NewUserService(userRepo domain.UserRepository, tokenManager app.TokenManager, tokenService TokenService, logger *zap.Logger, config *ServiceConfig) UserService {
	return &userService{
		userRepo:     userRepo,
		tokenManager: tokenManager,
		tokenService: tokenService,
		logger:       logger,
		config:       config,
	}
}

// domainToDTO converts domain model to DTO
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
		IsDeleted: user.IsDeleted,
		UpdatedAt: timex.Time(user.UpdatedAt),
		CreatedAt: timex.Time(user.CreatedAt),
	}
}

// Register user registration
// Register 用户注册
func (s *userService) Register(ctx context.Context, params *dto.UserCreateRequest, clientIP string, clientType string, userAgent string) (*dto.UserDTO, error) {
	// Check if registration is enabled
	// 检查注册是否启用
	if !s.IsRegisterEnabled(ctx) {
		return nil, code.ErrorUserRegisterIsDisable
	}

	// Validate username format
	// 验证用户名格式
	if !util.IsValidUsername(params.Username) {
		return nil, code.ErrorUserUsernameNotValid
	}

	// Validate password consistency
	// 验证密码一致性
	if params.Password != params.ConfirmPassword {
		return nil, code.ErrorUserPasswordNotMatch
	}

	// Check if email already exists
	// 检查邮箱是否已存在
	params.Email = strings.ToLower(strings.TrimSpace(params.Email))
	emailUser, err := s.userRepo.GetByEmail(ctx, params.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, code.ErrorDBQuery
	}
	if emailUser != nil {
		return nil, code.ErrorUserEmailAlreadyExists
	}

	// Check if username already exists
	// 检查用户名是否已存在
	nameUser, err := s.userRepo.GetByUsername(ctx, params.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, code.ErrorDBQuery
	}
	if nameUser != nil {
		return nil, code.ErrorUserAlreadyExists
	}

	// Generate password hash
	// 生成密码哈希
	password, err := util.GeneratePasswordHash(params.Password)
	if err != nil {
		return nil, code.ErrorPasswordNotValid
	}

	// Create user
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

	// Generate Token with proper IP and UA binding
	token, tokenStr, err := s.tokenService.CreateForLogin(ctx, user.UID, clientType, clientIP, userAgent)
	if err != nil {
		return nil, code.ErrorTokenGenerate.WithDetails(err.Error())
	}

	dto := s.domainToDTO(user)
	dto.Token = tokenStr
	dto.TokenID = token.ID
	return dto, nil
}

// Create user
func (s *userService) Create(ctx context.Context, params *dto.UserCreateRequest) (*dto.UserDTO, error) {
	// Validate username format
	// 验证用户名格式
	if !util.IsValidUsername(params.Username) {
		return nil, code.ErrorUserUsernameNotValid
	}

	// Validate password consistency
	// 验证密码一致性
	if params.Password != params.ConfirmPassword {
		return nil, code.ErrorUserPasswordNotMatch
	}

	// Check if email already exists
	// 检查邮箱是否已存在
	params.Email = strings.ToLower(strings.TrimSpace(params.Email))
	emailUser, err := s.userRepo.GetByEmail(ctx, params.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, code.ErrorDBQuery
	}
	if emailUser != nil {
		return nil, code.ErrorUserEmailAlreadyExists
	}

	// Check if username already exists
	// 检查用户名是否已存在
	nameUser, err := s.userRepo.GetByUsername(ctx, params.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, code.ErrorDBQuery
	}
	if nameUser != nil {
		return nil, code.ErrorUserAlreadyExists
	}

	// Generate password hash
	// 生成密码哈希
	password, err := util.GeneratePasswordHash(params.Password)
	if err != nil {
		return nil, code.ErrorPasswordNotValid
	}

	// Create user
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

	dto := s.domainToDTO(user)
	return dto, nil
}

// Update user
func (s *userService) Update(ctx context.Context, params *dto.UserUpdateRequest) error {
	// Validate username format
	// 验证用户名格式
	if !util.IsValidUsername(params.Username) {
		return code.ErrorUserUsernameNotValid
	}

	// Current user
	currentUser, err := s.userRepo.GetByUID(ctx, params.UID, false)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return code.ErrorDBQuery
	}

	if currentUser == nil {
		return code.ErrorUserNotFound
	}

	// Check if email already exists
	// 检查邮箱是否已存在
	params.Email = strings.ToLower(strings.TrimSpace(params.Email))
	emailUser, err := s.userRepo.GetByEmail(ctx, params.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return code.ErrorDBQuery
	}

	// Prevent check self email
	if emailUser != nil && emailUser.UID != currentUser.UID {
		return code.ErrorUserEmailAlreadyExists
	}

	// Check if username already exists
	// 检查用户名是否已存在
	nameUser, err := s.userRepo.GetByUsername(ctx, params.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return code.ErrorDBQuery
	}

	// Prevent check self username
	if nameUser != nil && nameUser.UID != currentUser.UID {
		return code.ErrorUserAlreadyExists
	}

	var password string
	// Generate new password is not empty
	if strings.TrimSpace(params.Password) != "" {
		// Generate password hash
		// 生成密码哈希
		password, err = util.GeneratePasswordHash(params.Password)
		if err != nil {
			return code.ErrorPasswordNotValid
		}
	}

	// Update user
	updatedUser := &domain.User{
		UID:       params.UID,
		Username:  params.Username,
		Email:     params.Email,
		Password:  password,
		IsDeleted: params.IsDeleted,
	}

	err = s.userRepo.Update(ctx, updatedUser)
	if err != nil {
		return code.ErrorUserUpdate.WithDetails(err.Error())
	}

	return nil
}

// Login user login
// Login 用户登录
func (s *userService) Login(ctx context.Context, params *dto.UserLoginRequest, clientIP string, clientType string, userAgent string) (*dto.UserDTO, error) {
	var user *domain.User
	var err error

	// Find user based on credential type
	// 根据凭证类型查找用户
	if util.IsValidEmail(params.Credentials) {
		email := strings.ToLower(strings.TrimSpace(params.Credentials))
		user, err = s.userRepo.GetByEmail(ctx, email)
		if err != nil {
			return nil, code.ErrorUserLoginPasswordFailed
		}
	} else {
		user, err = s.userRepo.GetByUsername(ctx, params.Credentials)
		if err != nil {
			return nil, code.ErrorUserLoginPasswordFailed
		}
	}

	// Validate password
	// 验证密码
	if !util.CheckPasswordHash(user.Password, params.Password) {
		return nil, code.ErrorUserLoginPasswordFailed
	}

	// Generate Token via TokenService
	// 生成 Token
	var token *domain.AuthToken
	var tokenStr string
	var errToken error

	if params.TokenID > 0 && strings.ToLower(clientType) == "webgui" {
		// Attempt to rotate existing login token
		// 尝试轮转现有的登录令牌
		token, tokenStr, errToken = s.tokenService.RotateForLogin(ctx, user.UID, params.TokenID, clientIP, userAgent)
		if errToken != nil {
			s.logger.Warn("UserService.Login rotate token failed, fallback to create new token",
				zap.Int64("uid", user.UID),
				zap.Int64("tokenId", params.TokenID),
				zap.Error(errToken),
			)
			token, tokenStr, err = s.tokenService.CreateForLogin(ctx, user.UID, clientType, clientIP, userAgent)
		}
	} else {
		token, tokenStr, err = s.tokenService.CreateForLogin(ctx, user.UID, clientType, clientIP, userAgent)
	}

	if err != nil {
		return nil, code.ErrorTokenGenerate.WithDetails(err.Error())
	}

	dto := s.domainToDTO(user)
	dto.Token = tokenStr
	dto.TokenID = token.ID
	return dto, nil
}

// ChangePassword change password
// ChangePassword 修改密码
func (s *userService) ChangePassword(ctx context.Context, uid int64, params *dto.UserChangePasswordRequest) error {
	// Validate password consistency
	// 验证密码一致性
	if params.Password != params.ConfirmPassword {
		return code.ErrorUserPasswordNotMatch
	}

	// Get user
	// 获取用户
	user, err := s.userRepo.GetByUID(ctx, uid, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.ErrorUserNotFound
		}
		return code.ErrorDBQuery
	}

	// Validate old password
	// 验证旧密码
	if !util.CheckPasswordHash(user.Password, params.OldPassword) {
		return code.ErrorUserOldPasswordFailed
	}

	// Generate new password hash
	// 生成新密码哈希
	password, err := util.GeneratePasswordHash(params.Password)
	if err != nil {
		return code.ErrorPasswordNotValid
	}

	// Update password
	// 更新密码
	err = s.userRepo.UpdatePassword(ctx, password, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}
	return nil
}

// GetInfo retrieves user information
// GetInfo 获取用户信息
func (s *userService) GetInfo(ctx context.Context, uid int64) (*dto.UserDTO, error) {
	user, err := s.userRepo.GetByUID(ctx, uid, true)
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

// GetAllUIDs retrieves all user UIDs
// GetAllUIDs 获取所有用户的 UID
func (s *userService) GetAllUIDs(ctx context.Context) ([]int64, error) {
	uids, err := s.userRepo.GetAllUIDs(ctx)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	return uids, nil
}

// GetList retrieves users with pagination // GetList 分页获取用户列表
func (s *userService) GetList(ctx context.Context, pager *app.Pager) ([]*dto.UserDTO, int64, error) {
	offset := app.GetPageOffset(pager.Page, pager.PageSize)
	limit := pager.PageSize

	users, total, err := s.userRepo.GetList(ctx, offset, limit)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var results []*dto.UserDTO
	for _, user := range users {
		results = append(results, s.domainToDTO(user))
	}
	return results, total, nil
}

// IsRegisterEnabled checks if registration is allowed
// IsRegisterEnabled 检查是否允许注册
func (s *userService) IsRegisterEnabled(ctx context.Context) bool {
	// Check if registration is enabled in config
	// 检查配置中是否启用了注册
	if s.config == nil || !s.config.User.RegisterIsEnable {
		return false
	}

	// If AdminUID is 0, registration is only allowed if there are no users
	// 如果 AdminUID 为 0，则仅在没有用户时允许注册
	if s.config.User.AdminUID == 0 {
		uids, err := s.userRepo.GetAllUIDs(ctx)
		if err == nil && len(uids) > 0 {
			return false
		}
	}

	return true
}

// Verify userService implements UserService interface
// 确保 userService 实现了 UserService 接口
var _ UserService = (*userService)(nil)
