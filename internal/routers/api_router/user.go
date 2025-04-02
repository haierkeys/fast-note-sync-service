package api_router

import (
	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/haierkeys/obsidian-better-sync-service/internal/service"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/app"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type User struct {
}

func NewUser() *User {
	return &User{}
}

// Register 用户注册
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
