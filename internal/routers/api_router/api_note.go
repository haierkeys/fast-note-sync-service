package api_router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
)

// Note 笔记 API 路由处理器
// 结构体名: Note
// 说明: 处理笔记相关的 HTTP 请求，并持有 WebSocket 服务引用以进行实时广播。
type Note struct {
	wss *app.WS
}

// NewNote 创建 Note 路由处理器实例
// 函数名: NewNote
// 函数使用说明: 初始化并返回一个新的 Note 结构体实例。
// 参数说明:
//   - wss *app.WS: WebSocket 服务实例，用于消息广播
//
// 返回值说明:
//   - *Note: 初始化后的 Note 实例
func NewNote(wss *app.WS) *Note {
	return &Note{wss: wss}
}

// Get 获取单条笔记详情
// 函数名: Get
// 函数使用说明: 处理获取单条笔记的 HTTP 请求。验证参数和用户身份，调用 Service 层获取笔记内容。
// 参数说明:
//   - c *gin.Context: Gin 上下文，包含请求参数 (vault, path, pathHash 等)
//
// 返回值说明:
//   - JSON: 包含笔记详情的响应数据
func (n *Note) Get(c *gin.Context) {
	params := &service.NoteGetRequestParams{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, params)
	if !valid {
		global.Logger.Error("apiRouter.Note.Get.BindAndValid err: %v", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}
	uid := app.GetUID(c)
	if uid == 0 {
		global.Logger.Error("apiRouter.Note.Get err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	params.PathHash = util.EncodeHash32(params.Path)

	svc := service.New(c).WithClientName(global.WebClientName)
	note, err := svc.NoteGet(uid, params)
	if err != nil {
		global.Logger.Error("apiRouter.Note.Get svc NoteGet err: %v", zap.Error(err))
		response.ToResponse(code.ErrorNoteGetFailed.WithDetails(err.Error()))
		return
	}

	// 解析内容中的 ![[ ]] 标签
	fileLinks, err := svc.FileResolveEmbedLinks(uid, params.Vault, note.Content)
	if err != nil {
		global.Logger.Error("apiRouter.Note.Get svc FileResolveEmbedLinks err: %v", zap.Error(err))
	}

	noteWithLinks := &service.NoteWithFileLinks{
		FileLinks: fileLinks,
	}
	convert.StructAssign(note, noteWithLinks)

	response.ToResponse(code.Success.WithData(noteWithLinks))
}

// List 获取笔记列表
// 函数名: List
// 函数使用说明: 处理获取笔记列表的 HTTP 请求。支持分页查询，返回不包含内容的笔记摘要列表。
// 参数说明:
//   - c *gin.Context: Gin 上下文，包含分页参数 (page, pageSize, vault)
//
// 返回值说明:
//   - JSON: 包含笔记列表的响应数据
func (n *Note) List(c *gin.Context) {
	params := &service.NoteListRequestParams{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, params)
	if !valid {
		global.Logger.Error("apiRouter.Note.List.BindAndValid errs: %v", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}
	uid := app.GetUID(c)

	if uid == 0 {
		global.Logger.Error("apiRouter.Note.List err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	svc := service.New(c).WithClientName(global.WebClientName)

	pager := &app.Pager{Page: app.GetPage(c), PageSize: app.GetPageSize(c)}

	notes, count, err := svc.NoteList(uid, params, pager)
	if err != nil {
		global.Logger.Error("apiRouter.Note.List svc NoteList err: %v", zap.Error(err))
		response.ToResponse(code.ErrorNoteListFailed.WithDetails(err.Error()))
		return
	}
	response.ToResponseList(code.Success, notes, count)
}

// CreateOrUpdate 创建或更新笔记
// 函数名: CreateOrUpdate
// 函数使用说明: 处理创建或更新笔记的 HTTP 请求。
//   - 自动计算路径和内容哈希
//   - 处理文件重命名（SrcPath != Path）
//   - 检查冲突并更新数据库
//   - 通过 WebSocket 广播变更通知
//
// 参数说明:
//   - c *gin.Context: Gin 上下文，包含笔记数据 (vault, path, content, etc.)
//
// 返回值说明:
//   - JSON: 操作结果和更新后的笔记数据
func (n *Note) CreateOrUpdate(c *gin.Context) {
	params := &service.NoteModifyOrCreateRequestParams{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, params)
	if !valid {
		global.Logger.Error("apiRouter.Note.CreateOrUpdate.BindAndValid err: %v", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}
	uid := app.GetUID(c)
	if uid == 0 {
		global.Logger.Error("apiRouter.Note.CreateOrUpdate err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	params.SrcPathHash = util.EncodeHash32(params.SrcPath)
	params.PathHash = util.EncodeHash32(params.Path)
	params.ContentHash = util.EncodeHash32(params.Content)

	if params.Mtime == 0 {
		params.Mtime = time.Now().UnixMilli()
	}
	if params.Ctime == 0 {
		params.Ctime = params.Mtime
	}

	svc := service.New(c).WithClientName(global.WebClientName)

	if params.SrcPath != "" && params.SrcPath != params.Path {
		noteSrc, err := svc.NoteGet(uid, &service.NoteGetRequestParams{
			Vault:    params.Vault,
			Path:     params.SrcPath,
			PathHash: params.SrcPathHash,
		})
		if err != nil {
			global.Logger.Error("apiRouter.Note.CreateOrUpdate svc NoteGet err: %v", zap.Error(err))
			response.ToResponse(code.ErrorNoteGetFailed.WithDetails(err.Error()))
			return
		}
		// 如果源笔记不存在，或者源笔记是删除状态，直接返回
		if noteSrc == nil || noteSrc.Action == "delete" {
			response.ToResponse(code.ErrorNoteNotFound)
			return
		}
	}

	checkParams := convert.StructAssign(params, &service.NoteUpdateCheckRequestParams{}).(*service.NoteUpdateCheckRequestParams)
	_, noteSelect, err := svc.NoteUpdateCheck(uid, checkParams)

	if err != nil {
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}

	if noteSelect != nil {
		if noteSelect.Action == "delete" {
			response.ToResponse(code.ErrorNoteNotFound)
			return
		}

		if params.ContentHash != noteSelect.ContentHash {
			params.Mtime = time.Now().UnixMilli()
		}
	}

	var noteNew *service.Note
	var noteOld *service.Note
	// 如果路径发生变化，删除旧笔记
	if params.SrcPath != "" && params.SrcPath != params.Path {
		params := &service.NoteDeleteRequestParams{
			Vault:    params.Vault,
			Path:     params.SrcPath,
			PathHash: params.SrcPathHash,
		}
		noteOld, err = svc.NoteDelete(uid, params)
		if err != nil {
			global.Logger.Error("apiRouter.Note.CreateOrUpdate svc NoteDelete err: %v", zap.Error(err))
			response.ToResponse(code.ErrorNoteDeleteFailed.WithDetails(err.Error()))
			return
		}
		n.wss.BroadcastToUser(uid, code.Success.Reset().WithData(noteOld).WithVault(params.Vault), "NoteSyncDelete")
	}

	_, noteNew, err = svc.NoteModifyOrCreate(uid, params, false)
	if err != nil {
		global.Logger.Error("apiRouter.Note.CreateOrUpdate svc NoteModifyOrCreate err: %v", zap.Error(err))
		response.ToResponse(code.ErrorNoteModifyOrCreateFailed.WithDetails(err.Error()))
		return
	}

	response.ToResponse(code.Success.WithData(noteNew))
	n.wss.BroadcastToUser(uid, code.Success.Reset().WithData(noteNew).WithVault(params.Vault), "NoteSyncModify")

	if params.SrcPath != "" && params.SrcPath != params.Path {
		svc.NoteMigratePush(noteOld.ID, noteNew.ID, uid)
	}

}

// Delete 删除笔记
// 函数名: Delete
// 函数使用说明: 处理删除笔记的 HTTP 请求。标记笔记为删除状态，并通过 WebSocket 广播删除通知。
// 参数说明:
//   - c *gin.Context: Gin 上下文，包含待删除笔记的标识 (vault, path)
//
// 返回值说明:
//   - JSON: 操作结果
func (n *Note) Delete(c *gin.Context) {
	params := &service.NoteDeleteRequestParams{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, params)
	if !valid {
		global.Logger.Error("apiRouter.Note.Delete.BindAndValid err: %v", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}
	uid := app.GetUID(c)
	if uid == 0 {
		global.Logger.Error("apiRouter.Note.Delete err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}
	params.PathHash = util.EncodeHash32(params.Path)
	svc := service.New(c).WithClientName(global.WebClientName)

	noteSrc, err := svc.NoteGet(uid, &service.NoteGetRequestParams{
		Vault:    params.Vault,
		Path:     params.Path,
		PathHash: params.PathHash,
	})
	if err != nil {
		global.Logger.Error("apiRouter.Note.Delete svc NoteGet err: %v", zap.Error(err))
		response.ToResponse(code.ErrorNoteGetFailed.WithDetails(err.Error()))
		return
	}
	// 如果源笔记不存在，或者源笔记是删除状态，直接返回
	if noteSrc == nil || noteSrc.Action == "delete" {
		response.ToResponse(code.ErrorNoteNotFound)
		return
	}

	note, err := svc.NoteDelete(uid, params)
	if err != nil {
		global.Logger.Error("apiRouter.Note.Delete svc NoteDelete err: %v", zap.Error(err))
		response.ToResponse(code.ErrorNoteDeleteFailed.WithDetails(err.Error()))
		return
	}
	response.ToResponse(code.Success.WithData(note))
	n.wss.BroadcastToUser(uid, code.Success.Reset().WithData(note).WithVault(params.Vault), "NoteSyncDelete")
}

// GetFileContent 获取文件或笔记的原始内容
// 函数名: GetFileContent
// 函数使用说明: 处理通过仓库名和路径获取内容的 HTTP 请求。
// 参数说明:
//   - c *gin.Context: Gin 上下文,包含查询参数 vault 和 path
//
// 返回值说明:
//   - 二进制流: 文件的原始内容
func (n *Note) GetFileContent(c *gin.Context) {
	params := &service.NoteGetRequestParams{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, params)
	if !valid {
		global.Logger.Error("apiRouter.Note.Get.BindAndValid err: %v", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}
	uid := app.GetUID(c)
	if uid == 0 {
		global.Logger.Error("apiRouter.Note.Get err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	params.PathHash = util.EncodeHash32(params.Path)

	svc := service.New(c).WithClientName(global.WebClientName)
	content, contentType, err := svc.NoteGetFileContent(uid, params)
	if err != nil {
		global.Logger.Error("apiRouter.Note.GetFileContent err", zap.Error(err))
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}

	// 如果内容为 nil, 表示资源未找到或已删除, 返回 404
	if content == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// 返回内容，设置从 Service 层识别出的 Content-Type
	c.Data(http.StatusOK, contentType, content)
}
