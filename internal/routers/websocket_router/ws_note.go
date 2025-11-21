package websocket_router

import (
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"

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
	if params.Mtime == 0 {
		c.ToResponse(code.ErrorInvalidParams.WithDetails("mtime is required"))
		return
	}
	if params.Ctime == 0 {
		c.ToResponse(code.ErrorInvalidParams.WithDetails("ctime is required"))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)

	checkParams := convert.StructAssign(params, &service.NoteUpdateCheckRequestParams{}).(*service.NoteUpdateCheckRequestParams)
	isNeedUpdate, isNeedSyncMtime, _, err := svc.NoteUpdateCheck(c.User.UID, checkParams)

	if err != nil {
		c.ToResponse(code.ErrorNoteModifyFailed.WithDetails(err.Error()))
		return
	}

	var note *service.Note
	if isNeedUpdate || isNeedSyncMtime {
		note, err = svc.NoteModifyOrCreate(c.User.UID, params, true)
		if err != nil {
			c.ToResponse(code.ErrorNoteModifyFailed.WithDetails(err.Error()))
			return
		}
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
 * NoteModify
 * @Description        处理文件修改的WebSocket消息
 * @Create             HaierKeys 2025-03-01 17:30
 * @Param              c  *app.WebsocketClient  WebSocket客户端连接
 * @Param              msg  *app.WebSocketMessage  接收到的WebSocket消息
 * @Return             无
 */
func NoteModifyCheck(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.NoteUpdateCheckRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.note.NoteModify.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)
	isNeedUpdate, isNeedSyncMtime, note, err := svc.NoteUpdateCheck(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorNoteModifyFailed.WithDetails(err.Error()))
		return
	}

	// 通知客户端上传笔记
	if isNeedUpdate {
		NoteCheck := convert.StructAssign(note, &service.NoteSyncNeedPushMessage{}).(*service.NoteSyncNeedPushMessage)
		c.ToResponse(code.Success.WithData(NoteCheck), "NoteSyncNeedPush")
		return
	}
	// 强制客户端更新mtime 不传输笔记内容
	if isNeedSyncMtime {
		NoteCheck := convert.StructAssign(note, &service.NoteSyncMtimeMessage{}).(*service.NoteSyncMtimeMessage)
		c.ToResponse(code.Success.WithData(NoteCheck), "NoteSyncMtime")
		return
	}
	c.ToResponse(code.SuccessNoUpdate.Reset())
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

	svc := service.New(c.Ctx).WithSF(c.SF)

	err := svc.NoteDelete(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorNoteDeleteFailed.WithDetails(err.Error()))
		return
	}

	c.ToResponse(code.Success)

	if len(*c.UserClients) > 0 {
		c.BroadcastResponse(code.Success, true, "NoteSyncDelete")
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

	svc := service.New(c.Ctx).WithSF(c.SF)

	list, err := svc.NoteListByLastTime(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorNoteModifyFailed.WithDetails(err.Error()))
		return
	}

	var cNotes map[string]service.NoteSyncCheckRequestParams = make(map[string]service.NoteSyncCheckRequestParams, 0)
	var cNotesKeys map[string]struct{} = make(map[string]struct{}, 0)

	if len(params.Notes) > 0 {
		for _, note := range params.Notes {
			cNotes[note.PathHash] = note
			cNotesKeys[note.PathHash] = struct{}{}
		}
	}

	var lastTime int64

	for _, note := range list {
		if note.UpdatedTimestamp >= lastTime {
			lastTime = note.UpdatedTimestamp
		}

		if note.Action == "delete" {
			c.ToResponse(code.Success.WithData(note), "NoteSyncDelete")
		} else {
			if cNote, ok := cNotes[note.PathHash]; ok {

				delete(cNotesKeys, note.PathHash)

				if note.ContentHash == cNote.ContentHash && note.Mtime == cNote.Mtime {
					continue
				} else if note.ContentHash != cNote.ContentHash {
					NoteCheck := convert.StructAssign(note, &service.NoteSyncNeedPushMessage{}).(*service.NoteSyncNeedPushMessage)
					c.ToResponse(code.Success.WithData(NoteCheck), "NoteSyncNeedPush")
				} else {
					NoteCheck := convert.StructAssign(note, &service.NoteSyncMtimeMessage{}).(*service.NoteSyncMtimeMessage)
					c.ToResponse(code.Success.WithData(NoteCheck), "NoteSyncMtime")
				}
			} else {
				c.ToResponse(code.Success.WithData(note), "NoteSyncModify")
			}
		}
	}

	if list == nil {
		lastTime = timex.Now().UnixMilli()
	}
	if len(cNotesKeys) > 0 {
		for pathHash := range cNotesKeys {
			note := cNotes[pathHash]
			NoteCheck := convert.StructAssign(&note, &service.NoteSyncNeedPushMessage{}).(*service.NoteSyncNeedPushMessage)
			c.ToResponse(code.Success.WithData(NoteCheck), "NoteSyncNeedPush")
		}
	}

	message := &service.NoteSyncEndMessage{
		Vault:    params.Vault,
		LastTime: lastTime,
	}
	c.ToResponse(code.Success.WithData(message), "NoteSyncEnd")
}

// 用户ws 服务器用户有效性验证
func UserInfo(c *app.WebsocketClient, uid int64) (*app.UserSelectEntity, error) {

	svc := service.New(c.Ctx).WithSF(c.SF)
	user, err := svc.UserInfo(uid)

	var userEntity *app.UserSelectEntity
	if user != nil {
		userEntity = convert.StructAssign(user, &app.UserSelectEntity{}).(*app.UserSelectEntity)
	}

	return userEntity, err

}
