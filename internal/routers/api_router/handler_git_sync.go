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

// GitSyncHandler git sync API router handler
type GitSyncHandler struct {
	*Handler
}

// NewGitSyncHandler creates GitSyncHandler instance
func NewGitSyncHandler(a *app.App) *GitSyncHandler {
	return &GitSyncHandler{
		Handler: NewHandler(a),
	}
}

// GetConfigs gets git sync configurations
// @Summary Get git sync configurations
// @Tags GitSync
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Success 200 {object} pkgapp.Res{data=[]dto.GitSyncConfigDTO} "Success"
// @Failure 401 {object} pkgapp.Res "Token Required"
// @Failure 500 {object} pkgapp.Res "Internal Server Error"
// @Router /api/git-sync/configs [get]
func (h *GitSyncHandler) GetConfigs(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	configs, err := h.App.GitSyncService.GetConfigs(c.Request.Context(), uid)
	if err != nil {
		h.logError(c.Request.Context(), "GitSyncHandler.GetConfigs", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(configs))
}

// UpdateConfig updates or creates git sync configuration
// @Summary Update git sync configuration
// @Tags GitSync
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.GitSyncConfigRequest true "Git Sync Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.GitSyncConfigDTO} "Success"
// @Failure 400 {object} pkgapp.Res "Invalid Params"
// @Failure 401 {object} pkgapp.Res "Token Required"
// @Failure 500 {object} pkgapp.Res "Internal Server Error"
// @Router /api/git-sync/config [post]
func (h *GitSyncHandler) UpdateConfig(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.GitSyncConfigRequest{}
	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	config, err := h.App.GitSyncService.UpdateConfig(c.Request.Context(), uid, params)
	if err != nil {
		h.logError(c.Request.Context(), "GitSyncHandler.UpdateConfig", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.SuccessUpdate.WithData(config))
}

// DeleteConfig deletes git sync configuration
// @Summary Delete git sync configuration
// @Tags GitSync
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params body dto.GitSyncDeleteRequest true "Git Sync ID"
// @Success 200 {object} pkgapp.Res "Success"
// @Failure 400 {object} pkgapp.Res "Invalid Params"
// @Failure 401 {object} pkgapp.Res "Token Required"
// @Failure 500 {object} pkgapp.Res "Internal Server Error"
// @Router /api/git-sync/config [delete]
func (h *GitSyncHandler) DeleteConfig(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.GitSyncDeleteRequest{}
	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	err := h.App.GitSyncService.DeleteConfig(c.Request.Context(), uid, params.ID)
	if err != nil {
		h.logError(c.Request.Context(), "GitSyncHandler.DeleteConfig", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success)
}

// Validate validates git sync configuration parameters
// @Summary Validate git sync parameters
// @Tags GitSync
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.GitSyncValidateRequest true "Validation Parameters"
// @Success 200 {object} pkgapp.Res "Success"
// @Failure 400 {object} pkgapp.Res "Invalid Params"
// @Failure 401 {object} pkgapp.Res "Token Required"
// @Failure 500 {object} pkgapp.Res "Internal Server Error"
// @Router /api/git-sync/validate [post]
func (h *GitSyncHandler) Validate(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.GitSyncValidateRequest{}
	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	err := h.App.GitSyncService.Validate(c.Request.Context(), params)
	if err != nil {
		h.logError(c.Request.Context(), "GitSyncHandler.Validate", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithDetails("Validation successful"))
}

// Execute manual sync task
// @Summary Trigger a manual git sync
// @Tags GitSync
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.GitSyncExecuteRequest true "Execute Parameters"
// @Success 200 {object} pkgapp.Res "Success"
// @Failure 400 {object} pkgapp.Res "Invalid Params"
// @Failure 401 {object} pkgapp.Res "Token Required"
// @Failure 500 {object} pkgapp.Res "Internal Server Error"
// @Router /api/git-sync/config/execute [post]
func (h *GitSyncHandler) Execute(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.GitSyncExecuteRequest{}
	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	err := h.App.GitSyncService.ExecuteSync(c.Request.Context(), uid, params.ID)
	if err != nil {
		h.logError(c.Request.Context(), "GitSyncHandler.Execute", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithDetails("Sync started in background"))
}

// CleanWorkspace cleans local git workspace
// @Summary Clean local git workspace
// @Tags GitSync
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.GitSyncCleanRequest true "Clean Parameters"
// @Success 200 {object} pkgapp.Res "Success"
// @Failure 400 {object} pkgapp.Res "Invalid Params"
// @Failure 401 {object} pkgapp.Res "Token Required"
// @Failure 500 {object} pkgapp.Res "Internal Server Error"
// @Router /api/git-sync/config/clean [delete]
func (h *GitSyncHandler) CleanWorkspace(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.GitSyncCleanRequest{}
	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	err := h.App.GitSyncService.CleanWorkspace(c.Request.Context(), uid, params.ConfigID)
	if err != nil {
		h.logError(c.Request.Context(), "GitSyncHandler.CleanWorkspace", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithDetails("Workspace cleaned"))
}

// GetHistories gets git sync histories
// @Summary Get git sync histories
// @Tags GitSync
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params query dto.GitSyncHistoryRequest true "Parameters"
// @Success 200 {object} pkgapp.Res{data=pkgapp.ListRes{list=[]dto.GitSyncHistoryDTO}} "Success"
// @Failure 401 {object} pkgapp.Res "Token Required"
// @Failure 500 {object} pkgapp.Res "Internal Server Error"
// @Router /api/git-sync/histories [get]
func (h *GitSyncHandler) GetHistories(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.GitSyncHistoryRequest{}
	pager := pkgapp.NewPager(c)

	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	list, total, err := h.App.GitSyncService.ListHistory(c.Request.Context(), uid, params.ConfigID, pager)
	if err != nil {
		h.logError(c.Request.Context(), "GitSyncHandler.GetHistories", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponseList(code.Success, list, int(total))
}

func (h *GitSyncHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}
