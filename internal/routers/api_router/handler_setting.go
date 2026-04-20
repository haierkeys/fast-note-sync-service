package api_router

import (
	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	apperrors "github.com/haierkeys/fast-note-sync-service/pkg/errors"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"go.uber.org/zap"
)

type SettingHandler struct {
	*Handler
}

func NewSettingHandler(appContainer *app.App, wss *pkgapp.WebsocketServer) *SettingHandler {
	return &SettingHandler{
		Handler: NewHandlerWithWSS(appContainer, wss),
	}
}

// Get retrieves a setting
// @Summary Get setting info
// @Description Get setting info for current user by path or pathHash
// @Tags Setting
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params query dto.SettingGetRequest true "Query Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.SettingDTO} "Success"
// @Router /api/setting [get]
func (h *SettingHandler) Get(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.SettingGetRequest{}
	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		h.App.Logger().Error("SettingHandler.Get.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	res, err := h.App.GetSettingService(h.getClientInfo(c)).Get(c.Request.Context(), uid, params)
	if err != nil {
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(res))
}

// List retrieves setting list
// @Summary Get setting list
// @Description Get setting list for current user with pagination and keyword filtering
// @Tags Setting
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params query dto.SettingListRequest true "Query Parameters"
// @Param pagination query pkgapp.PaginationRequest true "Pagination Parameters"
// @Success 200 {object} pkgapp.Res{data=pkgapp.ListRes{list=[]dto.SettingDTO}} "Success"
// @Router /api/settings [get]
func (h *SettingHandler) List(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.SettingListRequest{}
	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		h.App.Logger().Error("SettingHandler.List.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	pager := pkgapp.NewPager(c)

	res, count, err := h.App.GetSettingService(h.getClientInfo(c)).List(c.Request.Context(), uid, params, pager)
	if err != nil {
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponseList(code.Success, res, int(count))
}

// CreateOrUpdate creates or updates a setting
// @Summary Create or update setting
// @Description Create a new setting or update an existing one
// @Tags Setting
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.SettingModifyOrCreateRequest true "Create/Update Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.SettingDTO} "Success"
// @Router /api/setting [post]
func (h *SettingHandler) CreateOrUpdate(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.SettingModifyOrCreateRequest{}
	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		h.App.Logger().Error("SettingHandler.CreateOrUpdate.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	params.Mtime = timex.Now().UnixMilli()
	params.Ctime = timex.Now().UnixMilli()

	uid := pkgapp.GetUID(c)
	_, res, err := h.App.GetSettingService(h.getClientInfo(c)).ModifyOrCreate(c.Request.Context(), uid, params, false)
	if err != nil {
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(res))
	h.WSS.BroadcastToUser(uid, code.Success.WithData(dto.SettingSyncModifyMessage{
		Vault:            params.Vault,
		Path:             res.Path,
		PathHash:         res.PathHash,
		Content:          res.Content,
		ContentHash:      res.ContentHash,
		Ctime:            res.Ctime,
		Mtime:            res.Mtime,
		UpdatedTimestamp: res.UpdatedTimestamp,
	}).WithVault(params.Vault), string(dto.SettingSyncModify))
}

// Delete deletes a setting
// @Summary Delete setting
// @Description Soft delete a setting by path or pathHash
// @Tags Setting
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.SettingDeleteRequest true "Delete Parameters"
// @Success 200 {object} pkgapp.Res "Success"
// @Router /api/setting [delete]
func (h *SettingHandler) Delete(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.SettingDeleteRequest{}
	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		h.App.Logger().Error("SettingHandler.Delete.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	res, err := h.App.GetSettingService(h.getClientInfo(c)).Delete(c.Request.Context(), uid, params)
	if err != nil {
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success)
	h.WSS.BroadcastToUser(uid, code.Success.WithData(dto.SettingSyncDeleteMessage{
		Path:             res.Path,
		PathHash:         res.PathHash,
		Ctime:            res.Ctime,
		Mtime:            res.Mtime,
		UpdatedTimestamp: res.UpdatedTimestamp,
	}).WithVault(params.Vault), string(dto.SettingSyncDelete))
}

// Rename renames a setting
// @Summary Rename setting
// @Description Rename a setting and update its path and pathHash
// @Tags Setting
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.SettingRenameRequest true "Rename Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.SettingDTO} "Success"
// @Router /api/setting/rename [post]
func (h *SettingHandler) Rename(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.SettingRenameRequest{}
	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		h.App.Logger().Error("SettingHandler.Rename.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	res, err := h.App.GetSettingService(h.getClientInfo(c)).Rename(c.Request.Context(), uid, params)
	if err != nil {
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(res))

	// Broadcast old path deletion and new path modification
	h.WSS.BroadcastToUser(uid, code.Success.WithData(dto.SettingSyncDeleteMessage{
		Path:             params.OldPath,
		PathHash:         params.OldPathHash,
		Ctime:            res.Ctime,
		Mtime:            res.Mtime,
		UpdatedTimestamp: res.UpdatedTimestamp,
	}).WithVault(params.Vault), string(dto.SettingSyncDelete))

	h.WSS.BroadcastToUser(uid, code.Success.WithData(dto.SettingSyncModifyMessage{
		Vault:            params.Vault,
		Path:             res.Path,
		PathHash:         res.PathHash,
		Content:          res.Content,
		ContentHash:      res.ContentHash,
		Ctime:            res.Ctime,
		Mtime:            res.Mtime,
		UpdatedTimestamp: res.UpdatedTimestamp,
	}).WithVault(params.Vault), string(dto.SettingSyncModify))
}
