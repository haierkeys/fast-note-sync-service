package api_router

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	apperrors "github.com/haierkeys/fast-note-sync-service/pkg/errors"
	"go.uber.org/zap"
)

// UserHandler user API router handler
// UserHandler 用户 API 路由处理器
// Uses App Container to inject dependencies, supports unified error handling
// 使用 App Container 注入依赖，支持统一错误处理
type UserHandler struct {
	*Handler
}

// NewUserHandler creates UserHandler instance
// NewUserHandler 创建 UserHandler 实例
func NewUserHandler(a *app.App) *UserHandler {
	return &UserHandler{
		Handler: NewHandler(a),
	}
}

// Register user registration
// @Summary User registration
// @Description Handle user registration HTTP request, validate parameters and call UserService. Registration may be disabled in server settings.
// @Description 处理用户注册 HTTP 请求，验证参数并调用 UserService。注册功能可能在服务器设置中被禁用。
// @Tags User
// @Accept json
// @Produce json
// @Param params body dto.UserCreateRequest true "Register Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.UserDTO} "Success"
// @Failure 400 {object} pkgapp.Res "Invalid Parameters / Registration Disabled / User Already Exists"
// @Router /api/user/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.UserCreateRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("UserHandler.Register.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get request context (including Trace ID)
	// 获取请求上下文（包含 Trace ID）
	ctx := c.Request.Context()

	// Call UserService to perform registration
	// 调用 UserService 执行注册
	userDTO, err := h.App.UserService.Register(ctx, params)
	if err != nil {
		h.logError(ctx, "UserHandler.Register", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(userDTO))
}

// Login user login
// @Summary User login
// @Description Handle user login HTTP request, validate parameters and return auth token.
// @Description 处理用户登录 HTTP 请求，验证参数并返回认证 Token。
// @Tags User
// @Accept json
// @Produce json
// @Param params body dto.UserLoginRequest true "Login Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.UserDTO} "Success"
// @Failure 400 {object} pkgapp.Res "Invalid Parameters / Invalid Credentials"
// @Router /api/user/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.UserLoginRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("UserHandler.Login.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get request context and client IP
	// 获取请求上下文和客户端 IP
	ctx := c.Request.Context()
	clientIP := c.ClientIP()

	// Call UserService to perform login
	// 调用 UserService 执行登录
	userDTO, err := h.App.UserService.Login(ctx, params, clientIP)
	if err != nil {
		h.logError(ctx, "UserHandler.Login", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(userDTO))
}

// UserChangePassword changes user password
// @Summary Change user password
// @Description Handle password change request for current user, validate old password and update new password.
// @Description 处理当前用户的修改密码请求，验证旧密码并更新新密码。
// @Tags User
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.UserChangePasswordRequest true "Change Password Parameters"
// @Success 200 {object} pkgapp.Res "Success"
// @Failure 400 {object} pkgapp.Res "Invalid Parameters / Old Password Incorrect"
// @Failure 401 {object} pkgapp.Res "Unauthorized"
// @Router /api/user/change_password [post]
func (h *UserHandler) UserChangePassword(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.UserChangePasswordRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("UserHandler.UserChangePassword.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("UserHandler.UserChangePassword err uid=0")
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	// Call UserService to change password
	// 调用 UserService 修改密码
	err := h.App.UserService.ChangePassword(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "UserHandler.UserChangePassword", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.SuccessPasswordUpdate)
}

// UserInfo retrieves user info
// @Summary Get user info
// @Description Handle request to get current user info.
// @Description 处理获取当前用户信息的请求。
// @Tags User
// @Accept json
// @Produce json
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Success 200 {object} pkgapp.Res{data=dto.UserDTO} "Success"
// @Failure 401 {object} pkgapp.Res "Unauthorized"
// @Router /api/user/info [get]
func (h *UserHandler) UserInfo(c *gin.Context) {
	response := pkgapp.NewResponse(c)

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("UserHandler.UserInfo err uid=0")
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	// Call UserService to get user info
	// 调用 UserService 获取用户信息
	userDTO, err := h.App.UserService.GetInfo(ctx, uid)
	if err != nil {
		h.logError(ctx, "UserHandler.UserInfo", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(userDTO))
}

// logError records error log, including Trace ID
// logError 记录错误日志，包含 Trace ID
func (h *UserHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}
