package api_router

import (
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

// Note 表示笔记相关的 API 路由处理器，包含 WebSocket 服务用于实时同步。
type Note struct {
	wss *app.WebsocketServer
}

// NewNote 创建一个新的 Note 实例。
func NewNote(wss *app.WebsocketServer) *Note {
	return &Note{wss: wss}
}

// Get 获取指定路径的笔记内容。
// 接收路径或路径哈希作为参数，若未提供路径哈希则自动计算。
// 验证用户身份后调用服务层获取笔记数据，并返回结果。
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
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}

	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	svc := service.New(c)
	note, err := svc.NoteGet(uid, params)
	if err != nil {
		global.Logger.Error("apiRouter.Note.Get svc NoteGet err: %v", zap.Error(err))
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}
	response.ToResponse(code.Success.WithData(note))
}

// List 列出当前用户符合条件的笔记列表。
// 支持分页、过滤等参数，调用服务层获取笔记列表并返回。
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
	svc := service.New(c)
	notes, err := svc.NoteList(uid, params)
	if err != nil {
		global.Logger.Error("apiRouter.Note.List svc NoteList err: %v", zap.Error(err))
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}
	response.ToResponse(code.Success.WithData(notes))
}

// CreateOrUpdate 创建新笔记或更新已有笔记。
// 若提供了 SrcPath 且与 Path 不同，则先检查源笔记是否存在且未被删除，然后执行重命名（删除旧路径笔记 + 创建新路径笔记）。
// 自动填充缺失的时间戳和哈希值，冲突检测后广播同步事件。
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
		response.ToResponse(code.ErrorNotUserAuthToken)
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

	svc := service.New(c)

	if params.SrcPath != "" && params.SrcPath != params.Path {
		noteSrc, err := svc.NoteGet(uid, &service.NoteGetRequestParams{
			Vault:    params.Vault,
			Path:     params.SrcPath,
			PathHash: params.SrcPathHash,
		})
		if err != nil {
			global.Logger.Error("apiRouter.Note.CreateOrUpdate svc NoteGet err: %v", zap.Error(err))
			response.ToResponse(code.Failed.WithDetails(err.Error()))
			return
		}
		// 如果源笔记不存在，或者源笔记是删除状态，直接返回
		if noteSrc == nil || noteSrc.Action == "delete" {
			response.ToResponse(code.ErrorNoteNotFound)
			return
		}
	}

	checkParams := convert.StructAssign(params, &service.NoteUpdateCheckRequestParams{}).(*service.NoteUpdateCheckRequestParams)
	_, _, _, noteSelect, err := svc.NoteUpdateCheck(uid, checkParams)

	if err != nil {
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}

	if noteSelect != nil {
		if noteSelect.Action == "delete" {
			response.ToResponse(code.ErrorNoteDelete)
			return
		}

		if params.ContentHash != noteSelect.ContentHash {
			params.Mtime = time.Now().UnixMilli()
		}
	}

	// 如果路径发生变化，删除旧笔记
	if params.SrcPath != "" && params.SrcPath != params.Path {
		params := &service.NoteDeleteRequestParams{
			Vault:    params.Vault,
			Path:     params.SrcPath,
			PathHash: params.SrcPathHash,
		}
		note, err := svc.NoteDelete(uid, params)
		if err != nil {
			global.Logger.Error("apiRouter.Note.CreateOrUpdate svc NoteDelete err: %v", zap.Error(err))
			response.ToResponse(code.ErrorNoteDelete.WithDetails(err.Error()))
			return
		}
		n.wss.BroadcastToUser(uid, code.Success.WithData(note), "NoteSyncDelete")
	}

	_, note, err := svc.NoteModifyOrCreate(uid, params, false)
	if err != nil {
		global.Logger.Error("apiRouter.Note.CreateOrUpdate svc NoteModifyOrCreate err: %v", zap.Error(err))
		response.ToResponse(code.ErrorNoteModifyOrCreate.WithDetails(err.Error()))
		return
	}

	response.ToResponse(code.Success.WithData(note))
	n.wss.BroadcastToUser(uid, code.Success.WithData(note), "NoteSyncModify")
}

// Delete 删除指定路径的笔记。
// 验证用户身份后调用服务层执行删除操作，并通过 WebSocket 广播删除事件。
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
		response.ToResponse(code.ErrorNotUserAuthToken)
		return
	}
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}
	svc := service.New(c)
	note, err := svc.NoteDelete(uid, params)
	if err != nil {
		global.Logger.Error("apiRouter.Note.Delete svc NoteDelete err: %v", zap.Error(err))
		response.ToResponse(code.ErrorNoteDelete.WithDetails(err.Error()))
		return
	}
	response.ToResponse(code.Success.WithData(note))
	n.wss.BroadcastToUser(uid, code.Success.WithData(note), "NoteSyncDelete")
}
