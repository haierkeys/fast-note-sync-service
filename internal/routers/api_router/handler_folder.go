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

// List retrieves folder list
// @Summary Get folder list
// @Description Get folder list for current user by parent path or pathHash
// @Tags Folder
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params query dto.FolderListRequest true "Query Parameters"
// @Success 200 {object} pkgapp.Res{data=[]dto.FolderDTO} "Success"
// @Router /api/folder [get]
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

// Create creates a folder
// @Summary Create folder
// @Description Create a new folder or restore a deleted one by path
// @Tags Folder
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.FolderCreateRequest true "Create Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.FolderDTO} "Success"
// @Router /api/folder [post]
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

// Delete deletes a folder
// @Summary Delete folder
// @Description Soft delete a folder by path or pathHash
// @Tags Folder
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.FolderDeleteRequest true "Delete Parameters"
// @Success 200 {object} pkgapp.Res "Success"
// @Router /api/folder [delete]
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

// ListNotes retrieves notes in a folder
// @Summary List notes in folder
// @Description List non-deleted notes in a specific folder with pagination and sorting
// @Tags Folder
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params query dto.FolderContentRequest true "Query Parameters"
// @Success 200 {object} pkgapp.Res{data=[]dto.NoteDTO} "Success"
// @Router /api/folder/notes [get]
func (h *FolderHandler) ListNotes(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	var params dto.FolderContentRequest
	if err := c.ShouldBindQuery(&params); err != nil {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(err.Error()))
		return
	}

	uid := pkgapp.GetUID(c)
	res, err := h.appContainer.FolderService.ListNotes(c.Request.Context(), uid, &params)
	if err != nil {
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(res))
}

// ListFiles retrieves files in a folder
// @Summary List files in folder
// @Description List non-deleted files in a specific folder with pagination and sorting
// @Tags Folder
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params query dto.FolderContentRequest true "Query Parameters"
// @Success 200 {object} pkgapp.Res{data=[]dto.FileDTO} "Success"
// @Router /api/folder/files [get]
func (h *FolderHandler) ListFiles(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	var params dto.FolderContentRequest
	if err := c.ShouldBindQuery(&params); err != nil {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(err.Error()))
		return
	}

	uid := pkgapp.GetUID(c)
	res, err := h.appContainer.FolderService.ListFiles(c.Request.Context(), uid, &params)
	if err != nil {
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(res))
}
