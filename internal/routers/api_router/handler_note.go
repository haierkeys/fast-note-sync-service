package api_router

import (
	"bytes"
	"context"
	"net/http"
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

// List 获取笔记列表
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

// GetFileContent 获取文件或笔记的原始内容
func (h *NoteHandler) GetFileContent(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteGetRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.GetFileContent.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.GetFileContent err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// 计算 PathHash
	params.PathHash = util.EncodeHash32(params.Path)

	// 获取请求上下文
	ctx := c.Request.Context()


	noteSvc := h.App.GetNoteService(app.WebClientName, "")
	content, contentType, mtime, etag, err := noteSvc.GetFileContent(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "NoteHandler.GetFileContent", err)
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}

	// 如果内容为 nil, 表示资源未找到或已删除, 返回 404
	if content == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// 设置响应头
	if contentType != "" {
		c.Header("Content-Type", contentType)
	}
	c.Header("Cache-Control", "public, s-maxage=31536000, max-age=31536000, must-revalidate")
	if etag != "" {
		c.Header("ETag", etag)
	}

	http.ServeContent(c.Writer, c.Request, params.Path, time.UnixMilli(mtime), bytes.NewReader(content))
}

// logError 记录错误日志，包含 Trace ID
func (h *NoteHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}
