package api_router

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gookit/goutil/dump"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/middleware"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	apperrors "github.com/haierkeys/fast-note-sync-service/pkg/errors"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
)

// FileHandler 文件 API 路由处理器
type FileHandler struct {
	*Handler
}

// NewFileHandler 创建 FileHandler 实例
func NewFileHandler(a *app.App, wss *pkgapp.WebsocketServer) *FileHandler {
	return &FileHandler{
		Handler: NewHandlerWithWSS(a, wss),
	}
}

// List 获取文件列表
// @Summary 获取文件列表
// @Description 分页并支持搜索、过滤、排序地获取当前用户的附件列表
// @Tags 附件
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Produce json
// @Param params query dto.FileListRequest true "查询参数"
// @Success 200 {object} pkgapp.Res{data=[]dto.FileDTO} "成功"
// @Router /api/files [get]
func (h *FileHandler) List(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.FileListRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("FileHandler.List.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("FileHandler.List err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// 获取请求上下文
	ctx := c.Request.Context()

	pager := &pkgapp.Pager{Page: pkgapp.GetPage(c), PageSize: pkgapp.GetPageSize(c)}
	fileSvc := h.App.GetFileService(app.WebClientName, "")
	files, count, err := fileSvc.List(ctx, uid, params, pager)
	if err != nil {
		h.logError(ctx, "FileHandler.List", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponseList(code.Success, files, count)
}

// GetContent 获取文件或笔记的原始内容
// @Summary 获取附件内容
// @Description 根据路径获取附件的原始二进制数据，支持强缓存控制
// @Tags 附件
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Produce octet-stream
// @Param params query dto.FileGetRequest true "获取参数"
// @Success 200 {file} binary "成功"
// @Router /api/file [get]
func (h *FileHandler) GetContent(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.FileGetRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("FileHandler.GetContent.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("FileHandler.GetContent err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// 计算 PathHash
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// 获取请求上下文
	ctx := c.Request.Context()

	fileSvc := h.App.GetFileService(app.WebClientName, "")
	content, contentType, mtime, etag, err := fileSvc.GetContent(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "FileHandler.GetContent", err)
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

// GetSharedContent 获取分享的文件内容
// @Summary 获取被分享的附件内容
// @Description 通过分享 Token 授权后，获取特定附件的原始二进制数据
// @Tags 附件
// @Produce octet-stream
// @Success 200 {file} binary "成功"
// @Router /api/share/file [get]

// Delete 删除文件
// @Summary 删除附件
// @Description 永久删除指定的附件记录及其物理文件
// @Tags 附件
// @Security UserAuthToken
// @Param token header string true "认证 Token"
// @Produce json
// @Param params query dto.FileDeleteRequest true "删除参数"
// @Success 200 {object} pkgapp.Res{data=dto.FileDTO} "成功"
// @Router /api/file [delete]
func (h *FileHandler) Delete(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.FileDeleteRequest{}

	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)

	dump.P(params)
	if !valid {
		h.App.Logger().Error("FileHandler.Delete.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("FileHandler.Delete err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// 计算 PathHash
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// 获取请求上下文
	ctx := c.Request.Context()

	fileSvc := h.App.GetFileService(app.WebClientName, "")
	// 执行删除
	file, err := fileSvc.Delete(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "FileHandler.Delete", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(file))
	fileDeleteMessage := &dto.FileDeleteMessage{Path: file.Path}
	h.WSS.BroadcastToUser(uid, code.Success.WithData(fileDeleteMessage).WithVault(params.Vault), "FileSyncDelete")
}

// logError 记录错误日志，包含 Trace ID
func (h *FileHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}
