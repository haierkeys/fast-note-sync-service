package websocket_router

import (
	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/haierkeys/obsidian-better-sync-service/internal/service"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/app"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/code"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/timex"

	"go.uber.org/zap"
)

/**
 * NoteModify
 * @Description        处理文件修改的WebSocket消息
 * @Create             HaierKeys 2025-03-01 17:30
 * @Param              c  *app.WebsocketClient  WebSocket客户端连接
 * @Param              msg  *app.WebSocketMessage  接收到的WebSocket消息
 * @Return             无
 */
func NoteModifyByMtime(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.NoteModifyOrCreateRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.note.NoteModify.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx)
	note, err := svc.NoteModifyOrCreate(c.User.UID, params, true)

	if err != nil {
		c.ToResponse(code.ErrorNoteModifyFailed.WithDetails(err.Error()))
		return
	}
	if note == nil {
		c.ToResponse(code.SuccessNoUpdate.Reset())
	} else {
		c.ToResponse(code.Success.Reset())
	}

	if len(*c.UserClients) > 1 && note != nil {
		c.BroadcastResponse(code.Success.Reset().WithData(note), true, "NoteSyncModify")
	}
}

/**
 * NoteModifyOverride
 * @Description        处理文件修改的WebSocket消息
 * @Create             HaierKeys 2025-03-01 17:30
 * @Param              c  *app.WebsocketClient  WebSocket客户端连接
 * @Param              msg  *app.WebSocketMessage  接收到的WebSocket消息
 * @Return             无
 */
func NoteModifyOverride(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.NoteModifyOrCreateRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.note.NoteModify.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx)
	note, err := svc.NoteModifyOrCreate(c.User.UID, params, false)

	if err != nil {
		c.ToResponse(code.ErrorNoteModifyFailed.WithDetails(err.Error()))
		return
	}
	c.ToResponse(code.Success)

	if len(*c.UserClients) > 1 && note != nil {
		c.BroadcastResponse(code.Success.WithData(note), true, "NoteSyncModify")
	}
}

/**
 * NoteDelete
 * @Description        处理文件删除的WebSocket消息
 * @Create             HaierKeys 2025-03-01 17:30
 * @Param              c  *app.WebsocketClient  WebSocket客户端连接
 * @Param              msg  *app.WebSocketMessage  接收到的WebSocket消息
 * @Return             无
 */
func NoteDelete(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.NoteDeleteRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.note.NoteDelete.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx)
	note, err := svc.NoteDelete(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorNoteDeleteFailed.WithDetails(err.Error()))
		return
	}

	c.ToResponse(code.Success)
	if len(*c.UserClients) > 0 {
		c.BroadcastResponse(code.Success.WithData(note), true, "NoteSyncDelete")
	}
}

func NoteSync(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.NoteSyncRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.note.NoteModify.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx)
	list, err := svc.NoteListByLastTime(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorNoteModifyFailed.WithDetails(err.Error()))
		return
	}

	var lastTime int64

	for _, note := range list {
		if note.UpdatedTimestamp >= lastTime {
			lastTime = note.UpdatedTimestamp
		}
		if note.Action == "delete" {
			c.ToResponse(code.Success.WithData(note), "NoteSyncDelete")
		} else {
			c.ToResponse(code.Success.WithData(note), "NoteSyncModify")
		}
	}
	if list == nil {
		lastTime = timex.Now().UnixMilli()
	}

	message := &service.NoteSyncEndMessage{
		Vault:    params.Vault,
		LastTime: lastTime,
	}

	c.ToResponse(code.Success.WithData(message), "NoteSyncEnd")

}
