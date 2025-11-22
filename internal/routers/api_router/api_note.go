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

type Note struct {
	wss *app.WebsocketServer
}

func NewNote(wss *app.WebsocketServer) *Note {
	return &Note{wss: wss}
}

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

	params.PathHash = util.EncodeHash32(params.Path)
	params.ContentHash = util.EncodeHash32(params.Content)

	svc := service.New(c)

	checkParams := convert.StructAssign(params, &service.NoteUpdateCheckRequestParams{}).(*service.NoteUpdateCheckRequestParams)
	isNew, _, _, noteSelect, err := svc.NoteUpdateCheck(uid, checkParams)

	if err != nil {
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}
	if noteSelect.Content != params.Content {
		params.Mtime = time.Now().UnixMilli()
	}

	if isNew {
		if params.Mtime == 0 {
			params.Mtime = time.Now().UnixMilli()
		}
		if params.Ctime == 0 {
			params.Ctime = params.Mtime
		}
	}

	_, note, err := svc.NoteModifyOrCreate(uid, params, false)
	if err != nil {
		global.Logger.Error("apiRouter.Note.CreateOrUpdate svc NoteModifyOrCreate err: %v", zap.Error(err))
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}

	response.ToResponse(code.Success.WithData(note))
	n.wss.BroadcastToUser(uid, code.Success.WithData(note), "NoteSyncModify")
}

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
		response.ToResponse(code.Failed.WithDetails(err.Error()))
		return
	}
	response.ToResponse(code.Success.WithData(note))

	n.wss.BroadcastToUser(uid, code.Success.WithData(note), "NoteSyncDelete")

}

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
