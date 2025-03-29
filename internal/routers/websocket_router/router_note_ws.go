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
 * FileModify
 * @Description        处理文件修改的WebSocket消息
 * @Create             HaierKeys 2025-03-01 17:30
 * @Param              c  *app.WebsocketClient  WebSocket客户端连接
 * @Param              msg  *app.WebSocketMessage  接收到的WebSocket消息
 * @Return             无
 */
func FileModifyByMtime(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.FileModifyOrCreateRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.note.FileModify.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx)
	note, err := svc.FileModifyOrCreate(c.User.UID, params, true)

	if err != nil {
		c.ToResponse(code.ErrorFileModifyFailed.WithDetails(err.Error()))
		return
	}
	if note == nil {
		c.ToResponse(code.SuccessNoUpdate.Reset())
	} else {
		c.ToResponse(code.Success)
	}

	if len(*c.UserClients) > 1 && note != nil {
		c.BroadcastResponse(code.Success.Reset().WithData(note), true, "SyncFileModify")
	}
}

/**
 * FileModifyOverride
 * @Description        处理文件修改的WebSocket消息
 * @Create             HaierKeys 2025-03-01 17:30
 * @Param              c  *app.WebsocketClient  WebSocket客户端连接
 * @Param              msg  *app.WebSocketMessage  接收到的WebSocket消息
 * @Return             无
 */
func FileModifyOverride(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.FileModifyOrCreateRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.note.FileModify.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx)
	note, err := svc.FileModifyOrCreate(c.User.UID, params, false)

	if err != nil {
		c.ToResponse(code.ErrorFileModifyFailed.WithDetails(err.Error()))
		return
	}
	c.ToResponse(code.Success)

	if len(*c.UserClients) > 1 && note != nil {
		c.BroadcastResponse(code.Success.WithData(note), true, "SyncFileModify")
	}
}

/**
 * FileDelete
 * @Description        处理文件删除的WebSocket消息
 * @Create             HaierKeys 2025-03-01 17:30
 * @Param              c  *app.WebsocketClient  WebSocket客户端连接
 * @Param              msg  *app.WebSocketMessage  接收到的WebSocket消息
 * @Return             无
 */
func FileDelete(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.FileDeleteRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.note.FileDelete.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx)
	note, err := svc.FileDelete(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorFileDeleteFailed.WithDetails(err.Error()))
		return
	}

	c.ToResponse(code.Success)
	if len(*c.UserClients) > 0 {
		c.BroadcastResponse(code.Success.WithData(note), true, "SyncFileDelete")
	}
}

func SyncFiles(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.SyncFilesRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.note.FileModify.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx)
	list, err := svc.SyncFiles(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorFileModifyFailed.WithDetails(err.Error()))
		return
	}

	var lastUpdateTime timex.Time

	for _, note := range list {
		if note.UpdatedAt.After(lastUpdateTime) {
			lastUpdateTime = note.UpdatedAt
		}
		if note.Action == "delete" {
			c.ToResponse(code.Success.WithData(note), "SyncFileDelete")
		} else {

			c.ToResponse(code.Success.WithData(note), "SyncFileModify")
		}
	}
	if list == nil {
		lastUpdateTime = timex.Now()
	}

	message := &service.SyncFilesEndMessage{
		Vault:        params.Vault,
		LastUpdateAt: lastUpdateTime,
	}

	c.ToResponse(code.Success.WithData(message), "SyncFilesEnd")

}

/**
 * ContentModify
 * @Description        处理文件内容修改的WebSocket消息
 * @Create             HaierKeys 2025-03-01 17:30
 * @Param              c  *app.WebsocketClient  WebSocket客户端连接
 * @Param              msg  *app.WebSocketMessage  接收到的WebSocket消息
 * @Return             无
 */
func ContentModify(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.ContentModifyRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.note.ContentModify.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx)
	err := svc.ContentModify(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorFileContentModifyFailed.WithDetails(err.Error()))
		return
	}
	c.ToResponse(code.Success)
}
