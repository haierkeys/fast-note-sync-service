package api_router

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"go.uber.org/zap"
)

// ShareHandler 分享 API 路由处理器
type ShareHandler struct {
	*Handler
}

// NewShareHandler 创建 ShareHandler 实例
func NewShareHandler(app *app.App) *ShareHandler {
	return &ShareHandler{
		Handler: &Handler{App: app},
	}
}

// Create 创建分享
// @Summary 创建资源分享
// @Description 为指定的笔记或附件创建分享令牌，自动解析笔记中的附件引用并授权
// @Tags 分享
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Accept json
// @Produce json
// @Param params body dto.ShareCreateRequest true "分享参数"
// @Success 200 {object} pkgapp.Res{data=dto.ShareCreateResponse} "成功"
// @Router /api/share [post]
func (h *ShareHandler) Create(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.ShareCreateRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	ctx := c.Request.Context()

	// 调用服务层生成 Token (自动识别类型及解析关联资源)
	shareRes, err := h.App.ShareService.ShareGenerate(ctx, uid, params.Vault, params.Path, params.PathHash)
	if err != nil {
		if cObj, ok := err.(*code.Code); ok {
			response.ToResponse(cObj)
		} else {
			response.ToResponse(code.Failed.WithDetails(err.Error()))
		}
		return
	}

	response.ToResponse(code.Success.WithData(shareRes))
}

// GetShared 获取分享的单条笔记详情
// @Summary 获取被分享的笔记详情
// @Description 通过分享 Token 授权后，获取特定笔记内容（受限只读访问）
// @Tags 分享
// @Security ShareAuthToken
// @Param Share-Token header string true "认证 Token"
// @Produce json
// @Param params query dto.ShareResourceRequest true "获取参数"
// @Success 200 {object} pkgapp.Res{data=dto.NoteDTO} "成功"
// @Router /api/share/note [get]
func (h *ShareHandler) NoteGet(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.ShareResourceRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取授权信息
	shareEntity := pkgapp.GetShareEntity(c)
	if shareEntity == nil {
		response.ToResponse(code.ErrorInvalidAuthToken)
		return
	}
	ctx := c.Request.Context()

	// 获取分享信息以确认资源归属和权限
	share, err := h.App.ShareRepo.GetByID(ctx, shareEntity.SID)
	if err != nil {
		h.logError(ctx, "ShareHandler.GetShared.GetShare", err)
		response.ToResponse(code.ErrorShareNotFound)
		return
	}

	// 验证资源 ID 是否在授权列表中
	ids, ok := share.Resources["note"]
	ridStr := strconv.FormatInt(params.ID, 10)
	authorized := false
	if ok {
		for _, id := range ids {
			if id == ridStr {
				authorized = true
				break
			}
		}
	}

	if !authorized {
		response.ToResponse(code.ErrorInvalidAuthToken)
		return
	}

	// 直接通过 ID 获取笔记 (使用资源所有者的 UID)
	note, err := h.App.NoteRepo.GetByID(ctx, params.ID, share.UID)
	if err != nil {
		h.logError(ctx, "ShareHandler.GetShared.GetNote", err)
		response.ToResponse(code.ErrorNoteNotFound)
		return
	}

	noteDTO := &dto.NoteDTO{
		ID:               note.ID,
		Path:             note.Path,
		Content:          note.Content,
		ContentHash:      note.ContentHash,
		Version:          note.Version,
		Ctime:            note.Ctime,
		Mtime:            note.Mtime,
		UpdatedTimestamp: note.UpdatedTimestamp,
		UpdatedAt:        timex.Time(note.UpdatedAt),
		CreatedAt:        timex.Time(note.CreatedAt),
	}

	response.ToResponse(code.Success.WithData(noteDTO))
}

// GetSharedContent 获取分享的文件内容
// @Summary 获取分享的附件内容
// @Description 通过分享 Token 授权后，获取特定附件的原始二进制数据
// @Tags 分享
// @Security ShareAuthToken
// @Param Share-Token header string true "认证 Token"
// @Produce octet-stream
// @Param params query dto.ShareResourceRequest true "获取参数"
// @Success 200 {file} binary "成功"
// @Router /api/share/file [get]
func (h *ShareHandler) FileGet(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.ShareResourceRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取授权信息
	shareEntity := pkgapp.GetShareEntity(c)
	if shareEntity == nil {
		response.ToResponse(code.ErrorInvalidAuthToken)
		return
	}
	ctx := c.Request.Context()

	// 获取分享信息以确认资源归属和权限
	share, err := h.App.ShareRepo.GetByID(ctx, shareEntity.SID)
	if err != nil {
		h.logError(ctx, "ShareHandler.GetSharedContent.GetShare", err)
		response.ToResponse(code.ErrorShareNotFound)
		return
	}

	// 验证资源 ID 是否在授权列表中
	ids, ok := share.Resources["file"]
	ridStr := strconv.FormatInt(params.ID, 10)
	authorized := false
	if ok {
		for _, id := range ids {
			if id == ridStr {
				authorized = true
				break
			}
		}
	}

	if !authorized {
		response.ToResponse(code.ErrorInvalidAuthToken)
		return
	}

	// 1. 先通过 ID 从 Repo 获取文件元数据 (使用资源所有者的 UID)
	file, err := h.App.FileRepo.GetByID(ctx, params.ID, share.UID)
	if err != nil {
		h.logError(ctx, "ShareHandler.GetSharedContent.GetFile", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// 2. 调用 Service 获取实际内容
	fileSvc := h.App.GetFileService(app.WebClientName, "")
	content, contentType, mtime, etag, err := fileSvc.GetContent(ctx, share.UID, &dto.FileGetRequest{
		Vault:    "", // 对于 ID 查询，Vault 为空
		Path:     file.Path,
		PathHash: file.PathHash,
	})

	if err != nil {
		h.logError(ctx, "ShareHandler.GetSharedContent.Svc", err)
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}

	if content == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// 设置响应头并输出内容
	if contentType != "" {
		c.Header("Content-Type", contentType)
	}
	c.Header("Cache-Control", "public, s-maxage=31536000, max-age=31536000, must-revalidate")
	if etag != "" {
		c.Header("ETag", etag)
	}

	http.ServeContent(c.Writer, c.Request, file.Path, time.UnixMilli(mtime), bytes.NewReader(content))
}

// logError 记录错误日志，包含 Trace ID
func (h *ShareHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}
