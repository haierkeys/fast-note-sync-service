package api_router

import (
	"context"
	"net/http"
	"os"
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

// FileHandler file API router handler
// FileHandler 文件 API 路由处理器
type FileHandler struct {
	*Handler
}

// NewFileHandler creates FileHandler instance
// NewFileHandler 创建 FileHandler 实例
func NewFileHandler(a *app.App, wss *pkgapp.WebsocketServer) *FileHandler {
	return &FileHandler{
		Handler: NewHandlerWithWSS(a, wss),
	}
}

// List retrieves file list
// @Summary Get file list
// @Description Get attachment list for current user with pagination, search, filter, and sort support
// @Tags File
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params query dto.FileListRequest true "Query Parameters"
// @Param pagination query pkgapp.PaginationRequest true "Pagination Parameters"
// @Success 200 {object} pkgapp.Res{data=pkgapp.ListRes{list=[]dto.FileDTO}} "Success"
// @Router /api/files [get]
func (h *FileHandler) List(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.FileListRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("FileHandler.List.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("FileHandler.List err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	pager := pkgapp.NewPager(c)
	fileSvc := h.App.GetFileService(app.WebClientName, "")
	files, count, err := fileSvc.List(ctx, uid, params, pager)
	if err != nil {
		h.logError(ctx, "FileHandler.List", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponseList(code.Success, files, count)
}

// GetInfo retrieves raw content of file or note
// @Summary Get attachment content
// @Description Get raw binary data of an attachment by path, supports strong cache control
// @Tags File
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce octet-stream
// @Param params query dto.FileGetRequest true "Get Parameters"
// @Success 200 {file} binary "Success"
// @Router /api/file [get]
func (h *FileHandler) GetInfo(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.FileGetRequest{}

	// Parameter binding and validation
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("FileHandler.GetContent.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("FileHandler.GetContent err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Calculate PathHash
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// Get request context
	ctx := c.Request.Context()

	fileSvc := h.App.GetFileService(app.WebClientName, "")
	savePath, contentType, mtime, etag, fileName, err := fileSvc.GetContentInfo(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "FileHandler.GetContent", err)
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}

	// Open file for zero-copy serving
	file, err := os.Open(savePath)
	if err != nil {
		h.logError(ctx, "FileHandler.GetContent.Open", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	defer file.Close()

	// Set response headers
	if contentType != "" {
		c.Header("Content-Type", contentType)
	}
	c.Header("Cache-Control", "public, s-maxage=31536000, max-age=31536000, must-revalidate")
	if etag != "" {
		c.Header("ETag", etag)
	}

	http.ServeContent(c.Writer, c.Request, fileName, time.UnixMilli(mtime), file)
}

// GetSharedContent retrieves shared file content
// @Summary Get shared attachment content
// @Description Get raw binary data of a specific attachment via share token
// @Tags File
// @Produce octet-stream
// @Success 200 {file} binary "Success"
// @Router /api/share/file [get]

// Delete deletes a file
// @Summary Delete attachment
// @Description Permanently delete a specific attachment record and its physical file
// @Tags File
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params query dto.FileDeleteRequest true "Delete Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.FileDTO} "Success"
// @Router /api/file [delete]
func (h *FileHandler) Delete(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.FileDeleteRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)

	if !valid {
		h.App.Logger().Error("FileHandler.Delete.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("FileHandler.Delete err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Calculate PathHash
	// 计算 PathHash
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	fileSvc := h.App.GetFileService(app.WebClientName, "")
	// Execute deletion
	// 执行删除
	file, err := fileSvc.Delete(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "FileHandler.Delete", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(file))

	h.WSS.BroadcastToUser(uid, code.Success.WithData(
		dto.FileSyncDeleteMessage{
			Path:     file.Path,
			PathHash: file.PathHash,
			Ctime:    file.Ctime,
			Mtime:    file.Mtime,
			Size:     file.Size,
		},
	).WithVault(params.Vault), "FileSyncDelete")
}

// Get retrieves file metadata
// @Summary Get attachment info
// @Description Get attachment metadata (FileDTO) by path
// @Tags File
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params query dto.FileGetRequest true "Get Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.FileDTO} "Success"
// @Router /api/file/info [get]
func (h *FileHandler) Get(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.FileGetRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("FileHandler.Get.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("FileHandler.Get err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Calculate PathHash
	// 计算 PathHash
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	fileSvc := h.App.GetFileService(app.WebClientName, "")
	file, err := fileSvc.Get(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "FileHandler.Get", err)
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}

	if file == nil {
		response.ToResponse(code.ErrorNoteNotFound)
		return
	}

	response.ToResponse(code.Success.WithData(file))
}

// Restore restores a file from trash
// @Summary Restore attachment
// @Description Restore deleted attachment from trash
// @Tags File
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params body dto.FileRestoreRequest true "Restore Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.FileDTO} "Success"
// @Router /api/file/restore [put]
func (h *FileHandler) Restore(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.FileRestoreRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("FileHandler.Restore.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("FileHandler.Restore err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Calculate PathHash
	// 计算 PathHash
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	fileSvc := h.App.GetFileService(app.WebClientName, "")

	// Execute restore
	// 执行恢复
	file, err := fileSvc.Restore(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "FileHandler.Restore.FileRestore", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(file))
	h.WSS.BroadcastToUser(uid, code.Success.WithData(file).WithVault(params.Vault), "FileSyncUpdate")
}

// logError records error log, including Trace ID
// logError 记录错误日志，包含 Trace ID
func (h *FileHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}

// RecycleClear clears files from recycle bin
// @Summary Clear recycle bin
// @Description Permanently clear selected files from recycle bin
// @Tags File
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.FileRecycleClearRequest true "Clear Parameters"
// @Success 200 {object} pkgapp.Res "Success"
// @Router /api/file/recycle-clear [delete]
func (h *FileHandler) RecycleClear(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.FileRecycleClearRequest{}

	if valid, errs := pkgapp.BindAndValid(c, params); !valid {
		h.App.Logger().Error("FileHandler.RecycleClear.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("FileHandler.RecycleClear err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	ctx := c.Request.Context()
	fileSvc := h.App.GetFileService(app.WebClientName, "")
	if err := fileSvc.RecycleClear(ctx, uid, params); err != nil {
		h.logError(ctx, "FileHandler.RecycleClear", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success)
}

// Rename renames a file
// @Summary Rename attachment
// @Description Rename an attachment to a new path
// @Tags File
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.FileRenameRequest true "Rename Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.FileDTO} "Success"
// @Router /api/file/rename [post]
func (h *FileHandler) Rename(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.FileRenameRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("FileHandler.Rename.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("FileHandler.Rename err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Validate paths
	if !util.ValidatePath(params.Path) || !util.ValidatePath(params.OldPath) {
		response.ToResponse(code.ErrorInvalidPath)
		return
	}

	// Calculate PathHash
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}
	if params.OldPathHash == "" {
		params.OldPathHash = util.EncodeHash32(params.OldPath)
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	fileSvc := h.App.GetFileService(app.WebClientName, "")

	oldFile, newFile, err := fileSvc.Rename(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "FileHandler.Rename", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(newFile))

	// Broadcast WebSocket event: FileSyncRename
	// 广播 WebSocket 事件: 文件同步重命名
	h.WSS.BroadcastToUser(uid, code.Success.WithData(dto.FileSyncRenameMessage{
		Path:             newFile.Path,
		PathHash:         newFile.PathHash,
		ContentHash:      newFile.ContentHash,
		Ctime:            newFile.Ctime,
		Mtime:            newFile.Mtime,
		Size:             newFile.Size,
		UpdatedTimestamp: newFile.UpdatedTimestamp,
		OldPath:          oldFile.Path,
		OldPathHash:      oldFile.PathHash,
	}).WithVault(params.Vault), "FileSyncRename")
}
