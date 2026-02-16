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

// StorageHandler configuration API router handler
type StorageHandler struct {
	*Handler
}

// NewStorageHandler creates StorageHandler instance
func NewStorageHandler(a *app.App) *StorageHandler {
	return &StorageHandler{
		Handler: NewHandler(a),
	}
}

// CreateOrUpdate creates or updates storage configuration
// @Summary Create or update storage configuration
// @Tags Storage
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.StoragePostRequest true "Storage Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.StorageDTO} "Success"
// @Router /api/storage [post]
func (h *StorageHandler) CreateOrUpdate(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.StoragePostRequest{}

	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		h.App.Logger().Error("StorageHandler.CreateOrUpdate.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	storage, err := h.App.StorageService.CreateOrUpdate(c.Request.Context(), uid, params.ID, params)
	if err != nil {
		h.logError(c.Request.Context(), "StorageHandler.CreateOrUpdate", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	if params.ID > 0 {
		response.ToResponse(code.SuccessUpdate.WithData(storage))
	} else {
		response.ToResponse(code.SuccessCreate.WithData(storage))
	}
}

// List gets storage configuration list
// @Summary Get storage configuration list
// @Tags Storage
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Success 200 {object} pkgapp.Res{data=[]dto.StorageDTO} "Success"
// @Router /api/storage [get]
func (h *StorageHandler) List(c *gin.Context) {
	response := pkgapp.NewResponse(c)

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	list, err := h.App.StorageService.List(c.Request.Context(), uid)
	if err != nil {
		h.logError(c.Request.Context(), "StorageHandler.List", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(list))
}

// Delete deletes storage configuration
// @Summary Delete storage configuration
// @Tags Storage
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param id query int64 true "Storage ID"
// @Success 200 {object} pkgapp.Res "Success"
// @Router /api/storage [delete]
func (h *StorageHandler) Delete(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.StorageGetRequest{}

	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	err := h.App.StorageService.Delete(c.Request.Context(), uid, params.ID)
	if err != nil {
		h.logError(c.Request.Context(), "StorageHandler.Delete", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.SuccessDelete)
}

// EnabledTypes gets enabled storage types
// @Summary Get enabled storage types
// @Description Get list of enabled storage types. Possible values: localfs, oss, s3, r2, minio, webdav
// @Tags Storage
// @Produce json
// @Success 200 {object} pkgapp.Res{data=[]string} "Success. Data contains: localfs, oss, s3, r2, minio, webdav"
// @Router /api/storage/enabled_types [get]
func (h *StorageHandler) EnabledTypes(c *gin.Context) {
	response := pkgapp.NewResponse(c)

	types, err := h.App.StorageService.GetEnabledTypes()
	if err != nil {
		h.logError(c.Request.Context(), "StorageHandler.EnabledTypes", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(types))
}

func (h *StorageHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}
