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
