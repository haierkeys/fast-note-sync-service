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

// NoteHandler note API router handler
// NoteHandler 笔记 API 路由处理器
// Uses App Container to inject dependencies, supports unified error handling
// 使用 App Container 注入依赖，支持统一错误处理
type NoteHandler struct {
	*Handler
}

// NewNoteHandler creates NoteHandler instance
// NewNoteHandler 创建 NoteHandler 实例
func NewNoteHandler(a *app.App, wss *pkgapp.WebsocketServer) *NoteHandler {
	return &NoteHandler{
		Handler: NewHandlerWithWSS(a, wss),
	}
}

// Get retrieves note details
// @Summary Get note details
// @Description Get specific note content and metadata by path or path hash
// @Tags Note
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params query dto.NoteGetRequest true "Get Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.NoteWithFileLinksResponse} "Success"
// @Router /api/note [get]
func (h *NoteHandler) Get(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteGetRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.Get.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.Get err uid=0")
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

	noteSvc := h.App.GetNoteService(app.WebClientName, "")
	note, err := noteSvc.Get(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "NoteHandler.Get", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	// Parse ![[ ]] tags in content
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

// GetShared retrieves shared note details
// @Summary Get shared note details
// @Description Get specific note content (restricted read-only access) via share token
// @Tags Note
// @Security ShareAuthToken
// @Param Share-Token header string true "Auth Token"
// @Produce json
// @Success 200 {object} pkgapp.Res{data=dto.NoteDTO} "Success"
// @Router /api/share/note [get]

// List retrieves note list
// @Summary Get note list
// @Description Get note list for current user with pagination
// @Tags Note
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params query dto.NoteListRequest true "Query Parameters"
// @Success 200 {object} pkgapp.Res{data=[]dto.NoteDTO} "Success"
// @Router /api/notes [get]
func (h *NoteHandler) List(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteListRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.List.BindAndValid errs", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.List err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Get request context
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

// CreateOrUpdate creates or updates a note
// @Summary Create or update note
// @Description Handle note creation, modification, or renaming (identified by path change)
// @Tags Note
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.NoteModifyOrCreateRequest true "Note Content"
// @Success 200 {object} pkgapp.Res{data=dto.NoteDTO} "Success"
// @Router /api/note [post]
func (h *NoteHandler) CreateOrUpdate(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteModifyOrCreateRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.CreateOrUpdate.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.CreateOrUpdate err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Validate paths for directory traversal attacks
	if !util.ValidatePath(params.Path) {
		response.ToResponse(code.ErrorInvalidPath)
		return
	}
	if params.SrcPath != "" && !util.ValidatePath(params.SrcPath) {
		response.ToResponse(code.ErrorInvalidPath)
		return
	}

	// Apply default folder if configured
	// if defaultFolder := h.App.Config().App.DefaultAPIFolder; defaultFolder != "" {
	// 	params.Path = util.ApplyDefaultFolder(params.Path, defaultFolder)
	// 	if params.SrcPath != "" {
	// 		params.SrcPath = util.ApplyDefaultFolder(params.SrcPath, defaultFolder)
	// 	}
	// }

	// Calculate hash values
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

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	noteSvc := h.App.GetNoteService(app.WebClientName, "")

	// Handle rename scenarios
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

	// Check update
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

	// If path changed, delete old note
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

// Delete deletes a note
// @Summary Delete note
// @Description Move note to trash
// @Tags Note
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params query dto.NoteDeleteRequest true "Delete Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.NoteDTO} "Success"
// @Router /api/note [delete]
func (h *NoteHandler) Delete(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteDeleteRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.Delete.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.Delete err uid=0")
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

	noteSvc := h.App.GetNoteService(app.WebClientName, "")

	// Check if note exists
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

	// Execute deletion
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

// Restore restores a note from trash
// @Summary Restore note
// @Description Restore deleted note from trash
// @Tags Note
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params body dto.NoteRestoreRequest true "Restore Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.NoteDTO} "Success"
// @Router /api/note/restore [put]
func (h *NoteHandler) Restore(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteRestoreRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.Restore.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.Restore err uid=0")
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

	noteSvc := h.App.GetNoteService(app.WebClientName, "")

	// Check if note exists in trash
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

	// Execute restore
	// 执行恢复
	note, err := noteSvc.Restore(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "NoteHandler.Restore.NoteRestore", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(note))
	h.WSS.BroadcastToUser(uid, code.Success.WithData(note).WithVault(params.Vault), "NoteSyncModify")
}

// PatchFrontmatter modifies note frontmatter
// @Summary Modify note frontmatter
// @Description Update or delete note frontmatter fields
// @Tags Note
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.NotePatchFrontmatterRequest true "Frontmatter Modification Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.NoteDTO} "Success"
// @Router /api/note/frontmatter [patch]
func (h *NoteHandler) PatchFrontmatter(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NotePatchFrontmatterRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.PatchFrontmatter.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.PatchFrontmatter err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Apply default folder if configured
	// if defaultFolder := h.App.Config().App.DefaultAPIFolder; defaultFolder != "" {
	// 	params.Path = util.ApplyDefaultFolder(params.Path, defaultFolder)
	// }

	// Calculate PathHash
	// 计算 PathHash
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	noteSvc := h.App.GetNoteService(app.WebClientName, "")
	note, err := noteSvc.PatchFrontmatter(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "NoteHandler.PatchFrontmatter", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(note))
	h.WSS.BroadcastToUser(uid, code.Success.WithData(note).WithVault(params.Vault), "NoteSyncModify")
}

// Append appends content to a note
// @Summary Append content to note
// @Description Append content to the end of a note
// @Tags Note
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.NoteAppendRequest true "Append Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.NoteDTO} "Success"
// @Router /api/note/append [post]
func (h *NoteHandler) Append(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteAppendRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.Append.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.Append err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Apply default folder if configured
	// if defaultFolder := h.App.Config().App.DefaultAPIFolder; defaultFolder != "" {
	// 	params.Path = util.ApplyDefaultFolder(params.Path, defaultFolder)
	// }

	// Calculate PathHash
	// 计算 PathHash
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	noteSvc := h.App.GetNoteService(app.WebClientName, "")
	note, err := noteSvc.AppendContent(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "NoteHandler.Append", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(note))
	h.WSS.BroadcastToUser(uid, code.Success.WithData(note).WithVault(params.Vault), "NoteSyncModify")
}

// Prepend inserts content at the beginning of a note
// @Summary Prepend content to note
// @Description Insert content at the beginning of a note (after frontmatter)
// @Tags Note
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.NotePrependRequest true "Prepend Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.NoteDTO} "Success"
// @Router /api/note/prepend [post]
func (h *NoteHandler) Prepend(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NotePrependRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.Prepend.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.Prepend err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Apply default folder if configured
	// if defaultFolder := h.App.Config().App.DefaultAPIFolder; defaultFolder != "" {
	// 	params.Path = util.ApplyDefaultFolder(params.Path, defaultFolder)
	// }

	// Calculate PathHash
	// 计算 PathHash
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	noteSvc := h.App.GetNoteService(app.WebClientName, "")
	note, err := noteSvc.PrependContent(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "NoteHandler.Prepend", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(note))
	h.WSS.BroadcastToUser(uid, code.Success.WithData(note).WithVault(params.Vault), "NoteSyncModify")
}

// Replace performs find and replace in a note
// @Summary Find and replace in note
// @Description Perform find and replace operation in a note, supporting regular expressions
// @Tags Note
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.NoteReplaceRequest true "Find and Replace Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.NoteReplaceResponse} "Success"
// @Router /api/note/replace [post]
func (h *NoteHandler) Replace(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteReplaceRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.Replace.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.Replace err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Apply default folder if configured
	// if defaultFolder := h.App.Config().App.DefaultAPIFolder; defaultFolder != "" {
	// 	params.Path = util.ApplyDefaultFolder(params.Path, defaultFolder)
	// }

	// Calculate PathHash
	// 计算 PathHash
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	noteSvc := h.App.GetNoteService(app.WebClientName, "")
	result, err := noteSvc.ReplaceContent(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "NoteHandler.Replace", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(result))
	if result.Note != nil {
		h.WSS.BroadcastToUser(uid, code.Success.WithData(result.Note).WithVault(params.Vault), "NoteSyncModify")
	}
}

// Move moves a note to a new path
// @Summary Move note
// @Description Move a note to a new path
// @Tags Note
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Accept json
// @Produce json
// @Param params body dto.NoteMoveRequest true "Move Parameters"
// @Success 200 {object} pkgapp.Res{data=dto.NoteDTO} "Success"
// @Router /api/note/move [post]
func (h *NoteHandler) Move(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteMoveRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.Move.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.Move err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Validate paths for directory traversal attacks
	if !util.ValidatePath(params.Path) || !util.ValidatePath(params.Destination) {
		response.ToResponse(code.ErrorInvalidPath)
		return
	}

	// Apply default folder if configured
	// if defaultFolder := h.App.Config().App.DefaultAPIFolder; defaultFolder != "" {
	// 	params.Path = util.ApplyDefaultFolder(params.Path, defaultFolder)
	// 	params.Destination = util.ApplyDefaultFolder(params.Destination, defaultFolder)
	// }

	// Calculate PathHash
	// 计算 PathHash
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	noteSvc := h.App.GetNoteService(app.WebClientName, "")

	// Get old note for broadcasting deletion event
	// 获取旧笔记用于广播删除事件
	oldNote, _ := noteSvc.Get(ctx, uid, &dto.NoteGetRequest{
		Vault:    params.Vault,
		Path:     params.Path,
		PathHash: params.PathHash,
	})

	note, err := noteSvc.Move(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "NoteHandler.Move", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(note))

	// Broadcast WebSocket events: delete old path + create new path
	// 广播 WebSocket 事件: 删除旧路径 + 新建新路径
	if oldNote != nil {
		h.WSS.BroadcastToUser(uid, code.Success.WithData(oldNote).WithVault(params.Vault), "NoteSyncDelete")
	}
	h.WSS.BroadcastToUser(uid, code.Success.WithData(note).WithVault(params.Vault), "NoteSyncModify")
}

// GetBacklinks retrieves backlinks to a specific note
// @Summary Get backlinks
// @Description Get all other notes that link to the specified note
// @Tags Note
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params query dto.NoteLinkQueryRequest true "Query Parameters"
// @Success 200 {object} pkgapp.Res{data=[]dto.NoteLinkItem} "Success"
// @Router /api/note/backlinks [get]
func (h *NoteHandler) GetBacklinks(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteLinkQueryRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.GetBacklinks.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.GetBacklinks err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Apply default folder if configured
	// if defaultFolder := h.App.Config().App.DefaultAPIFolder; defaultFolder != "" {
	// 	params.Path = util.ApplyDefaultFolder(params.Path, defaultFolder)
	// }

	// Calculate PathHash
	// 计算 PathHash
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	links, err := h.App.NoteLinkService.GetBacklinks(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "NoteHandler.GetBacklinks", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(links))
}

// GetOutlinks retrieves outgoing links from a specific note
// @Summary Get outgoing links
// @Description Get other notes that the specified note links to
// @Tags Note
// @Security UserAuthToken
// @Param token header string true "Auth Token"
// @Produce json
// @Param params query dto.NoteLinkQueryRequest true "Query Parameters"
// @Success 200 {object} pkgapp.Res{data=[]dto.NoteLinkItem} "Success"
// @Router /api/note/outlinks [get]
func (h *NoteHandler) GetOutlinks(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	params := &dto.NoteLinkQueryRequest{}

	// Parameter binding and validation
	// 参数绑定和验证
	valid, errs := pkgapp.BindAndValid(c, params)
	if !valid {
		h.App.Logger().Error("NoteHandler.GetOutlinks.BindAndValid err", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	// Get UID
	// 获取用户 ID
	uid := pkgapp.GetUID(c)
	if uid == 0 {
		h.App.Logger().Error("NoteHandler.GetOutlinks err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	// Apply default folder if configured
	// if defaultFolder := h.App.Config().App.DefaultAPIFolder; defaultFolder != "" {
	// 	params.Path = util.ApplyDefaultFolder(params.Path, defaultFolder)
	// }

	// Calculate PathHash
	// 计算 PathHash
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// Get request context
	// 获取请求上下文
	ctx := c.Request.Context()

	links, err := h.App.NoteLinkService.GetOutlinks(ctx, uid, params)
	if err != nil {
		h.logError(ctx, "NoteHandler.GetOutlinks", err)
		apperrors.ErrorResponse(c, err)
		return
	}

	response.ToResponse(code.Success.WithData(links))
}

// logError records error log, including Trace ID
// logError 记录错误日志，包含 Trace ID
func (h *NoteHandler) logError(ctx context.Context, method string, err error) {
	traceID := middleware.GetTraceID(ctx)
	h.App.Logger().Error(method,
		zap.Error(err),
		zap.String("traceId", traceID),
	)
}
