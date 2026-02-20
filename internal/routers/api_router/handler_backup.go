package api_router

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/gookit/goutil/dump"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	apperrors "github.com/haierkeys/fast-note-sync-service/pkg/errors"
	"go.uber.org/zap"
)

// BackupHandler backup API router handler
type BackupHandler struct {
	*Handler
}

// NewBackupHandler creates BackupHandler instance
func NewBackupHandler(a *app.App) *BackupHandler {
	return &BackupHandler{
		Handler: NewHandler(a),
	}
}

// GetConfigs gets backup configurations
// @Summary Get backup configurations
// @Tags Backup
// @Security UserAuthToken
// @Produce json
// @Success 200 {object} pkgapp.Res{data=[]dto.BackupConfigDTO} "Success"
// @Failure 401 {object} pkgapp.Res "Token Required"
// @Failure 500 {object} pkgapp.Res "Internal Server Error"
// @Router /api/backup/configs [get]
func (h *BackupHandler) GetConfigs(c *gin.Context) {
	response := pkgapp.NewResponse(c)

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	configs, err := h.App.BackupService.GetConfigs(c.Request.Context(), uid)
	if err != nil {
		h.logError(c.Request.Context(), "BackupHandler.GetConfigs", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(configs))
}

// UpdateConfig updates backup configuration
// @Summary Update backup configuration
// @Tags Backup
// @Security UserAuthToken
// @Accept json
// @Produce json
// @Param params body dto.BackupConfigRequest true "Backup Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.BackupConfigDTO} "Success"
// @Failure 400 {object} pkgapp.Res "Invalid Params"
// @Failure 401 {object} pkgapp.Res "Token Required"
// @Failure 500 {object} pkgapp.Res "Internal Server Error"
// @Router /api/backup/config [post]
func (h *BackupHandler) UpdateConfig(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.BackupConfigRequest{}

	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	config, err := h.App.BackupService.UpdateConfig(c.Request.Context(), uid, params)
	if err != nil {
		h.logError(c.Request.Context(), "BackupHandler.UpdateConfig", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.SuccessUpdate.WithData(config))
}

// DeleteConfig deletes backup configuration
// @Summary Delete backup configuration
// @Tags Backup
// @Security UserAuthToken
// @Produce json
// @Param params query dto.BackupExecuteRequest true "Config ID"
// @Success 200 {object} pkgapp.Res "Success"
// @Failure 400 {object} pkgapp.Res "Invalid Params"
// @Failure 401 {object} pkgapp.Res "Token Required"
// @Failure 500 {object} pkgapp.Res "Internal Server Error"
// @Router /api/backup/config [delete]
func (h *BackupHandler) DeleteConfig(c *gin.Context) {

	response := pkgapp.NewResponse(c)
	params := &dto.BackupExecuteRequest{} // Reusing ID struct
	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	dump.P(params)

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	err := h.App.BackupService.DeleteConfig(c.Request.Context(), uid, params.ID)
	if err != nil {
		h.logError(c.Request.Context(), "BackupHandler.DeleteConfig", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success)
}

// ListHistory gets backup history list
// @Summary Get backup history list
// @Tags Backup
// @Security UserAuthToken
// @Produce json
// @Param params query dto.BackupHistoryListRequest true "Backup History List Parameters"
// @Success 200 {object} pkgapp.Res{data=pkgapp.ListRes{list=[]dto.BackupHistoryDTO}} "Success"
// @Failure 400 {object} pkgapp.Res "Invalid Params"
// @Failure 401 {object} pkgapp.Res "Token Required"
// @Failure 500 {object} pkgapp.Res "Internal Server Error"
// @Router /api/backup/historys [get]
func (h *BackupHandler) ListHistory(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.BackupHistoryListRequest{}
	pager := pkgapp.NewPager(c)

	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Override pager with params if provided (though NewPager usually handles page/pageSize from query)
	// But BindAndValid parses them into params.
	// Let's sync them or just rely on params.
	// Actually pkgapp.NewPager extracts page and page_size from context query params.
	// Since we bind them to struct as well, we can just ensure consistency.
	// The service uses pager.Page and pager.PageSize.
	// Let's just use pager as is, because BindAndValid also reads from query.

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	list, total, err := h.App.BackupService.ListHistory(c.Request.Context(), uid, params.ConfigID, pager)
	if err != nil {
		h.logError(c.Request.Context(), "BackupHandler.ListHistory", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponseList(code.Success, list, int(total))
}

// Execute triggers a backup manually
// @Summary Trigger a backup manually
// @Tags Backup
// @Security UserAuthToken
// @Produce json
// @Param params body dto.BackupExecuteRequest true "Backup Execute Parameters"
// @Success 200 {object} pkgapp.Res "Success"
// @Failure 400 {object} pkgapp.Res "Invalid Params"
// @Failure 401 {object} pkgapp.Res "Token Required"
// @Failure 500 {object} pkgapp.Res "Internal Server Error"
// @Router /api/backup/execute [post]
func (h *BackupHandler) Execute(c *gin.Context) {

	response := pkgapp.NewResponse(c)
	params := &dto.BackupExecuteRequest{}
	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	err := h.App.BackupService.ExecuteUserBackup(c.Request.Context(), uid, params.ID)
	if err != nil {
		h.logError(c.Request.Context(), "BackupHandler.Execute", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithDetails("Backup started in background"))
}

func (h *BackupHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}
