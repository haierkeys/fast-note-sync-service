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

// NoteHistoryHandler 笔记历史 API 路由处理器
// 使用 App Container 注入依赖，支持统一错误处理
type NoteHistoryHandler struct {
	*Handler
}

// NewNoteHistoryHandler 创建 NoteHistoryHandler 实例
func NewNoteHistoryHandler(a *app.App, wss *pkgapp.WebsocketServer) *NoteHistoryHandler {
	return &NoteHistoryHandler{
		Handler: NewHandlerWithWSS(a, wss),
	}
}

// NoteHistoryGetRequestParams 获取笔记历史详情请求参数
type NoteHistoryGetRequestParams struct {
	ID int64 `form:"id" binding:"required"`
}

// Get 获取单条笔记历史详情
// @Summary 获取笔记历史详情
// @Description 根据历史记录 ID 获取单条特定的笔记历史内容
// @Tags 笔记历史
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Produce json
// @Param id query int64 true "历史记录 ID"
// @Success 200 {object} pkgapp.Res{data=dto.NoteHistoryDTO} "成功"
// @Router /api/note/history [get]
func (h *NoteHistoryHandler) Get(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &NoteHistoryGetRequestParams{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHistoryHandler.Get.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHistoryHandler.Get err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

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

// List 获取笔记历史列表
// @Summary 获取笔记历史列表
// @Description 分页获取特定笔记的所有历史修改记录
// @Tags 笔记历史
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Produce json
// @Param params query dto.NoteHistoryListRequest true "查询参数"
// @Success 200 {object} pkgapp.Res{data=[]dto.NoteHistoryDTO} "成功"
// @Router /api/note/histories [get]
func (h *NoteHistoryHandler) List(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteHistoryListRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHistoryHandler.List.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHistoryHandler.List err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

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

// logError 记录错误日志，包含 Trace ID
func (h *NoteHistoryHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}

// Restore 从历史版本恢复笔记内容
// @Summary 从历史版本恢复笔记
// @Description 将笔记内容恢复到指定的历史版本
// @Tags 笔记历史
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Accept json
// @Produce json
// @Param params body dto.NoteHistoryRestoreRequest true "恢复参数"
// @Success 200 {object} pkgapp.Res{data=dto.NoteDTO} "成功"
// @Router /api/note/history/restore [put]
func (h *NoteHistoryHandler) Restore(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteHistoryRestoreRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHistoryHandler.Restore.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHistoryHandler.Restore err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// 获取请求上下文
	ctx := c.Request.Context()

	// 执行恢复
	note, err := h.App.NoteHistoryService.RestoreFromHistory(ctx, uid, params.HistoryID)
	if err != nil {
		h.logError(ctx, "NoteHistoryHandler.Restore", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(note).WithVault(params.Vault))

	// 广播恢复事件到其他客户端
	h.WSS.BroadcastToUser(uid, code.Success.WithData(note).WithVault(params.Vault), "NoteSyncHistoryRestore")
}
