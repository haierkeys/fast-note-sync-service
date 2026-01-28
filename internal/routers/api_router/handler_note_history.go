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
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
)

// NoteHistoryHandler note history API router handler
// NoteHistoryHandler 笔记历史 API 路由处理器
// Uses App Container to inject dependencies, supports unified error handling
// 使用 App Container 注入依赖，支持统一错误处理
type NoteHistoryHandler struct {
	*Handler
}

// NewNoteHistoryHandler creates NoteHistoryHandler instance
// NewNoteHistoryHandler 创建 NoteHistoryHandler 实例
func NewNoteHistoryHandler(a *app.App, wss *pkgapp.WebsocketServer) *NoteHistoryHandler {
	return &NoteHistoryHandler{
		Handler: NewHandlerWithWSS(a, wss),
	}
}

// NoteHistoryGetRequestParams request parameters for getting note history details
// NoteHistoryGetRequestParams 获取笔记历史详情请求参数
type NoteHistoryGetRequestParams struct {
	ID int64 `form:"id" binding:"required"`
}

// Get retrieves specific note history details
// @Summary Get note history details
// @Description Get specific note history content by history record ID
// @Tags Note History
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param id query int64 true "History Record ID"
// @Success 200 {object} pkgapp.Res{data=dto.NoteHistoryDTO} "Success"
// @Router /api/note/history [get]
func (h *NoteHistoryHandler) Get(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &NoteHistoryGetRequestParams{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHistoryHandler.Get.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHistoryHandler.Get err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	history, err := h.App.NoteHistoryService.Get(ctx, uid, params.ID)
	if err != nil {
		h.logError(ctx, "NoteHistoryHandler.Get", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(history))
}

// List retrieves note history list
// @Summary Get note history list
// @Description Get all history records for a specific note with pagination
// @Tags Note History
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params query dto.NoteHistoryListRequest true "Query Parameters"
// @Success 200 {object} pkgapp.Res{data=[]dto.NoteHistoryDTO} "Success"
// @Router /api/note/histories [get]
func (h *NoteHistoryHandler) List(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteHistoryListRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHistoryHandler.List.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHistoryHandler.List err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	pager := &pkgapp.Pager{Page: pkgapp.GetPage(c), PageSize: pkgapp.GetPageSize(c)}

	list, count, err := h.App.NoteHistoryService.List(ctx, uid, params, pager)
	if err != nil {
		h.logError(ctx, "NoteHistoryHandler.List", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponseList(code.Success, list, int(count))
}

// logError records error log, including Trace ID
// logError 记录错误日志，包含 Trace ID
func (h *NoteHistoryHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}

// Restore restores note content from history
// @Summary Restore note from history
// @Description Restore note content to a specific history version
// @Tags Note History
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.NoteHistoryRestoreRequest true "Restore Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.NoteDTO} "Success"
// @Router /api/note/history/restore [put]
func (h *NoteHistoryHandler) Restore(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteHistoryRestoreRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHistoryHandler.Restore.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHistoryHandler.Restore err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	// Execute restore
	// 执行恢复
	note, err := h.App.NoteHistoryService.RestoreFromHistory(ctx, uid, params.HistoryID)
	if err != nil {
		h.logError(ctx, "NoteHistoryHandler.Restore", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(note).WithVault(params.Vault))
	h.WSS.BroadcastToUser(uid, code.Success.WithData(note).WithVault(params.Vault), "NoteSyncModify")
}
