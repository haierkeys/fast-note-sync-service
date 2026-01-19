package api_router

import (
	"context"
	"time"

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

// NoteHandler 笔记 API 路由处理器
// 使用 App Container 注入依赖，支持统一错误处理
type NoteHandler struct {
	*Handler
}

// NewNoteHandler 创建 NoteHandler 实例
func NewNoteHandler(a *app.App, wss *pkgapp.WebsocketServer) *NoteHandler {
	return &NoteHandler{
		Handler: NewHandlerWithWSS(a, wss),
	}
}

// Get 获取单条笔记详情
// @Summary 获取笔记详情
// @Description 根据路径或路径哈希获取单条笔记的具体内容和元数据
// @Tags 笔记
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Produce json
// @Param params query dto.NoteGetRequest true "获取参数"
// @Success 200 {object} pkgapp.Res{data=dto.NoteWithFileLinksResponse} "成功"
// @Router /api/note [get]
func (h *NoteHandler) Get(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteGetRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.Get.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.Get err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// 计算 PathHash
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// 获取请求上下文
	ctx := c.Request.Context()

	noteSvc := h.App.GetNoteService(app.WebClientName, "")
	note, err := noteSvc.Get(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "NoteHandler.Get", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	// 解析内容中的 ![[ ]] 标签
	fileLinks, err := h.App.FileService.ResolveEmbedLinks(ctx, uid, params.Vault, note.Content)
	if err != nil {
		h.App.Logger().Error("NoteHandler.Get FileResolveEmbedLinks err", zap.Error(err))
	}

	noteWithLinks := &dto.NoteWithFileLinksResponse{
		ID:               note.ID,
		Path:             note.Path,
		PathHash:         note.PathHash,
		Content:          note.Content,
		ContentHash:      note.ContentHash,
		FileLinks:        fileLinks,
		Version:          note.Version,
		Ctime:            note.Ctime,
		Mtime:            note.Mtime,
		UpdatedTimestamp: note.UpdatedTimestamp,
		UpdatedAt:        note.UpdatedAt,
		CreatedAt:        note.CreatedAt,
	}

	response.ToResponse(code.Success.WithData(noteWithLinks))
}

// GetShared 获取分享的单条笔记详情
// @Summary 获取被分享的笔记详情
// @Description 通过分享 Token 授权后，获取特定笔记内容（受限只读访问）
// @Tags 笔记
// @Security ShareAuthToken
// @Param Share-Token header string true "认证 Token"
// @Produce json
// @Success 200 {object} pkgapp.Res{data=dto.NoteDTO} "成功"
// @Router /api/share/note [get]

// List 获取笔记列表
// @Summary 获取笔记列表
// @Description 分页获取当前用户的笔记列表
// @Tags 笔记
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Produce json
// @Param params query dto.NoteListRequest true "查询参数"
// @Success 200 {object} pkgapp.Res{data=[]dto.NoteDTO} "成功"
// @Router /api/notes [get]
func (h *NoteHandler) List(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteListRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.List.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.List err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// 获取请求上下文
	ctx := c.Request.Context()

	noteSvc := h.App.GetNoteService(app.WebClientName, "")
	pager := &pkgapp.Pager{Page: pkgapp.GetPage(c), PageSize: pkgapp.GetPageSize(c)}

	notes, count, err := noteSvc.List(ctx, uid, params, pager)
	if err != nil {
		h.logError(ctx, "NoteHandler.List", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponseList(code.Success, notes, count)
}

// CreateOrUpdate 创建或更新笔记
// @Summary 创建或更新笔记
// @Description 处理笔记的新增、修改或重命名（通过路径变化识别）
// @Tags 笔记
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Accept json
// @Produce json
// @Param params body dto.NoteModifyOrCreateRequest true "笔记内容"
// @Success 200 {object} pkgapp.Res{data=dto.NoteDTO} "成功"
// @Router /api/note [post]
func (h *NoteHandler) CreateOrUpdate(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteModifyOrCreateRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.CreateOrUpdate.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.CreateOrUpdate err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// 计算哈希值
	if params.SrcPathHash == "" {
		params.SrcPathHash = util.EncodeHash32(params.SrcPath)
	}
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}
	if params.ContentHash == "" {
		params.ContentHash = util.EncodeHash32(params.Content)
	}
	if params.Mtime == 0 {
		params.Mtime = time.Now().UnixMilli()
	}
	if params.Ctime == 0 {
		params.Ctime = params.Mtime
	}

	// 获取请求上下文
	ctx := c.Request.Context()

	noteSvc := h.App.GetNoteService(app.WebClientName, "")

	// 处理重命名场景
	if params.SrcPath != "" && params.SrcPath != params.Path {
		noteSrc, err := noteSvc.Get(ctx, uid, &dto.NoteGetRequest{
			Vault:    params.Vault,
			Path:     params.SrcPath,
			PathHash: params.SrcPathHash,
		})
		if err != nil {
			h.logError(ctx, "NoteHandler.CreateOrUpdate.NoteGet", err)
			apperrors.ErrorResponse(c, err)
			return
		}
		if noteSrc == nil || noteSrc.Action == "delete" {
			response.ToResponse(code.ErrorNoteNotFound)
			return
		}
	}

	// 检查更新
	checkParams := &dto.NoteUpdateCheckRequest{
		Vault:       params.Vault,
		Path:        params.Path,
		PathHash:    params.PathHash,
		ContentHash: params.ContentHash,
		Ctime:       params.Ctime,
		Mtime:       params.Mtime,
	}
	_, noteSelect, err := noteSvc.UpdateCheck(ctx, uid, checkParams)
	if err != nil {
		h.logError(ctx, "NoteHandler.CreateOrUpdate.NoteUpdateCheck", err)
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}

	if noteSelect != nil {
		if noteSelect.Action != "delete" && params.SrcPath != "" && params.SrcPathHash != params.PathHash {
			response.ToResponse(code.ErrorRenameNoteTargetExist)
			return
		}
		if params.ContentHash != noteSelect.ContentHash {
			params.Mtime = time.Now().UnixMilli()
		}
	}

	var noteNew *dto.NoteDTO
	var noteOld *dto.NoteDTO

	// 如果路径发生变化，删除旧笔记
	if params.SrcPath != "" && params.SrcPath != params.Path {
		deleteParams := &dto.NoteDeleteRequest{
			Vault:    params.Vault,
			Path:     params.SrcPath,
			PathHash: params.SrcPathHash,
		}
		noteOld, err = noteSvc.Delete(ctx, uid, deleteParams)
		if err != nil {
			h.logError(ctx, "NoteHandler.CreateOrUpdate.NoteDelete", err)
			apperrors.ErrorResponse(c, err)
			return
		}
		h.WSS.BroadcastToUser(uid, code.Success.WithData(noteOld).WithVault(params.Vault), "NoteSyncDelete")
	}

	_, noteNew, err = noteSvc.ModifyOrCreate(ctx, uid, params, false)
	if err != nil {
		h.logError(ctx, "NoteHandler.CreateOrUpdate.NoteModifyOrCreate", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(noteNew))
	h.WSS.BroadcastToUser(uid, code.Success.WithData(noteNew).WithVault(params.Vault), "NoteSyncModify")

	if params.SrcPath != "" && params.SrcPath != params.Path {
		noteSvc.MigratePush(noteOld.ID, noteNew.ID, uid)
	}
}

// Delete 删除笔记
// @Summary 删除笔记
// @Description 将笔记移至回收站
// @Tags 笔记
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Produce json
// @Param params query dto.NoteDeleteRequest true "删除参数"
// @Success 200 {object} pkgapp.Res{data=dto.NoteDTO} "成功"
// @Router /api/note [delete]
func (h *NoteHandler) Delete(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteDeleteRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.Delete.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.Delete err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// 计算 PathHash
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// 获取请求上下文
	ctx := c.Request.Context()

	noteSvc := h.App.GetNoteService(app.WebClientName, "")

	// 检查笔记是否存在
	noteSrc, err := noteSvc.Get(ctx, uid, &dto.NoteGetRequest{
		Vault:    params.Vault,
		Path:     params.Path,
		PathHash: params.PathHash,
	})
	if err != nil {
		h.logError(ctx, "NoteHandler.Delete.NoteGet", err)
		apperrors.ErrorResponse(c, err)
		return
	}
	if noteSrc == nil || noteSrc.Action == "delete" {
		response.ToResponse(code.ErrorNoteNotFound)
		return
	}

	// 执行删除
	note, err := noteSvc.Delete(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "NoteHandler.Delete.NoteDelete", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(note))
	h.WSS.BroadcastToUser(uid, code.Success.WithData(note).WithVault(params.Vault), "NoteSyncDelete")
}

// Restore 恢复笔记（从回收站恢复）
// @Summary 恢复笔记
// @Description 从回收站恢复被删除的笔记
// @Tags 笔记
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Produce json
// @Param params body dto.NoteRestoreRequest true "恢复参数"
// @Success 200 {object} pkgapp.Res{data=dto.NoteDTO} "成功"
// @Router /api/note/restore [put]
func (h *NoteHandler) Restore(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteRestoreRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.Restore.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.Restore err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// 计算 PathHash
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// 获取请求上下文
	ctx := c.Request.Context()

	noteSvc := h.App.GetNoteService(app.WebClientName, "")

	// 检查笔记是否存在于回收站
	noteSrc, err := noteSvc.Get(ctx, uid, &dto.NoteGetRequest{
		Vault:     params.Vault,
		Path:      params.Path,
		PathHash:  params.PathHash,
		IsRecycle: true,
	})
	if err != nil {
		h.logError(ctx, "NoteHandler.Restore.NoteGet", err)
		apperrors.ErrorResponse(c, err)
		return
	}
	if noteSrc == nil || noteSrc.Action != "delete" {
		response.ToResponse(code.ErrorNoteNotFound)
		return
	}

	// 执行恢复
	note, err := noteSvc.Restore(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "NoteHandler.Restore.NoteRestore", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(note))
	h.WSS.BroadcastToUser(uid, code.Success.WithData(note).WithVault(params.Vault), "NoteSyncRestore")
}

// logError 记录错误日志，包含 Trace ID
func (h *NoteHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}
