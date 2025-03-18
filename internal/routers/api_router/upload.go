package api_router

import (
	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/haierkeys/obsidian-better-sync-service/internal/service"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/app"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Upload struct{}

func NewUpload() Upload {
	return Upload{}
}

// Upload 上传文件
func (u Upload) Upload(c *gin.Context) {

	params := &service.ClientUploadParams{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, params)

	if !valid {
		global.Logger.Error("api_router.UserUpload.BindAndValid errs: %v", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	var svcUploadFileData *service.FileInfo
	var svc = service.New(c)
	var err error

	file, fileHeader, errf := c.Request.FormFile("imagefile")
	if errf != nil {
		global.Logger.Error("api_router.UserUpload.ErrorInvalidParams len 0", zap.Error(errf))
		response.ToResponse(code.ErrorInvalidParams)
	}
	defer file.Close()

	svcUploadFileData, err = svc.UploadFile(file, fileHeader, params)
	if err != nil {
		global.Logger.Error("svc.UploadFile err: %v", zap.Error(err))
		response.ToResponse(code.ErrorUploadFileFailed.WithDetails(err.Error()))
		return
	}

	response.ToResponse(code.Success.WithData(svcUploadFileData))

}

// Upload 上传文件
func (u Upload) UserUpload(c *gin.Context) {

	params := &service.ClientUploadParams{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, params)

	if !valid {
		global.Logger.Error("api_router.UserUpload.BindAndValid errs: %v", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	var svcUploadFileData *service.FileInfo
	var svc = service.New(c)
	var err error

	file, fileHeader, errf := c.Request.FormFile("imagefile")
	if errf != nil {
		global.Logger.Error("api_router.UserUpload.ErrorInvalidParams len 0", zap.Error(errf))
		response.ToResponse(code.ErrorInvalidParams)
	}
	defer file.Close()

	uid := app.GetUid(c)
	if uid == 0 {
		global.Logger.Error("api_router.UserUpload svc UserLogin err uid=0")
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}
	svcUploadFileData, err = svc.UserUploadFile(uid, file, fileHeader, params)
	if err != nil {
		global.Logger.Error("svc.UplUserUploadFileoadFile err: %v", zap.Error(err))
		response.ToResponse(code.ErrorUploadFileFailed.WithDetails(err.Error()))
		return
	}

	response.ToResponse(code.Success.WithData(svcUploadFileData))
}
