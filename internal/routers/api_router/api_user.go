package api_router

import (
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// User 用户 API 路由处理器
// 结构体名: User
// 说明: 处理用户注册、登录、信息查询及密码修改等 HTTP 请求。
type User struct {
}

// NewUser 创建 User 路由处理器实例
// 函数名: NewUser
// 函数使用说明: 初始化并返回一个新的 User 结构体实例。
// 返回值说明:
//   - *User: 初始化后的 User 实例
func NewUser() *User {
	return &User{}
}

// Register 用户注册
// 函数名: Register
// 函数使用说明: 处理用户注册的 HTTP 请求。验证参数，调用 Service 层执行注册逻辑，并返回注册结果。
// 参数说明:
//   - c *gin.Context: Gin 上下文，包含注册参数 (username, password, etc.)
//
// 返回值说明:
//   - JSON: 注册成功或失败的响应信息
func (h *User) Register(c *gin.Context) {
	response := app.NewResponse(c)
	params := &service.UserCreateRequest{}

	valid, errs := app.BindAndValid(c, params)

	if !valid {
		global.Logger.Error("api_router.user.Register.BindAndValid errs: %v", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c)
	svcData, err := svc.UserRegister(params)

	if err != nil {
		global.Logger.Error("api_router.user.Register svc UserRegister err: %v", zap.Error(err))
		response.ToResponse(code.ErrorUserRegister.WithDetails(err.Error()))
		return
	}

	response.ToResponse(code.Success.WithData(svcData))
}

// Login 用户登录
// 函数名: Login
// 函数使用说明: 处理用户登录的 HTTP 请求。验证参数，调用 Service 层执行登录逻辑，并返回认证 Token。
// 参数说明:
//   - c *gin.Context: Gin 上下文，包含登录参数 (username, password)
//
// 返回值说明:
//   - JSON: 登录成功返回 Token，失败返回错误信息
func (h *User) Login(c *gin.Context) {

	params := &service.UserLoginRequest{}
	response := app.NewResponse(c)

	valid, errs := app.BindAndValid(c, params)

	if !valid {
		global.Logger.Error("api_router.user.Login.BindAndValid errs: %v", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c)
	svcData, err := svc.UserLogin(params)

	if err != nil {

		global.Logger.Error("api_router.user.Login svc UserLogin err: %v", zap.Error(err))
		response.ToResponse(code.ErrorUserLoginFailed.WithDetails(err.Error()))
		return
	}

	response.ToResponse(code.Success.WithData(svcData))
}

// UserChangePassword 修改用户密码
// 函数名: UserChangePassword
// 函数使用说明: 处理修改用户密码的 HTTP 请求。验证用户身份和参数，调用 Service 层更新密码。
// 参数说明:
//   - c *gin.Context: Gin 上下文，包含旧密码和新密码参数
//
// 返回值说明:
//   - JSON: 操作结果
func (h *User) UserChangePassword(c *gin.Context) {
	params := &service.UserChangePasswordRequest{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, params)
	if !valid {
		global.Logger.Error("apiRouter.user.UserChangePassword.BindAndValid errs: %v", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}
	uid := app.GetUID(c)
	if uid == 0 {
		global.Logger.Error("apiRouter.user.UserChangePassword err uid=0")
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}
	svc := service.New(c)
	err := svc.UserChangePassword(uid, params)
	if err != nil {
		global.Logger.Error("apiRouter.user.UserChangePassword svc UserChangePassword err: %v", zap.Error(err))
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}
	response.ToResponse(code.SuccessPasswordUpdate)
}

// UserInfo 获取用户信息
// 函数名: UserInfo
// 函数使用说明: 处理获取当前用户信息的 HTTP 请求。验证用户身份，调用 Service 层获取用户详情。
// 参数说明:
//   - c *gin.Context: Gin 上下文，包含用户认证信息
//
// 返回值说明:
//   - JSON: 包含用户信息的响应数据
func (h *User) UserInfo(c *gin.Context) {
	response := app.NewResponse(c)
	uid := app.GetUID(c)
	if uid == 0 {
		global.Logger.Error("apiRouter.user.UserInfo err uid=0")
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}
	svc := service.New(c)
	user, err := svc.UserInfo(uid)
	if err != nil {
		global.Logger.Error("apiRouter.user.UserInfo svc UserInfo err: %v", zap.Error(err))
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}
	response.ToResponse(code.Success.WithData(user))
}
