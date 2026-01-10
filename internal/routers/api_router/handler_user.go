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

// UserHandler 用户 API 路由处理器
// 使用 App Container 注入依赖，支持统一错误处理
type UserHandler struct {
	*Handler
}

// NewUserHandler 创建 UserHandler 实例
func NewUserHandler(a *app.App) *UserHandler {
	return &UserHandler{
		Handler: NewHandler(a),
	}
}

// Register 用户注册
// 处理用户注册的 HTTP 请求，验证参数并调用 UserService
func (h *UserHandler) Register(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.UserCreateRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("UserHandler.Register.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取请求上下文（包含 Trace ID）
	ctx := c.Request.Context()

	// 调用 UserService 执行注册
	userDTO, err := h.App.UserService.Register(ctx, params)
	if err != nil {
		h.logError(ctx, "UserHandler.Register", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(userDTO))
}

// Login 用户登录
// 处理用户登录的 HTTP 请求，验证参数并返回认证 Token
func (h *UserHandler) Login(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.UserLoginRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("UserHandler.Login.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取请求上下文和客户端 IP
	ctx := c.Request.Context()
	clientIP := c.ClientIP()

	// 调用 UserService 执行登录
	userDTO, err := h.App.UserService.Login(ctx, params, clientIP)
	if err != nil {
		h.logError(ctx, "UserHandler.Login", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(userDTO))
}

// UserChangePassword 修改用户密码
// 处理修改用户密码的 HTTP 请求
func (h *UserHandler) UserChangePassword(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.UserChangePasswordRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("UserHandler.UserChangePassword.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("UserHandler.UserChangePassword err uid=0")
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	// 获取请求上下文
	ctx := c.Request.Context()

	// 调用 UserService 修改密码
	err := h.App.UserService.ChangePassword(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "UserHandler.UserChangePassword", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.SuccessPasswordUpdate)
}

// UserInfo 获取用户信息
// 处理获取当前用户信息的 HTTP 请求
func (h *UserHandler) UserInfo(c *gin.Context) {
	response := pkgapp.NewResponse(c)

	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("UserHandler.UserInfo err uid=0")
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	// 获取请求上下文
	ctx := c.Request.Context()

	// 调用 UserService 获取用户信息
	userDTO, err := h.App.UserService.GetInfo(ctx, uid)
	if err != nil {
		h.logError(ctx, "UserHandler.UserInfo", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(userDTO))
}

// logError 记录错误日志，包含 Trace ID
func (h *UserHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}
