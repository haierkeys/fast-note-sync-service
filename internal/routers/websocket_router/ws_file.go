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

// NoteModify 处理文件修改的 WebSocket 消息
// 函数名: NoteModify
// 函数使用说明: 处理客户端发送的笔记修改或创建消息，进行参数校验、更新检查并在需要时写回数据库或通知其他客户端。
// 参数说明:
//   - c *app.WebsocketClient: 当前 WebSocket 客户端连接，包含上下文、用户信息、发送响应等能力。
//   - msg *app.WebSocketMessage: 接收到的 WebSocket 消息，包含消息数据和类型。
//
// 返回值说明:
//   - 无
func FileModify(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.NoteModifyOrCreateRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.note.NoteModify.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}
	if params.PathHash == "" {
		c.ToResponse(code.ErrorInvalidParams.WithDetails("pathHash is required"))
		return
	}
	if params.ContentHash == "" {
		c.ToResponse(code.ErrorInvalidParams.WithDetails("contentHash is required"))
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

	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	checkParams := convert.StructAssign(params, &service.NoteUpdateCheckRequestParams{}).(*service.NoteUpdateCheckRequestParams)
	isNew, isNeedUpdate, isNeedSyncMtime, _, err := svc.NoteUpdateCheck(c.User.UID, checkParams)

	if err != nil {
		c.ToResponse(code.ErrorNoteModifyFailed.WithDetails(err.Error()))
		return
	}

	var note *service.Note

	if isNew || isNeedSyncMtime || isNeedUpdate {
		_, note, err = svc.NoteModifyOrCreate(c.User.UID, params, true)
		if err != nil {
			c.ToResponse(code.ErrorNoteModifyFailed.WithDetails(err.Error()))
			return
		}
		// 通知所有客户端更新mtime
		if isNeedSyncMtime {
			c.ToResponse(code.Success.Reset())
			c.BroadcastResponse(code.Success.Reset().WithData(note), false, "NoteSyncModify")
			return
		} else if isNeedUpdate || isNew {

			c.ToResponse(code.Success.Reset())
			c.BroadcastResponse(code.Success.Reset().WithData(note), true, "NoteSyncModify")
			return
		}
	} else {
		c.ToResponse(code.SuccessNoUpdate.Reset())
		return
	}

}

// NoteModifyCheck 检查文件修改必要性
// 函数名: NoteModifyCheck
// 函数使用说明: 仅用于检查客户端提供的笔记状态与服务器状态的差异，决定客户端是否需要上传笔记或只需同步 mtime。
// 参数说明:
//   - c *app.WebsocketClient: 当前 WebSocket 客户端连接，包含上下文和用户信息。
//   - msg *app.WebSocketMessage: 接收到的消息，包含需要检查的笔记信息。
//
// 返回值说明:
//   - 无
func FileModifyCheck(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.NoteUpdateCheckRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.note.NoteModify.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)
	_, isNeedUpdate, isNeedSyncMtime, note, err := svc.NoteUpdateCheck(c.User.UID, params)

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

// NoteDelete 处理文件删除的 WebSocket 消息
// 函数名: NoteDelete
// 函数使用说明: 接收客户端的笔记删除请求，执行删除操作并通知其他客户端同步删除事件。
// 参数说明:
//   - c *app.WebsocketClient: 当前 WebSocket 客户端连接，包含发送响应与广播能力。
//   - msg *app.WebSocketMessage: 接收到的删除请求消息，包含要删除的笔记标识等参数。
//
// 返回值说明:
//   - 无
func FileDelete(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.NoteDeleteRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.note.NoteDelete.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)

	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	note, err := svc.NoteDelete(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorNoteDeleteFailed.WithDetails(err.Error()))
		return
	}

	c.ToResponse(code.Success)
	c.BroadcastResponse(code.Success.WithData(note), true, "NoteSyncDelete")
}

// NoteSync 处理全量或增量笔记同步
// 函数名: NoteSync
// 函数使用说明: 根据客户端提供的本地笔记列表与服务器端最近更新列表比较，决定返回哪些笔记需要上传、需要同步 mtime、需要删除或需要更新；最后返回同步结束消息。
// 参数说明:
//   - c *app.WebsocketClient: 当前 WebSocket 客户端连接，包含上下文与响应发送能力。
//   - msg *app.WebSocketMessage: 接收到的同步请求，包含客户端的笔记摘要和同步起始时间等信息。
//
// 返回值说明:
//   - 无
func FileSync(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.NoteSyncRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.note.NoteModify.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)

	svc.VaultGetOrCreate(params.Vault, c.User.UID)

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
