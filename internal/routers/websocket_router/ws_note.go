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

type NoteMessage struct {
	Path             string `json:"path" form:"path"`                 // 路径信息（文件路径）
	PathHash         string `json:"pathHash" form:"pathHash"`         // 路径哈希值，用于快速查找
	Content          string `json:"content" form:"content"`           // 内容详情（完整文本）
	ContentHash      string `json:"contentHash" form:"contentHash"`   // 内容哈希，用于判定内容是否变更
	Ctime            int64  `json:"ctime" form:"ctime"`               // 创建时间戳（秒）
	Mtime            int64  `json:"mtime" form:"mtime"`               // 文件修改时间戳（秒）
	UpdatedTimestamp int64  `json:"lastTime" form:"updatedTimestamp"` // 记录更新时间戳（用于同步）
}

// NoteSyncEndMessage 同步结束时返回的信息结构。
type NoteSyncEndMessage struct {
	LastTime           int64 `json:"lastTime" form:"lastTime"`                     // 本次同步更新时间
	NeedUploadCount    int64 `json:"needUploadCount" form:"needUploadCount"`       // 需要上传的笔记数量
	NeedModifyCount    int64 `json:"needModifyCount" form:"needModifyCount"`       // 需要修改的笔记数量
	NeedSyncMtimeCount int64 `json:"needSyncMtimeCount" form:"needSyncMtimeCount"` // 需要同步修改时间的笔记数量
	NeedDeleteCount    int64 `json:"needDeleteCount" form:"needDeleteCount"`       // 需要删除的笔记数量
}

// NoteSyncNeedPushMessage 服务端告知客户端需要推送的文件信息。
type NoteSyncNeedPushMessage struct {
	Path string `json:"path" form:"path"` // 路径
}

// NoteSyncMtimeMessage 同步时用于更新 mtime 的消息结构。
type NoteSyncMtimeMessage struct {
	Path  string `json:"path" form:"path"`   // 路径
	Ctime int64  `json:"ctime" form:"ctime"` // 创建时间戳
	Mtime int64  `json:"mtime" form:"mtime"` // 修改时间戳
}

type NoteDeleteMessage struct {
	Path string `json:"path" form:"path"` // 路径信息（文件路径）
}

type NoteRenameMessage struct {
	Vault   string `json:"vault" form:"vault" binding:"required"`     // 仓库标识
	Path    string `json:"path" form:"path" binding:"required"`       // 新路径
	OldPath string `json:"oldPath" form:"oldPath" binding:"required"` // 旧路径
}

// NoteModify 处理文件修改的 WebSocket 消息
// 函数名: NoteModify
// 函数使用说明: 处理客户端发送的笔记修改或创建消息，进行参数校验、更新检查并在需要时写回数据库或通知其他客户端。
// 参数说明:
//   - c *app.WebsocketClient: 当前 WebSocket 客户端连接，包含上下文、用户信息、发送响应等能力。
//   - msg *app.WebSocketMessage: 接收到的 WebSocket 消息，包含消息数据和类型。
//
// 返回值说明:
//   - 无
func NoteModify(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.NoteModifyOrCreateRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("websocket_router.note.NoteModify.BindAndValid errs: %v", zap.Error(errs))
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

	svc := service.New(c.Ctx).WithSF(c.SF).WithClientName(c.ClientName).WithClientVersion(c.ClientVersion)

	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	checkParams := convert.StructAssign(params, &service.NoteUpdateCheckRequestParams{}).(*service.NoteUpdateCheckRequestParams)
	updateMode, nodeCheck, err := svc.NoteUpdateCheck(c.User.UID, checkParams)

	if err != nil {
		c.ToResponse(code.ErrorNoteModifyOrCreateFailed.WithDetails(err.Error()))
		return
	}

	switch updateMode {
	case "UpdateContent", "Create":

		_, note, err := svc.NoteModifyOrCreate(c.User.UID, params, true)
		if err != nil {
			c.ToResponse(code.ErrorNoteModifyOrCreateFailed.WithDetails(err.Error()))
			return
		}

		// 通知所有客户端更新mtime
		noteMessage := &NoteMessage{
			Path:             note.Path,
			PathHash:         note.PathHash,
			Content:          note.Content,
			ContentHash:      note.ContentHash,
			Ctime:            note.Ctime,
			Mtime:            note.Mtime,
			UpdatedTimestamp: note.UpdatedTimestamp,
		}

		c.ToResponse(code.Success.Reset())
		c.BroadcastResponse(code.Success.Reset().WithData(noteMessage).WithVault(params.Vault), true, "NoteSyncModify")
		return

	case "UpdateMtime":
		// 通知 客户端 Note 修改时间更新
		noteSyncMtimeMessage := &NoteSyncMtimeMessage{
			Path:  nodeCheck.Path,
			Ctime: nodeCheck.Ctime,
			Mtime: nodeCheck.Mtime,
		}
		c.ToResponse(code.Success.WithData(noteSyncMtimeMessage), "NoteSyncMtime")
		return
	default:
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
func NoteModifyCheck(c *app.WebsocketClient, msg *app.WebSocketMessage) {

	params := &service.NoteUpdateCheckRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("websocket_router.note.NoteModify.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF).WithClientName(c.ClientName).WithClientVersion(c.ClientVersion)

	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	updateMode, nodeCheck, err := svc.NoteUpdateCheck(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorNoteUpdateCheckFailed.WithDetails(err.Error()))
		return
	}

	// 通知客户端上传笔记
	switch updateMode {
	case "UpdateContent", "Create":
		noteSyncNeedPushMessage := &NoteSyncNeedPushMessage{
			Path: nodeCheck.Path,
		}
		c.ToResponse(code.Success.WithData(noteSyncNeedPushMessage), "NoteSyncNeedPush")
		return
	case "UpdateMtime":
		// 强制客户端更新mtime 不传输笔记内容
		noteSyncMtimeMessage := &NoteSyncMtimeMessage{
			Path:  nodeCheck.Path,
			Ctime: nodeCheck.Ctime,
			Mtime: nodeCheck.Mtime,
		}
		c.ToResponse(code.Success.WithData(noteSyncMtimeMessage), "NoteSyncMtime")
		return
	default:
		c.ToResponse(code.SuccessNoUpdate.Reset())
		return
	}
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
func NoteDelete(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.NoteDeleteRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("websocket_router.note.NoteDelete.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}
	handleNoteDelete(c, params)
}

func handleNoteDelete(c *app.WebsocketClient, params *service.NoteDeleteRequestParams) {

	svc := service.New(c.Ctx).WithSF(c.SF).WithClientName(c.ClientName).WithClientVersion(c.ClientVersion)

	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	note, err := svc.NoteDelete(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorNoteDeleteFailed.WithDetails(err.Error()))
		return
	}

	c.ToResponse(code.Success)
	c.BroadcastResponse(code.Success.WithData(note).WithVault(params.Vault), true, "NoteSyncDelete")
}

// NoteRename 处理文件重命名的 WebSocket 消息
// 函数名: NoteRename
// 函数使用说明: 接收客户端的笔记重命名请求，执行重命名操作，并通知所有客户端同步删除旧路径和创建新路径。
// 参数说明:
//   - c *app.WebsocketClient: 当前 WebSocket 客户端连接。
//   - msg *app.WebSocketMessage: 接收到的重命名请求消息。
//
// 返回值说明:
//   - 无
func NoteRename(c *app.WebsocketClient, msg *app.WebSocketMessage) {

	//先创建
	NoteModify(c, msg)

	//从 修改 里的可选参数里拿出 rename 参数
	params := &service.NoteModifyOrCreateRequestParams{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("websocket_router.note.NoteRename.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	handleNoteDelete(c, &service.NoteDeleteRequestParams{
		Vault:    params.Vault,
		Path:     params.OldPath,
		PathHash: params.OldPathHash,
	})

	svc := service.New(c.Ctx).WithSF(c.SF).WithClientName(c.ClientName).WithClientVersion(c.ClientVersion)

	err := svc.NoteRename(c.User.UID, &service.NoteRenameRequestParams{
		Vault:       params.Vault,
		Path:        params.Path,
		PathHash:    params.PathHash,
		OldPath:     params.OldPath,
		OldPathHash: params.OldPathHash,
	})

	if err != nil {
		c.ToResponse(code.ErrorNoteRenameFailed.WithDetails(err.Error()))
		return
	}
	// 相应成功
	c.ToResponse(code.Success.Reset())

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
func NoteSync(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.NoteSyncRequestParams{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("websocket_router.note.NoteModify.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF).WithClientName(c.ClientName).WithClientVersion(c.ClientVersion)

	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	list, err := svc.NoteListByLastTime(c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorNoteListFailed.WithDetails(err.Error()))
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
	var needUploadCount int64
	var needModifyCount int64
	var needSyncMtimeCount int64
	var needDeleteCount int64

	for _, note := range list {
		if note.UpdatedTimestamp >= lastTime {
			lastTime = note.UpdatedTimestamp
		}
		if note.Action == "delete" {
			// 客户端有,服务端已经删除, 通知客户端删除
			if _, ok := cNotes[note.PathHash]; ok {
				delete(cNotesKeys, note.PathHash)
				noteDeleteMessage := &NoteDeleteMessage{
					Path: note.Path,
				}
				c.ToResponse(code.Success.WithData(noteDeleteMessage), "NoteSyncDelete")
				needDeleteCount++
			}
		} else {
			//检查客户端是否有
			if cNote, ok := cNotes[note.PathHash]; ok {

				delete(cNotesKeys, note.PathHash)

				if note.ContentHash == cNote.ContentHash && note.Mtime == cNote.Mtime {
					//内容和修改时间一致, 跳过
					continue
				} else if note.ContentHash != cNote.ContentHash {
					// 内容不一致
					if cNote.Mtime < note.Mtime {

						switch c.OfflineSyncStrategy {
						//当忽略时间并合并时,登记需要合并的, 通知客户端上传笔记
						case "ignoreTimeMerge":

							c.DiffMergePathsMu.Lock()
							c.DiffMergePaths[note.Path] = note.Path
							c.DiffMergePathsMu.Unlock()

							noteSyncNeedPushMessage := &NoteSyncNeedPushMessage{
								Path: note.Path,
							}
							c.ToResponse(code.Success.Reset().WithData(noteSyncNeedPushMessage), "NoteSyncNeedPush")
							needUploadCount++
						// 当设置新笔记才进行合并, 因为本地笔记比较老, 服务器通知客户端使用云端笔记覆盖本地
						// 不设置 默认也一样覆盖
						case "newTimeMerge", "":
							noteMessage := &NoteMessage{
								Path:             note.Path,
								PathHash:         note.PathHash,
								Content:          note.Content,
								ContentHash:      note.ContentHash,
								Ctime:            note.Ctime,
								Mtime:            note.Mtime,
								UpdatedTimestamp: note.UpdatedTimestamp,
							}

							c.ToResponse(code.Success.Reset().WithData(noteMessage), "NoteSyncModify")
							needModifyCount++
						}
						// 服务端修改时间比客户端新, 通知客户端更新笔记

					} else {
						// 客户端笔记 比服务端笔记新, 通知客户端上传笔记
						if c.OfflineSyncStrategy == "ignoreTimeMerge" || c.OfflineSyncStrategy == "newTimeMerge" {
							c.DiffMergePathsMu.Lock()
							c.DiffMergePaths[note.Path] = note.Path
							c.DiffMergePathsMu.Unlock()
						}

						noteSyncNeedPushMessage := &NoteSyncNeedPushMessage{
							Path: note.Path,
						}
						c.ToResponse(code.Success.Reset().WithData(noteSyncNeedPushMessage), "NoteSyncNeedPush")
						needUploadCount++
					}
				} else {
					// 内容一致, 但修改时间不一致, 通知客户端更新笔记修改时间
					noteSyncMtimeMessage := &NoteSyncMtimeMessage{
						Path:  note.Path,
						Ctime: note.Ctime,
						Mtime: note.Mtime,
					}
					c.ToResponse(code.Success.WithData(noteSyncMtimeMessage), "NoteSyncMtime")
					needSyncMtimeCount++
				}
			} else {
				// 客户端没有的文件, 通知客户端创建文件
				noteMessage := &NoteMessage{
					Path:             note.Path,
					PathHash:         note.PathHash,
					Content:          note.Content,
					ContentHash:      note.ContentHash,
					Ctime:            note.Ctime,
					Mtime:            note.Mtime,
					UpdatedTimestamp: note.UpdatedTimestamp,
				}
				c.ToResponse(code.Success.WithData(noteMessage), "NoteSyncModify")
				needModifyCount++
			}
		}
	}

	if list == nil {
		lastTime = timex.Now().UnixMilli()
	}
	if len(cNotesKeys) > 0 {
		for pathHash := range cNotesKeys {
			note := cNotes[pathHash]
			NoteCheck := convert.StructAssign(&note, &NoteSyncNeedPushMessage{}).(*NoteSyncNeedPushMessage)
			c.ToResponse(code.Success.WithData(NoteCheck), "NoteSyncNeedPush")
			needUploadCount++
		}
	}

	c.IsFirstSync = true

	message := &NoteSyncEndMessage{
		LastTime:           lastTime,
		NeedUploadCount:    needUploadCount,
		NeedModifyCount:    needModifyCount,
		NeedSyncMtimeCount: needSyncMtimeCount,
		NeedDeleteCount:    needDeleteCount,
	}
	c.ToResponse(code.Success.WithData(message).WithVault(params.Vault), "NoteSyncEnd")
}

// UserInfo 验证并获取用户信息
// 函数名: UserInfo
// 函数使用说明: 从 service 层获取用户信息并转换成 WebSocket 需要的 UserSelectEntity 结构体（用于 WebSocket 用户验证）。
// 参数说明:
//   - c *app.WebsocketClient: 当前 WebSocket 客户端连接，包含上下文与服务工厂（SF）。
//   - uid int64: 要查询的用户 ID。
//
// 返回值说明:
//   - *app.UserSelectEntity: 如果查询到用户则返回转换后的用户实体，否则返回 nil。
//   - error: 查询过程中的错误（若有）。
func UserInfo(c *app.WebsocketClient, uid int64) (*app.UserSelectEntity, error) {

	svc := service.New(c.Ctx).WithSF(c.SF)
	user, err := svc.UserInfo(uid)

	var userEntity *app.UserSelectEntity
	if user != nil {
		userEntity = convert.StructAssign(user, &app.UserSelectEntity{}).(*app.UserSelectEntity)
	}

	return userEntity, err

}
