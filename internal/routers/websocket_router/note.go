package websocket_router

import (
	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/haierkeys/obsidian-better-sync-service/internal/service"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/app"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/code"

	"go.uber.org/zap"
)

/**
 * FileCreate
 * @Description        处理文件创建的WebSocket消息
 * @Create             HaierKeys 2025-03-01 17:30
 * @Param              c  *app.WebsocketClient  WebSocket客户端连接
 * @Param              msg  *app.WebSocketMessage  接收到的WebSocket消息
 * @Return             无
 */
func FileCreate(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.FileCreateRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.user.Login.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx)
	svcData, err := svc.FileCreate(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorUserLoginFailed.WithDetails(err.Error()))
		return
	}
	c.ToResponse(code.Success.WithData(svcData))
}

/**
 * FileModify
 * @Description        处理文件修改的WebSocket消息
 * @Create             HaierKeys 2025-03-01 17:30
 * @Param              c  *app.WebsocketClient  WebSocket客户端连接
 * @Param              msg  *app.WebSocketMessage  接收到的WebSocket消息
 * @Return             无
 */
func FileModify(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.FileModifyRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("api_router.user.Login.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx)
	err := svc.FileModify(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorUserLoginFailed.WithDetails(err.Error()))
		return
	}
	c.ToResponse(code.Success)
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
		global.Logger.Error("api_router.user.Login.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx)
	err := svc.ContentModify(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorUserLoginFailed.WithDetails(err.Error()))
		return
	}
	c.ToResponse(code.Success)
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
		global.Logger.Error("api_router.user.Login.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx)
	err := svc.FileDelete(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorUserLoginFailed.WithDetails(err.Error()))
		return
	}
	c.ToResponse(code.Success)
}
