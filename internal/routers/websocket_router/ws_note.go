package websocket_router

import (
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/diff"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
)

// NoteWSHandler WebSocket 笔记处理器
// 使用 App Container 注入依赖
type NoteWSHandler struct {
	*WSHandler
}

// NewNoteWSHandler 创建 NoteWSHandler 实例
func NewNoteWSHandler(a *app.App) *NoteWSHandler {
	return &NoteWSHandler{
		WSHandler: NewWSHandler(a),
	}
}

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
	LastTime           int64           `json:"lastTime" form:"lastTime"`                     // 本次同步更新时间
	NeedUploadCount    int64           `json:"needUploadCount" form:"needUploadCount"`       // 需要上传的笔记数量
	NeedModifyCount    int64           `json:"needModifyCount" form:"needModifyCount"`       // 需要修改的笔记数量
	NeedSyncMtimeCount int64           `json:"needSyncMtimeCount" form:"needSyncMtimeCount"` // 需要同步修改时间的笔记数量
	NeedDeleteCount    int64           `json:"needDeleteCount" form:"needDeleteCount"`       // 需要删除的数量
	Messages           []queuedMessage `json:"messages"`                                     // 合并的消息队列
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
//   - c *pkgapp.WebsocketClient: 当前 WebSocket 客户端连接，包含上下文、用户信息、发送响应等能力。
//   - msg *pkgapp.WebSocketMessage: 接收到的 WebSocket 消息，包含消息数据和类型。
//
// 返回值说明:
//   - 无
func (h *NoteWSHandler) NoteModify(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.NoteModifyOrCreateRequest{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.logError(c, "websocket_router.note.NoteModify.BindAndValid", errs)
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

	pkgapp.NoteModifyLog(c.User.UID, "NoteModify", params.Path, params.Vault)

	ctx := c.Ctx.Request.Context()


	noteSvc := h.App.GetNoteService(c.ClientName, c.ClientVersion)

	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	checkParams := convert.StructAssign(params, &dto.NoteUpdateCheckRequest{}).(*dto.NoteUpdateCheckRequest)
	updateMode, nodeCheck, err := noteSvc.UpdateCheck(ctx, c.User.UID, checkParams)

	if err != nil {
		c.ToResponse(code.ErrorNoteModifyOrCreateFailed.WithDetails(err.Error()))
		return
	}

	switch updateMode {
	case "UpdateContent", "Create":

		var isExcludeSelf bool = true

		if c.OfflineSyncStrategy != "" && nodeCheck != nil && nodeCheck.ContentHash != params.BaseHash && nodeCheck.ContentHash != params.ContentHash {
			c.DiffMergePathsMu.RLock()
			_, ok := c.DiffMergePaths[params.Path]
			c.DiffMergePathsMu.RUnlock()

			// 如果是 diff 合并，需要跳过
			if ok {
				c.DiffMergePathsMu.Lock()
				delete(c.DiffMergePaths, params.Path)
				c.DiffMergePathsMu.Unlock()

				var history *service.NoteHistoryDTO
				// 如果 忽略时间合并， 如果内容在快照中找到 则 忽略掉 插件端更新
				if c.OfflineSyncStrategy == "ignoreTimeMerge" {
					history, err = h.App.NoteHistoryService.GetByNoteIDAndHash(ctx, c.User.UID, nodeCheck.ID, params.ContentHash)
					if err != nil {
						c.ToResponse(code.ErrorNoteModifyOrCreateFailed.WithDetails(err.Error()))
						return
					}
				}

				if history == nil {

					var baseContent string

					if params.BaseHash != params.ContentHash && params.BaseHash != "" {
						noteHistory, err := h.App.NoteHistoryService.GetByNoteIDAndHash(ctx, c.User.UID, nodeCheck.ID, params.BaseHash)
						if err != nil {
							c.ToResponse(code.ErrorNoteModifyOrCreateFailed.WithDetails(err.Error()))
							return
						}

						if noteHistory != nil {
							baseContent = noteHistory.Content
						} else {
							baseContent = nodeCheck.Content
						}
					} else {
						baseContent = nodeCheck.Content
					}

					clientContent := params.Content
					serverContent := nodeCheck.Content

					params.Content, err = diff.MergeTexts(baseContent, clientContent, serverContent, params.Mtime <= nodeCheck.Mtime)
					if err != nil {
						c.ToResponse(code.ErrorNoteModifyOrCreateFailed.WithDetails(err.Error()))
						return
					}
					params.ContentHash = util.EncodeHash32(params.Content)
					params.Mtime = timex.Now().Unix()

					isExcludeSelf = false
				}
			}
		}

		_, note, err := noteSvc.ModifyOrCreate(ctx, c.User.UID, params, true)
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

		c.ToResponse(code.Success)
		c.BroadcastResponse(code.Success.WithData(noteMessage).WithVault(params.Vault), isExcludeSelf, "NoteSyncModify")
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
		c.ToResponse(code.SuccessNoUpdate)
		return
	}
}

// NoteModifyCheck 检查文件修改必要性
// 函数名: NoteModifyCheck
// 函数使用说明: 仅用于检查客户端提供的笔记状态与服务器状态的差异，决定客户端是否需要上传笔记或只需同步 mtime。
// 参数说明:
//   - c *pkgapp.WebsocketClient: 当前 WebSocket 客户端连接，包含上下文和用户信息。
//   - msg *pkgapp.WebSocketMessage: 接收到的消息，包含需要检查的笔记信息。
//
// 返回值说明:
//   - 无
func (h *NoteWSHandler) NoteModifyCheck(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {

	params := &dto.NoteUpdateCheckRequest{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.logError(c, "websocket_router.note.NoteModifyCheck.BindAndValid", errs)
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	ctx := c.Ctx.Request.Context()


	noteSvc := h.App.GetNoteService(c.ClientName, c.ClientVersion)

	pkgapp.NoteModifyLog(c.User.UID, "NoteModifyCheck", params.Path, params.Vault)

	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	updateMode, nodeCheck, err := noteSvc.UpdateCheck(ctx, c.User.UID, params)

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
		c.ToResponse(code.SuccessNoUpdate)
		return
	}
}

// NoteDelete 处理文件删除的 WebSocket 消息
// 函数名: NoteDelete
// 函数使用说明: 接收客户端的笔记删除请求，执行删除操作并通知其他客户端同步删除事件。
// 参数说明:
//   - c *pkgapp.WebsocketClient: 当前 WebSocket 客户端连接，包含发送响应与广播能力。
//   - msg *pkgapp.WebSocketMessage: 接收到的删除请求消息，包含要删除的笔记标识等参数。
//
// 返回值说明:
//   - 无
func (h *NoteWSHandler) NoteDelete(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.NoteDeleteRequest{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.logError(c, "websocket_router.note.NoteDelete.BindAndValid", errs)
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	pkgapp.NoteModifyLog(c.User.UID, "NoteDelete", params.Path, params.Vault)

	h.handleNoteDelete(c, params)
}

func (h *NoteWSHandler) handleNoteDelete(c *pkgapp.WebsocketClient, params *dto.NoteDeleteRequest) {

	ctx := c.Ctx.Request.Context()


	noteSvc := h.App.GetNoteService(c.ClientName, c.ClientVersion)

	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	note, err := noteSvc.Delete(ctx, c.User.UID, params)

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
//   - c *pkgapp.WebsocketClient: 当前 WebSocket 客户端连接。
//   - msg *pkgapp.WebSocketMessage: 接收到的重命名请求消息。
//
// 返回值说明:
//   - 无
func (h *NoteWSHandler) NoteRename(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {

	//先创建
	h.NoteModify(c, msg)

	//从 修改 里的可选参数里拿出 rename 参数
	params := &dto.NoteModifyOrCreateRequest{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.logError(c, "websocket_router.note.NoteRename.BindAndValid", errs)
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	pkgapp.NoteModifyLog(c.User.UID, "NoteRename", params.Path, params.Vault)

	h.handleNoteDelete(c, &dto.NoteDeleteRequest{
		Vault:    params.Vault,
		Path:     params.OldPath,
		PathHash: params.OldPathHash,
	})

	ctx := c.Ctx.Request.Context()


	noteSvc := h.App.GetNoteService(c.ClientName, c.ClientVersion)

	err := noteSvc.Rename(ctx, c.User.UID, &dto.NoteRenameRequest{
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
	c.ToResponse(code.Success)

}

// NoteSync 处理全量或增量笔记同步
// 函数名: NoteSync
// 函数使用说明: 根据客户端提供的本地笔记列表与服务器端最近更新列表比较，决定返回哪些笔记需要上传、需要同步 mtime、需要删除或需要更新；最后返回同步结束消息。
// 参数说明:
//   - c *pkgapp.WebsocketClient: 当前 WebSocket 客户端连接，包含上下文与响应发送能力。
//   - msg *pkgapp.WebSocketMessage: 接收到的同步请求，包含客户端的笔记摘要和同步起始时间等信息。
//
// 返回值说明:
//   - 无
func (h *NoteWSHandler) NoteSync(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.NoteSyncRequest{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.logError(c, "websocket_router.note.NoteSync.BindAndValid", errs)
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	ctx := c.Ctx.Request.Context()


	noteSvc := h.App.GetNoteService(c.ClientName, c.ClientVersion)

	pkgapp.NoteModifyLog(c.User.UID, "NoteSync", "", params.Vault)

	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	list, err := noteSvc.ListByLastTime(ctx, c.User.UID, params)

	if err != nil {
		c.ToResponse(code.ErrorNoteListFailed.WithDetails(err.Error()))
		return
	}

	var cNotes map[string]dto.NoteSyncCheckRequest = make(map[string]dto.NoteSyncCheckRequest, 0)
	var cNotesKeys map[string]struct{} = make(map[string]struct{}, 0)

	if len(params.Notes) > 0 {
		for _, note := range params.Notes {
			cNotes[note.PathHash] = note
			cNotesKeys[note.PathHash] = struct{}{}
		}
	}

	// 创建消息队列，用于收集所有待发送的消息
	var messageQueue []queuedMessage

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
				// 将消息添加到队列而非立即发送
				messageQueue = append(messageQueue, queuedMessage{
					Action: "NoteSyncDelete",
					Data:   noteDeleteMessage,
				})
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
							c.DiffMergePaths[note.Path] = struct{}{}
							c.DiffMergePathsMu.Unlock()

							noteSyncNeedPushMessage := &NoteSyncNeedPushMessage{
								Path: note.Path,
							}
							// 将消息添加到队列而非立即发送
							messageQueue = append(messageQueue, queuedMessage{
								Action: "NoteSyncNeedPush",
								Data:   noteSyncNeedPushMessage,
							})
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

							// 将消息添加到队列而非立即发送
							messageQueue = append(messageQueue, queuedMessage{
								Action: "NoteSyncModify",
								Data:   noteMessage,
							})
							needModifyCount++
						}
						// 服务端修改时间比客户端新, 通知客户端更新笔记

					} else {
						// 客户端笔记 比服务端笔记新, 通知客户端上传笔记
						if c.OfflineSyncStrategy == "ignoreTimeMerge" || c.OfflineSyncStrategy == "newTimeMerge" {
							c.DiffMergePathsMu.Lock()
							c.DiffMergePaths[note.Path] = struct{}{}
							c.DiffMergePathsMu.Unlock()
						}

						noteSyncNeedPushMessage := &NoteSyncNeedPushMessage{
							Path: note.Path,
						}
						// 将消息添加到队列而非立即发送
						messageQueue = append(messageQueue, queuedMessage{
							Action: "NoteSyncNeedPush",
							Data:   noteSyncNeedPushMessage,
						})
						needUploadCount++
					}
				} else {
					// 内容一致, 但修改时间不一致, 通知客户端更新笔记修改时间
					noteSyncMtimeMessage := &NoteSyncMtimeMessage{
						Path:  note.Path,
						Ctime: note.Ctime,
						Mtime: note.Mtime,
					}
					// 将消息添加到队列而非立即发送
					messageQueue = append(messageQueue, queuedMessage{
						Action: "NoteSyncMtime",
						Data:   noteSyncMtimeMessage,
					})
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
				// 将消息添加到队列而非立即发送
				messageQueue = append(messageQueue, queuedMessage{
					Action: "NoteSyncModify",
					Data:   noteMessage,
				})
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

			data := NoteSyncNeedPushMessage{
				Path: note.Path,
			}

			// 将消息添加到队列而非立即发送
			messageQueue = append(messageQueue, queuedMessage{
				Action: "NoteSyncNeedPush",
				Data:   data,
			})

			needUploadCount++
		}
	}

	c.IsFirstSync = true

	// 发送 NoteSyncEnd 消息，包含所有合并的消息
	message := &NoteSyncEndMessage{
		LastTime:           lastTime,
		NeedUploadCount:    needUploadCount,
		NeedModifyCount:    needModifyCount,
		NeedSyncMtimeCount: needSyncMtimeCount,
		NeedDeleteCount:    needDeleteCount,
		Messages:           messageQueue,
	}
	c.ToResponse(code.Success.WithData(message).WithVault(params.Vault), "NoteSyncEnd")
}

// UserInfo 验证并获取用户信息
// 函数名: UserInfo
// 函数使用说明: 从 service 层获取用户信息并转换成 WebSocket 需要的 UserSelectEntity 结构体（用于 WebSocket 用户验证）。
// 参数说明:
//   - c *pkgapp.WebsocketClient: 当前 WebSocket 客户端连接，包含上下文与服务工厂（SF）。
//   - uid int64: 要查询的用户 ID。
//
// 返回值说明:
//   - *pkgapp.UserSelectEntity: 如果查询到用户则返回转换后的用户实体，否则返回 nil。
//   - error: 查询过程中的错误（若有）。
func (h *NoteWSHandler) UserInfo(c *pkgapp.WebsocketClient, uid int64) (*pkgapp.UserSelectEntity, error) {

	ctx := c.Ctx.Request.Context()
	user, err := h.App.UserService.GetInfo(ctx, uid)

	var userEntity *pkgapp.UserSelectEntity
	if user != nil {
		userEntity = convert.StructAssign(user, &pkgapp.UserSelectEntity{}).(*pkgapp.UserSelectEntity)
	}

	return userEntity, err

}
