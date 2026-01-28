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

// Create creates a share
// @Summary Create resource share
// @Description Create a share token for a specific note or attachment, automatically resolve attachment references and authorize
// @Tags Share
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.ShareCreateRequest true "Share Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.ShareCreateResponse} "Success"
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

// GetShared retrieves shared note details
// @Summary Get shared note details
// @Description Get specific note content (restricted read-only access) via share token
// @Tags Share
// @Security ShareAuthToken
// @Param Share-Token header string true "Auth Token"
// @Produce json
// @Param params query dto.ShareResourceRequest true "Get Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.NoteDTO} "Success"
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

	// 获取授权 Token
	token, _ := c.Get("share_token")
	shareToken, _ := token.(string)
	if shareToken == "" {
		response.ToResponse(code.ErrorInvalidAuthToken)
		return
	}

	ctx := c.Request.Context()
	noteDTO, err := h.App.ShareService.GetSharedNote(ctx, shareToken, params.ID)
	if err != nil {
		if cObj, ok := err.(*code.Code); ok {
			response.ToResponse(cObj)
		} else {
			h.logError(ctx, "ShareHandler.NoteGet", err)
			response.ToResponse(code.Failed.WithDetails(err.Error()))
		}
		return
	}

	response.ToResponse(code.Success.WithData(noteDTO))
}

// GetSharedContent retrieves shared file content
// @Summary Get shared attachment content
// @Description Get raw binary data of a specific attachment via share token
// @Tags Share
// @Security ShareAuthToken
// @Param Share-Token header string true "Auth Token"
// @Produce octet-stream
// @Param params query dto.ShareResourceRequest true "Get Parameters"
// @Success 200 {file} binary "Success"
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

	// 获取授权 Token
	token, _ := c.Get("share_token")
	shareToken, _ := token.(string)
	if shareToken == "" {
		response.ToResponse(code.ErrorInvalidAuthToken)
		return
	}

	ctx := c.Request.Context()
	content, contentType, mtime, etag, fileName, err := h.App.ShareService.GetSharedFile(ctx, shareToken, params.ID)

	if err != nil {
		if cObj, ok := err.(*code.Code); ok {
			response.ToResponse(cObj)
		} else {
			h.logError(ctx, "ShareHandler.FileGet", err)
			response.ToResponse(code.Failed.WithDetails(err.Error()))
		}
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

	http.ServeContent(c.Writer, c.Request, fileName, time.UnixMilli(mtime), bytes.NewReader(content))
}

// logError 记录错误日志，包含 Trace ID
func (h *ShareHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}
