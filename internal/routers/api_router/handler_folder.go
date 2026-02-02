package api_router

import (
	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	apperrors "github.com/haierkeys/fast-note-sync-service/pkg/errors"
)

type FolderHandler struct {
	appContainer *app.App
}

func NewFolderHandler(appContainer *app.App) *FolderHandler {
	return &FolderHandler{appContainer: appContainer}
}

func (h *FolderHandler) List(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	var params dto.FolderListRequest
	if err := c.ShouldBindQuery(&params); err != nil {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(err.Error()))
		return
	}

	uid := pkgapp.GetUID(c)
	res, err := h.appContainer.FolderService.List(c.Request.Context(), uid, &params)
	if err != nil {
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(res))
}

func (h *FolderHandler) Create(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	var params dto.FolderCreateRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(err.Error()))
		return
	}

	uid := pkgapp.GetUID(c)
	res, err := h.appContainer.FolderService.UpdateOrCreate(c.Request.Context(), uid, &params)
	if err != nil {
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(res))
}

func (h *FolderHandler) Delete(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	var params dto.FolderDeleteRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(err.Error()))
		return
	}

	uid := pkgapp.GetUID(c)
	err := h.appContainer.FolderService.Delete(c.Request.Context(), uid, &params)
	if err != nil {
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success)
}
