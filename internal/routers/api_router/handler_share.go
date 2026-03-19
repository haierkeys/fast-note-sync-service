package api_router

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"go.uber.org/zap"
)

// ShareHandler share API router handler
// ShareHandler 分享 API 路由处理器
type ShareHandler struct {
	*Handler
}

// NewShareHandler creates ShareHandler instance
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

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	ctx := c.Request.Context()

	// Call service layer to generate Token (automatically identify type and resolve associated resources)
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

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get authorization Token
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

	// Parameter binding and validation
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get authorization Token
	token, _ := c.Get("share_token")
	shareToken, _ := token.(string)
	if shareToken == "" {
		response.ToResponse(code.ErrorInvalidAuthToken)
		return
	}

	ctx := c.Request.Context()
	savePath, contentType, mtime, etag, fileName, err := h.App.ShareService.GetSharedFileInfo(ctx, shareToken, params.ID)

	if err != nil {
		if cObj, ok := err.(*code.Code); ok {
			response.ToResponse(cObj)
		} else {
			h.logError(ctx, "ShareHandler.FileGet", err)
			response.ToResponse(code.Failed.WithDetails(err.Error()))
		}
		return
	}

	// Open file for zero-copy serving
	file, err := os.Open(savePath)
	if err != nil {
		h.logError(ctx, "ShareHandler.FileGet.Open", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	defer file.Close()

	// Set response headers and output content
	if contentType != "" {
		c.Header("Content-Type", contentType)
	}
	c.Header("Cache-Control", "public, s-maxage=31536000, max-age=31536000, must-revalidate")
	if etag != "" {
		c.Header("ETag", etag)
	}

	http.ServeContent(c.Writer, c.Request, fileName, time.UnixMilli(mtime), file)
}

// Query queries a share by path
// @Summary Query share by path
// @Description Get share token and info by vault and path
// @Tags Share
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Param params query dto.ShareQueryRequest true "Query Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.ShareCreateResponse} "Success"
// @Router /api/share [get]
func (h *ShareHandler) Query(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.ShareQueryRequest{}

	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	ctx := c.Request.Context()

	share, err := h.App.ShareService.GetShareByPath(ctx, uid, params.Vault, params.PathHash)
	if err != nil {
		if cObj, ok := err.(*code.Code); ok {
			response.ToResponse(cObj)
		} else {
			response.ToResponse(code.Failed.WithDetails(err.Error()))
		}
		return
	}

	// Generate Token again (since it's not stored in DB, we use the SID encryption scheme)
	token, err := h.App.TokenManager.ShareGenerate(share.ID, uid, share.Resources)
	if err != nil {
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}

	// Determine main ID and type for response
	var mainID int64
	var mainType string
	if ids, ok := share.Resources["note"]; ok && len(ids) > 0 {
		mainID, _ = strconv.ParseInt(ids[0], 10, 64)
		mainType = "note"
	} else if ids, ok := share.Resources["file"]; ok && len(ids) > 0 {
		mainID, _ = strconv.ParseInt(ids[0], 10, 64)
		mainType = "file"
	}

	response.ToResponse(code.Success.WithData(&dto.ShareCreateResponse{
		ID:        mainID,
		Type:      mainType,
		Token:     token,
		ExpiresAt: share.ExpiresAt,
	}))
}

// Cancel cancels a share by ID or path
// @Summary Cancel share
// @Description Cancel a share by ID or path parameters
// @Tags Share
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.ShareCancelRequest true "Cancel Parameters"
// @Success 200 {object} pkgapp.Res "Success"
// @Router /api/share [delete]
func (h *ShareHandler) Cancel(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.ShareCancelRequest{}

	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	ctx := c.Request.Context()

	var err error
	if params.ID > 0 {
		err = h.App.ShareService.StopShare(ctx, uid, params.ID)
	} else if params.Vault != "" && params.PathHash != "" {
		err = h.App.ShareService.StopShareByPath(ctx, uid, params.Vault, params.PathHash)
	} else {
		response.ToResponse(code.ErrorInvalidParams.WithDetails("Either ID or Vault + PathHash must be provided"))
		return
	}

	if err != nil {
		if cObj, ok := err.(*code.Code); ok {
			response.ToResponse(cObj)
		} else {
			response.ToResponse(code.Failed.WithDetails(err.Error()))
		}
		return
	}

	response.ToResponse(code.Success)
}

// List lists all shares of a user
// @Summary List shares
// @Description Get all active and inactive shares of the user
// @Tags Share
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Success 200 {object} pkgapp.Res{data=dto.ShareListResponse} "Success"
// @Router /api/shares [get]
func (h *ShareHandler) List(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	uid := pkgapp.GetUID(c)
	ctx := c.Request.Context()

	items, err := h.App.ShareService.ListSharesWithDetail(ctx, uid)
	if err != nil {
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}

	response.ToResponse(code.Success.WithData(&dto.ShareListResponse{Items: items}))
}

// logError records error log, including Trace ID
// logError 记录错误日志，包含 Trace ID
func (h *ShareHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}
