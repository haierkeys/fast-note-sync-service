package api_router

import (
	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"go.uber.org/zap"
)

type WebGUI struct{}

func NewWebGUI() *WebGUI {
	return &WebGUI{}
}

type webGUI struct {
	FontSet                 string `json:"fontSet" form:"fontSet"`
	RegisterIsEnable        bool   `json:"registerIsEnable" form:"registerIsEnable"`
	FileChunkSize           string `json:"fileChunkSize" form:"fileChunkSize"`
	SoftDeleteRetentionTime string `json:"softDeleteRetentionTime" form:"softDeleteRetentionTime"`
	UploadSessionTimeout    string `json:"uploadSessionTimeout" form:"uploadSessionTimeout"`
	AdminUID                int    `json:"adminUid" form:"adminUid"`
}

func (w *WebGUI) Config(c *gin.Context) {
	response := app.NewResponse(c)
	data := webGUI{
		FontSet:          global.Config.WebGUI.FontSet,
		RegisterIsEnable: global.Config.User.RegisterIsEnable,
		AdminUID:         global.Config.User.AdminUID,
	}
	response.ToResponse(code.Success.WithData(data))
}

func (w *WebGUI) GetConfig(c *gin.Context) {

	response := app.NewResponse(c)
	uid := app.GetUID(c)
	if uid == 0 {
		global.Logger.Error("apiRouter.Note.Get err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}
	// 当登录用户不是管理员时，直接报错
	if global.Config.User.AdminUID != 0 && uid != int64(global.Config.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	data := &webGUI{
		FontSet:                 global.Config.WebGUI.FontSet,
		RegisterIsEnable:        global.Config.User.RegisterIsEnable,
		FileChunkSize:           global.Config.App.FileChunkSize,
		SoftDeleteRetentionTime: global.Config.App.SoftDeleteRetentionTime,
		UploadSessionTimeout:    global.Config.App.UploadSessionTimeout,
		AdminUID:                global.Config.User.AdminUID,
	}

	response.ToResponse(code.Success.WithData(data))
}

func (w *WebGUI) UpdateConfig(c *gin.Context) {

	params := &webGUI{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, params)
	if !valid {
		global.Logger.Error("apiRouter.Note.CreateOrUpdate.BindAndValid err: %v", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}
	uid := app.GetUID(c)
	if uid == 0 {
		global.Logger.Error("apiRouter.Note.CreateOrUpdate err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}
	// 当登录用户不是管理员时，直接报错
	if global.Config.User.AdminUID != 0 && uid != int64(global.Config.User.AdminUID) {
		response.ToResponse(code.ErrorUserIsNotAdmin)
		return
	}

	global.Dump(params)

	global.Config.WebGUI.FontSet = params.FontSet
	global.Config.User.RegisterIsEnable = params.RegisterIsEnable
	global.Config.App.FileChunkSize = params.FileChunkSize
	global.Config.App.SoftDeleteRetentionTime = params.SoftDeleteRetentionTime
	global.Config.App.UploadSessionTimeout = params.UploadSessionTimeout
	global.Config.User.AdminUID = params.AdminUID
	response.ToResponse(code.Success.WithData(params))
	global.Config.Save()

}
