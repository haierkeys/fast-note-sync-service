package api_router

import (
	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
)

type NoteHistory struct{}

func NewNoteHistory() *NoteHistory {
	return &NoteHistory{}
}

// NoteHistoryGetRequestParams 获取笔记历史详情请求参数
type NoteHistoryGetRequestParams struct {
	ID int64 `form:"id" binding:"required"`
}

// Get 获取单条笔记历史详情
// 函数名: Get
// 函数使用说明: 处理获取单条笔记历史的 HTTP 请求。
// 参数说明:
//   - c *gin.Context: Gin 上下文，包含请求参数 (id)
//
// 返回值说明:
//   - JSON: 包含笔记历史详情的响应数据
func (n *NoteHistory) Get(c *gin.Context) {
	params := &NoteHistoryGetRequestParams{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, params)
	if !valid {
		global.Logger.Error("apiRouter.NoteHistory.Get.BindAndValid err: %v", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}
	uid := app.GetUID(c)
	if uid == 0 {
		global.Logger.Error("apiRouter.NoteHistory.Get err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	svc := service.New(c).WithClientName(global.WebClientName)
	history, err := svc.NoteHistoryGet(uid, params.ID)
	if err != nil {
		global.Logger.Error("apiRouter.NoteHistory.Get svc NoteHistoryGet err: %v", zap.Error(err))
		response.ToResponse(code.ErrorNoteGetFailed.WithDetails(err.Error()))
		return
	}
	response.ToResponse(code.Success.WithData(history))
}

// List 获取笔记历史列表
// 函数名: List
// 函数使用说明: 处理获取笔记历史列表的 HTTP 请求。支持分页查询。
// 参数说明:
//   - c *gin.Context: Gin 上下文，包含分页参数 (page, pageSize, noteId)
//
// 返回值说明:
//   - JSON: 包含笔记历史列表的响应数据
func (n *NoteHistory) List(c *gin.Context) {
	params := &service.NoteHistoryListRequestParams{}
	response := app.NewResponse(c)
	valid, errs := app.BindAndValid(c, params)
	if !valid {
		global.Logger.Error("apiRouter.NoteHistory.List.BindAndValid errs: %v", zap.Error(errs))
		response.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}
	uid := app.GetUID(c)
	if uid == 0 {
		global.Logger.Error("apiRouter.NoteHistory.List err uid=0")
		response.ToResponse(code.ErrorInvalidUserAuthToken)
		return
	}

	svc := service.New(c).WithClientName(global.WebClientName)

	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	pager := &app.Pager{Page: app.GetPage(c), PageSize: app.GetPageSize(c)}

	list, count, err := svc.NoteHistoryList(uid, params, pager)
	if err != nil {
		global.Logger.Error("apiRouter.NoteHistory.List svc NoteHistoryList err: %v", zap.Error(err))
		response.ToResponse(code.ErrorNoteListFailed.WithDetails(err.Error()))
		return
	}
	response.ToResponseList(code.Success, list, int(count))
}
