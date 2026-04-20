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

// SyncLogHandler sync log API router handler
// SyncLogHandler 同步日志 API 路由处理器
type SyncLogHandler struct {
	*Handler
}

// NewSyncLogHandler creates SyncLogHandler instance
// NewSyncLogHandler 创建 SyncLogHandler 实例
func NewSyncLogHandler(a *app.App) *SyncLogHandler {
	return &SyncLogHandler{
		Handler: NewHandler(a),
	}
}

// List retrieves sync log list with pagination
// @Summary Get sync log list
// @Description Get sync log list for current user with optional type/action filters and pagination
// @Tags Sync Log
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params query dto.SyncLogListRequest true "Query Parameters"
// @Param pagination query pkgapp.PaginationRequest true "Pagination Parameters"
// @Success 200 {object} pkgapp.Res{data=pkgapp.ListRes{list=[]dto.SyncLogDTO}} "Success"
// @Router /api/sync-logs [get]
func (h *SyncLogHandler) List(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.SyncLogListRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("SyncLogHandler.List.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("SyncLogHandler.List err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	// Get VaultID by vault name if provided (0 means no vault scope filter)
	// 如果传入 vault 名称则解析 VaultID（0 表示不限 vault）
	var vaultID int64
	if params.Vault != "" {
		var err2 error
		vaultID, err2 = h.App.VaultService.MustGetID(ctx, uid, params.Vault)
		if err2 != nil {
			h.syncLogErr(ctx, "SyncLogHandler.List.VaultService.MustGetID", err2)
			apperrors.ErrorResponse(c, err2)
			return
		}
	}

	pager := pkgapp.NewPager(c)

	list, total, err := h.App.SyncLogService.List(ctx, uid, vaultID, params.Type, params.Action, pager.Page, pager.PageSize)
	if err != nil {
		h.syncLogErr(ctx, "SyncLogHandler.List", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponseList(code.Success, list, int(total))
}

// syncLogErr records error log
// syncLogErr 记录错误日志
func (h *SyncLogHandler) syncLogErr(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}
