package websocket_router

import (
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/diff"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"

	"go.uber.org/zap"
)

// NoteWSHandler WebSocket note handler
// NoteWSHandler WebSocket 笔记处理器
// Uses App Container to inject dependencies
// 使用 App Container 注入依赖
type NoteWSHandler struct {
	*WSHandler
}

// NewNoteWSHandler creates NoteWSHandler instance
// NewNoteWSHandler 创建 NoteWSHandler 实例
func NewNoteWSHandler(a *app.App) *NoteWSHandler {
	return &NoteWSHandler{
		WSHandler: NewWSHandler(a),
	}
}

// NoteModify handles WebSocket messages for file modification
// 函数名: NoteModify
// Function name: NoteModify
// usage: Handles note modification or creation messages sent by clients, performs parameter validation, update checks, and writes back to the database or notifies other clients when necessary.
// 函数使用说明: 处理客户端发送的笔记修改或创建消息，进行参数校验、更新检查并在需要时写回数据库或通知其他客户端。
// Parameters:
//   - c *pkgapp.WebsocketClient: Current WebSocket client connection, including context, user info, response sending capability, etc.
//
// 参数说明:
//   - c *pkgapp.WebsocketClient: 当前 WebSocket 客户端连接，包含上下文、用户信息、发送响应等能力。
//   - msg *pkgapp.WebSocketMessage: Received WebSocket message, containing message data and type.
//
// 参数说明:
//   - msg *pkgapp.WebSocketMessage: 接收到的 WebSocket 消息，包含消息数据和类型.
//
// Return:
//   - None
//
// 返回值说明:
//   - 无
func (h *NoteWSHandler) NoteModify(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.NoteModifyOrCreateRequest{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.note.NoteModify.BindAndValid")
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

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "NoteModify", params.Path, params.Vault)

	ctx := c.Context()

	noteSvc := h.App.GetNoteService(c.ClientName, c.ClientVersion)

	// Check and create vault, internally uses SF to merge concurrent requests, avoiding duplicate creation issues
	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	checkParams := convert.StructAssign(params, &dto.NoteUpdateCheckRequest{}).(*dto.NoteUpdateCheckRequest)
	updateMode, nodeCheck, err := noteSvc.UpdateCheck(ctx, c.User.UID, checkParams)

	if err != nil {
		h.respondError(c, code.ErrorNoteModifyOrCreateFailed, err, "websocket_router.note.NoteModify.UpdateCheck")
		return
	}

	switch updateMode {
	case "UpdateContent", "Create":

		var isExcludeSelf bool = true

		// Perform conflict detection when a note with the same name exists on the server
		// 当服务器存在同名笔记时，进行冲突检测
		if nodeCheck != nil {
			serverHash := nodeCheck.ContentHash
			baseHash := params.BaseHash
			contentHash := params.ContentHash

			// Skip update and return success (no update) to client when content hasn't changed
			// 当内容未变化时，跳过更新，给客户端返回成功(无更新)
			if serverHash == contentHash {

				h.App.Logger().Debug("server content equals client content, skipping update",
					zap.String(logger.FieldTraceID, c.TraceID),
					zap.Int64(logger.FieldUID, c.User.UID),
					zap.String(logger.FieldPath, params.Path),
					zap.String("contentHash", contentHash))
				c.ToResponse(code.SuccessNoUpdate)
				return
			}

			c.DiffMergePathsMu.RLock()
			_, mergeIsNeed := c.DiffMergePaths[params.Path]
			c.DiffMergePathsMu.RUnlock()

			if mergeIsNeed {
				c.DiffMergePathsMu.Lock()
				delete(c.DiffMergePaths, params.Path)
				c.DiffMergePathsMu.Unlock()

				// Skip merge and use client to override server directly if no offline sync strategy is set
				// 没有设置离线同步策略时，跳过合并，直接使用客户端覆盖服务端
				if c.OfflineSyncStrategy == "" {
					h.App.Logger().Debug("no offline sync strategy, skipping merge, using client to override server",
						zap.String(logger.FieldTraceID, c.TraceID),
						zap.Int64(logger.FieldUID, c.User.UID),
						zap.String(logger.FieldPath, params.Path))

					c.DiffMergePathsMu.Lock()
					delete(c.DiffMergePaths, params.Path)
					c.DiffMergePathsMu.Unlock()

					// Skip merge and use client to override server directly when server version is found to be an ancestor of client version
					// 当发现服务器版本是客户端版本的前身时，跳过合并，直接使用客户端覆盖服务端
				} else if serverHash == baseHash {
					h.App.Logger().Debug("server version is client version's ancestor, skipping merge, using client to override server",
						zap.String(logger.FieldTraceID, c.TraceID),
						zap.Int64(logger.FieldUID, c.User.UID),
						zap.String(logger.FieldPath, params.Path),
						zap.String("baseHash", baseHash))

					// Perform merge operation
					// 执行合并操作
					// case 1: baseHash is empty, client side creates new note, note with same name exists on server
					// case 1: baseHash 为空时，插件端 新创建笔记, 服务端存在同名笔记
					// case 2: baseHash is not empty, client side note and server side note have same base source, server side modification time is later than client side
					// case 2: baseHash 不为空时，插件端 笔记 和 服务端笔记 同一base源 , 服务端笔记版修改时间大于插件端
					// case 3: baseHash is not empty, client side note and server side note have same base source, server side modification time is earlier than client side
					// case 3: baseHash 不为空时，插件端 笔记 和 服务端笔记 同一base源 , 服务端笔记版修改时间小于插件端
					// case 4: baseHash is not empty, client side note and server side note from different base source, server side modification time is later than client side
					// case 4: baseHash 不为空时，插件端 笔记 和 服务端笔记 不同base源, 服务端笔记版修改时间大于插件端
					// case 5: baseHash is not empty, client side note and server side note from different base source, server side modification time is earlier than client side
					// case 5: baseHash 不为空时，插件端 笔记 和 服务端笔记 不同base源, 服务端笔记版修改时间小于插件端
					// Question 1: Some edited content matches server note snapshot but time dimension differs, they should not be identified as the same version
					// 问题1. 某编辑内容和服务器笔记快照一致 但是时间维度不一致 不应该将他们识别为同一版本
					// Question 2: Because historical snapshots are only generated every 30s... client side basehash has a high probability of not finding basehash, and we can't generate a snapshot for every change... too wasteful.
					// 问题2. 因为历史快照是 30s 才生成一份.. 导致插件端的basehash 有很大概率找不到 basehash, 又不能每次变更都生成一个快照.. 太浪费了
				} else {

					h.App.Logger().Info("potential merge conflict detected",
						zap.String(logger.FieldTraceID, c.TraceID),
						zap.Int64(logger.FieldUID, c.User.UID),
						zap.String(logger.FieldPath, params.Path),
						zap.String("serverHash", serverHash),
						zap.String("baseHash", baseHash),
						zap.String("contentHash", contentHash),
						zap.String("offlineSyncStrategy", c.OfflineSyncStrategy))

					// If it's a diff merge, perform merge logic
					// Note: Logic to skip merge based on contentHash matching historical snapshot has been removed
					// Reason: This logic caused valid user modifications to be silently discarded (when content happened to be same as some historical snapshot)
					// 如果是 diff 合并，需要执行合并逻辑
					// 注意：已移除基于 contentHash 匹配历史快照跳过合并的逻辑
					// 原因：该逻辑会导致用户有效修改被静默丢弃（当内容恰好与某个历史快照相同时）

					var baseContent string
					var baseHashNotFound bool

					// Find merge base version
					// When baseHash is valid and different from contentHash, try to find it in history
					// 查找合并基准版本
					// 当 baseHash 有效且与 contentHash 不同时，尝试从历史记录中查找
					if !params.BaseHashMissing {
						noteHistory, err := h.App.NoteHistoryService.GetByNoteIDAndHash(ctx, c.User.UID, nodeCheck.ID, baseHash)
						if err != nil {
							h.respondError(c, code.ErrorNoteModifyOrCreateFailed, err, "websocket_router.note.NoteModify.GetByNoteIDAndHash")
							return
						}

						if noteHistory != nil {
							baseContent = noteHistory.Content
						} else {
							// History record not found
							// 历史记录未找到
							h.App.Logger().Warn("history record not found for baseHash",
								zap.String(logger.FieldTraceID, c.TraceID),
								zap.Int64(logger.FieldUID, c.User.UID),
								zap.String(logger.FieldPath, params.Path),
								zap.String("baseHash", baseHash))
							baseHashNotFound = true
						}
					} else {
						// baseHash is empty or client marked as unavailable
						// baseHash 为空或客户端标记为不可用
						if baseHash == "" || params.BaseHashMissing {
							h.App.Logger().Warn("baseHash is empty or missing",
								zap.String(logger.FieldTraceID, c.TraceID),
								zap.Int64(logger.FieldUID, c.User.UID),
								zap.String(logger.FieldPath, params.Path),
								zap.Bool("baseHashMissing", params.BaseHashMissing))
							baseHashNotFound = true
						}
					}

					// When baseHash is not found, use server current content as base and continue merging
					// This usually happens when: another device goes online to sync during the delayed historical record creation (20s)
					// Using server content as base correctly merges in most scenarios
					// 当 baseHash 找不到时，使用服务端当前内容作为 base 继续合并
					// 这种情况通常发生在：历史记录延迟创建（20秒）期间另一设备上线同步
					// 使用服务端内容作为 base 在大多数场景下能正确合并
					if baseHashNotFound {
						h.App.Logger().Warn("baseHash not found, using server content as merge base",
							zap.String(logger.FieldTraceID, c.TraceID),
							zap.Int64(logger.FieldUID, c.User.UID),
							zap.String(logger.FieldPath, params.Path),
							zap.String("baseHash", baseHash),
							zap.Bool("baseHashMissing", params.BaseHashMissing))
						baseContent = nodeCheck.Content
					}

					clientContent := params.Content
					serverContent := nodeCheck.Content

					// Determine patch application order
					// ignoreTimeMerge strategy: ignore timestamp, fixed use client priority
					// When both sides modify different areas, result is consistent (patch application order doesn't affect)
					// When both sides modify same area, hasConflict will detect conflict and create conflict file
					// 确定 patch 应用顺序
					// ignoreTimeMerge 策略：忽略时间戳，固定使用客户端优先
					// 当两边修改不同区域时，结果一致（patch 应用顺序不影响）
					// 当两边修改同一区域时，hasConflict 会检测到冲突并创建冲突文件
					var pc1First bool
					if c.OfflineSyncStrategy == "ignoreTimeMerge" {
						pc1First = true
					} else {
						// Other strategies: use time to determine priority
						// 其他策略：使用时间决定优先级
						pc1First = params.Mtime <= nodeCheck.Mtime
					}

					var mergeResult diff.MergeResult
					if !baseHashNotFound {
						// Use text merge with conflict detection
						// 使用带冲突检测的文本检测
						mergeResult, err = diff.MergeTexts(baseContent, clientContent, serverContent, pc1First)
						if err != nil {
							h.respondError(c, code.ErrorNoteModifyOrCreateFailed, err, "websocket_router.note.NoteModify.MergeTexts")
							return
						}

						h.App.Logger().Info("merge completed",
							zap.String(logger.FieldTraceID, c.TraceID),
							zap.Int64(logger.FieldUID, c.User.UID),
							zap.String(logger.FieldPath, params.Path),
							zap.Bool("hasConflict", mergeResult.HasConflict),
							zap.Int("baseLen", len(baseContent)),
							zap.Int("clientLen", len(clientContent)),
							zap.Int("serverLen", len(serverContent)),
							zap.Int("resultLen", len(mergeResult.Content)),
							zap.Bool("pc1First", pc1First))
					}

					// Check if conflict exists, perform further merge operations
					// 检查是否存在冲突， 执行进一步合并操作
					if mergeResult.HasConflict || baseHashNotFound {

						// Notify user of merge conflict, need to handle redundant note content
						// 通知用户出现合并冲突, 需要处理笔记冗余内容
						// todo 先暂时不通知用户出现冲突
						// c.ToResponse(code.ErrorSyncConflict.WithData(dto.NoteSyncNeedPushMessage{
						// 	Path:     params.Path,
						// 	PathHash: params.PathHash,
						// }), dto.NoteSyncNeedPush)

						// Force merge to keep all text from PC1 and PC2
						// 强制合并 保留PC1 PC2全部文本
						mergeResult.Content, err = diff.MergeTextsIgnoreConflictIgnoreDelete(baseContent, clientContent, serverContent, pc1First)
						if err != nil {
							h.respondError(c, code.ErrorNoteModifyOrCreateFailed, err, "websocket_router.note.NoteModify.MergeTextsIgnoreConflictIgnoreDelete")
							return
						}

						// // 创建冲突文件保存客户端内容
						// conflictReq := &dto.ConflictFileRequest{
						// 	Vault:             params.Vault,
						// 	OriginalPath:      params.Path,
						// 	ClientContent:     params.Content,
						// 	ClientContentHash: params.ContentHash,
						// 	Ctime:             params.Ctime,
						// 	Mtime:             params.Mtime,
						// }

						// conflictResp, err := h.App.ConflictService.CreateConflictFile(ctx, c.User.UID, conflictReq)
						// if err != nil {
						// 	h.App.Logger().Error("failed to create conflict file",
						// 		zap.String(logger.FieldTraceID, c.TraceID),
						// 		zap.Int64(logger.FieldUID, c.User.UID),
						// 		zap.String(logger.FieldPath, params.Path),
						// 		zap.Error(err))
						// 	h.respondError(c, code.ErrorNoteModifyOrCreateFailed, err, "websocket_router.note.NoteModify.CreateConflictFile")
						// 	return
						// }

						// h.App.Logger().Info("merge conflict detected, conflict file created",
						// 	zap.String(logger.FieldTraceID, c.TraceID),
						// 	zap.Int64(logger.FieldUID, c.User.UID),
						// 	zap.String(logger.FieldPath, params.Path),
						// 	zap.String("conflictPath", conflictResp.ConflictPath),
						// 	zap.String("conflictInfo", mergeResult.ConflictInfo))

						// // 返回冲突文件创建成功的响应
						// c.ToResponse(code.ErrorConflictFileCreated.WithData(conflictResp))
						// return
					}

					params.Content = mergeResult.Content
					params.ContentHash = util.EncodeHash32(params.Content)
					params.Mtime = timex.Now().UnixMilli()

					isExcludeSelf = false

				}
			}

		}

		_, note, err := noteSvc.ModifyOrCreate(ctx, c.User.UID, params, true)
		if err != nil {
			h.respondError(c, code.ErrorNoteModifyOrCreateFailed, err, "websocket_router.note.NoteModify.ModifyOrCreate")
			return
		}

		// Notify all clients to update mtime
		// 通知所有客户端更新mtime

		c.ToResponse(code.Success)
		c.BroadcastResponse(code.Success.WithData(
			dto.NoteSyncModifyMessage{
				Path:             note.Path,
				PathHash:         note.PathHash,
				Content:          note.Content,
				ContentHash:      note.ContentHash,
				Ctime:            note.Ctime,
				Mtime:            note.Mtime,
				UpdatedTimestamp: note.UpdatedTimestamp,
			},
		).WithVault(params.Vault), isExcludeSelf, dto.NoteSyncModify)
		return

	case "UpdateMtime":
		// Notify client of note modification time update
		// 通知 客户端 Note 修改时间更新
		c.ToResponse(code.Success.WithData(
			dto.NoteSyncMtimeMessage{
				Path:  nodeCheck.Path,
				Ctime: nodeCheck.Ctime,
				Mtime: nodeCheck.Mtime,
			},
		).WithVault(params.Vault), dto.NoteSyncMtime)
		return
	default:
		c.ToResponse(code.SuccessNoUpdate)
		return
	}
}

// NoteModifyCheck checks the necessity of file modification
// 函数名: NoteModifyCheck
// Function name: NoteModifyCheck
// usage: Only used to check difference between note status provided by client and server status, deciding if client needs to upload note or just sync mtime.
// 函数使用说明: 仅用于检查客户端提供的笔记状态与服务器状态的差异，决定客户端是否需要上传笔记或只需同步 mtime。
// Parameters:
//   - c *pkgapp.WebsocketClient: Current WebSocket client connection, including context and user info.
//
// 参数说明:
//   - c *pkgapp.WebsocketClient: 当前 WebSocket 客户端连接，包含上下文和用户信息。
//   - msg *pkgapp.WebSocketMessage: Received message, containing note info needing check.
//
// 参数说明:
//   - msg *pkgapp.WebSocketMessage: 接收到的消息，包含需要检查的笔记信息。
//
// Return:
//   - None
//
// 返回值说明:
//   - 无
func (h *NoteWSHandler) NoteModifyCheck(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {

	params := &dto.NoteUpdateCheckRequest{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.note.NoteModifyCheck.BindAndValid")
		return
	}

	ctx := c.Context()

	noteSvc := h.App.GetNoteService(c.ClientName, c.ClientVersion)

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "NoteModifyCheck", params.Path, params.Vault)

	// Check and create vault, internally uses SF to merge concurrent requests, avoiding duplicate creation issues
	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	updateMode, nodeCheck, err := noteSvc.UpdateCheck(ctx, c.User.UID, params)

	if err != nil {
		h.respondError(c, code.ErrorNoteUpdateCheckFailed, err, "websocket_router.note.NoteModifyCheck.UpdateCheck")
		return
	}

	// Notify client to upload note
	// 通知客户端上传笔记
	switch updateMode {
	case "UpdateContent", "Create":
		c.ToResponse(code.Success.WithData(
			dto.NoteSyncNeedPushMessage{
				Path:     nodeCheck.Path,
				PathHash: nodeCheck.PathHash,
			},
		).WithVault(params.Vault), dto.NoteSyncNeedPush)
		return
	case "UpdateMtime":
		// Force client to update mtime without transferring note content
		// 强制客户端更新mtime 不传输笔记内容
		c.ToResponse(code.Success.WithData(
			dto.NoteSyncMtimeMessage{
				Path:  nodeCheck.Path,
				Ctime: nodeCheck.Ctime,
				Mtime: nodeCheck.Mtime,
			},
		).WithVault(params.Vault), dto.NoteSyncMtime)
		return
	default:
		c.ToResponse(code.SuccessNoUpdate)
		return
	}
}

// NoteDelete handles WebSocket messages for file deletion
// 函数名: NoteDelete
// Function name: NoteDelete
// usage: Receives client note deletion request, performs deletion, and notifies other clients to sync deletion events.
// 函数使用说明: 接收客户端的笔记删除请求，执行删除操作并通知其他客户端同步删除事件。
// Parameters:
//   - c *pkgapp.WebsocketClient: Current WebSocket client connection, including response sending and broadcasting capabilities.
//
// 参数说明:
//   - c *pkgapp.WebsocketClient: 当前 WebSocket 客户端连接，包含发送响应与广播能力。
//   - msg *pkgapp.WebSocketMessage: Received deletion request message, containing parameters like note identifier to delete.
//
// 参数说明:
//   - msg *pkgapp.WebSocketMessage: 接收到的删除请求消息，包含要删除的笔记标识等参数。
//
// Return:
//   - None
//
// 返回值说明:
//   - 无
func (h *NoteWSHandler) NoteDelete(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.NoteDeleteRequest{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.note.NoteDelete.BindAndValid")
		return
	}

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "NoteDelete", params.Path, params.Vault)

	ctx := c.Context()

	noteSvc := h.App.GetNoteService(c.ClientName, c.ClientVersion)

	// Check and create vault, internally uses SF to merge concurrent requests, avoiding duplicate creation issues
	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	note, err := noteSvc.Delete(ctx, c.User.UID, params)

	if err != nil {
		h.respondError(c, code.ErrorNoteDeleteFailed, err, "websocket_router.note.handleNoteDelete.Delete")
		return
	}

	c.ToResponse(code.Success)
	c.BroadcastResponse(code.Success.WithData(
		dto.NoteSyncDeleteMessage{
			Path:     note.Path,
			PathHash: note.PathHash,
			Ctime:    note.Ctime,
			Mtime:    note.Mtime,
			Size:     note.Size,
		},
	).WithVault(params.Vault), true, dto.NoteSyncDelete)
}

// NoteRename handles WebSocket messages for file renaming
// 函数名: NoteRename
// Function name: NoteRename
// usage: Receives client note renaming request, performs renaming, and notifies all clients to sync old path deletion and new path creation.
// 函数使用说明: 接收客户端的笔记重命名请求，执行重命名操作，并通知所有客户端同步删除旧路径和创建新路径。
// Parameters:
//   - c *pkgapp.WebsocketClient: Current WebSocket client connection.
//
// 参数说明:
//   - c *pkgapp.WebsocketClient: 当前 WebSocket 客户端连接。
//   - msg *pkgapp.WebSocketMessage: Received renaming request message.
//
// 参数说明:
//   - msg *pkgapp.WebSocketMessage: 接收到的重命名请求消息。
//
// Return:
//   - None
//
// 返回值说明:
//   - 无
func (h *NoteWSHandler) NoteRename(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.NoteRenameRequest{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.note.NoteRename.BindAndValid")
		return
	}

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "NoteRename", params.Path, params.Vault)

	uid := c.User.UID
	oldNote, newNote, err := h.App.GetNoteService(c.ClientName, c.ClientVersion).Rename(c.Context(), uid, params)
	if err != nil {
		h.respondError(c, code.ErrorRenameNoteTargetExist, err, "websocket_router.note.NoteRename.Rename")
		return
	}

	c.ToResponse(code.Success)
	c.BroadcastResponse(code.Success.WithData(
		dto.NoteSyncRenameMessage{
			Path:        newNote.Path,
			PathHash:    newNote.PathHash,
			ContentHash: newNote.ContentHash,
			Ctime:       newNote.Ctime,
			Mtime:       newNote.Mtime,
			Size:        newNote.Size,
			OldPath:     oldNote.Path,
			OldPathHash: oldNote.PathHash,
		},
	).WithVault(params.Vault), true, dto.NoteSyncRename)
}

func (h *NoteWSHandler) NoteRePush(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.NoteGetRequest{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.note.NoteReceiveMissing.BindAndValid")
		return
	}

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "NoteRePush", params.Path, params.Vault)

	uid := c.User.UID
	note, err := h.App.GetNoteService(c.ClientName, c.ClientVersion).Get(c.Context(), uid, params)
	if err != nil {
		h.respondError(c, code.ErrorNoteNotFound, err, "websocket_router.note.NoteRePush.Get")
		return
	}

	if note != nil && note.Action != "delete" {
		c.ToResponse(code.Success.WithData(
			dto.NoteSyncModifyMessage{
				Path:             note.Path,
				PathHash:         note.PathHash,
				Content:          note.Content,
				ContentHash:      note.ContentHash,
				Ctime:            note.Ctime,
				Mtime:            note.Mtime,
				UpdatedTimestamp: note.UpdatedTimestamp,
			},
		).WithVault(params.Vault), dto.NoteSyncModify)
	} else {
		c.ToResponse(code.ErrorNoteNotFound)
	}

}

// NoteSync handles full or incremental note sync
// 函数名: NoteSync
// Function name: NoteSync
// usage: Compares local note list provided by client with server side recent update list, deciding which notes need uploading, mtime sync, deletion or update; finally returns sync end message.
// 函数使用说明: 根据客户端提供的本地笔记列表与服务器端最近更新列表比较，决定返回哪些笔记需要上传、需要同步 mtime、需要删除或需要更新；最后返回同步结束消息。
// Parameters:
//   - c *pkgapp.WebsocketClient: Current WebSocket client connection, including context and response sending capability.
//
// 参数说明:
//   - c *pkgapp.WebsocketClient: 当前 WebSocket 客户端连接，包含上下文与响应发送能力。
//   - msg *pkgapp.WebSocketMessage: Received sync request, containing client note digest and sync start time, etc.
//
// 参数说明:
//   - msg *pkgapp.WebSocketMessage: 接收到的同步请求，包含客户端的笔记摘要和同步起始时间等信息。
//
// Return:
//   - None
//
// 返回值说明:
//   - 无
func (h *NoteWSHandler) NoteSync(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.NoteSyncRequest{}

	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.note.NoteSync.BindAndValid")
		return
	}

	ctx := c.Context()

	noteSvc := h.App.GetNoteService(c.ClientName, c.ClientVersion)

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "NoteSync", "", params.Vault)

	// Check and create vault, internally uses SF to merge concurrent requests, avoiding duplicate creation issues
	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	list, err := noteSvc.ListByLastTime(ctx, c.User.UID, params)

	if err != nil {
		h.respondError(c, code.ErrorNoteListFailed, err, "websocket_router.note.NoteSync.ListByLastTime")
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

	// Create message queue for collecting all messages to be sent
	// 创建消息队列，用于收集所有待发送的消息
	var messageQueue []dto.WSQueuedMessage

	var lastTime int64
	var needUploadCount int64
	var needModifyCount int64
	var needSyncMtimeCount int64
	var needDeleteCount int64

	var cDelNotesKeys map[string]struct{} = make(map[string]struct{}, 0)

	// Handle notes deleted by client
	// 处理客户端删除的笔记
	if len(params.DelNotes) > 0 {
		for _, delNote := range params.DelNotes {
			// Check if note exists before deleting
			// 删除前检查笔记是否存在
			getCheckParams := &dto.NoteGetRequest{
				Vault:    params.Vault,
				PathHash: delNote.PathHash,
			}
			checkNote, err := noteSvc.Get(ctx, c.User.UID, getCheckParams)

			// If note exists, execute delete
			// 如果笔记存在，执行删除
			if err == nil && checkNote != nil && checkNote.Action != "delete" {
				delParams := &dto.NoteDeleteRequest{
					Vault:    params.Vault,
					Path:     delNote.Path,
					PathHash: delNote.PathHash,
				}
				note, err := noteSvc.Delete(ctx, c.User.UID, delParams)
				if err != nil {
					h.App.Logger().Error("websocket_router.note.NoteSync.noteSvc.Delete",
						zap.String(logger.FieldTraceID, c.TraceID),
						zap.Int64(logger.FieldUID, c.User.UID),
						zap.String(logger.FieldPath, delNote.Path),
						zap.Error(err))
					continue
				}

				// Record PathHash deleted by client to avoid duplicate sending
				// 记录客户端已主动删除的 PathHash，避免重复下发
				cDelNotesKeys[delNote.PathHash] = struct{}{}

				// Broadcast deletion to other clients
				// 将删除消息广播给其他客户端
				c.BroadcastResponse(code.Success.WithData(
					dto.NoteSyncDeleteMessage{
						Path:     note.Path,
						PathHash: note.PathHash,
						Ctime:    note.Ctime,
						Mtime:    note.Mtime,
						Size:     note.Size,
					},
				).WithVault(params.Vault), true, dto.NoteSyncDelete)

			} else {
				// Note does not exist, but we still need to record exclusion and broadcast delete message to ensure data consistency
				// 笔记不存在，但仍需记录排除并广播删除消息，以确保数据一致性

				h.App.Logger().Debug("websocket_router.note.NoteSync.noteSvc.Get check failed (not found or already deleted), broadcasting delete anyway",
					zap.String(logger.FieldTraceID, c.TraceID),
					zap.String("pathHash", delNote.PathHash))

				// Record PathHash
				// 记录 PathHash
				cDelNotesKeys[delNote.PathHash] = struct{}{}

				// Broadcast deletion with available info (Path/PathHash)
				// 使用现有信息(Path/PathHash)广播删除
				c.BroadcastResponse(code.Success.WithData(
					dto.NoteSyncDeleteMessage{
						Path:     delNote.Path,
						PathHash: delNote.PathHash,
						Ctime:    0,
						Mtime:    0,
						Size:     0,
					},
				).WithVault(params.Vault), true, dto.NoteSyncDelete)
			}
		}
	}

	// Handle notes missing on client (only for incremental sync)
	// 处理客户端缺失的笔记（仅限增量同步）
	if params.LastTime > 0 && len(params.MissingNotes) > 0 {
		for _, missingNote := range params.MissingNotes {
			getParams := &dto.NoteGetRequest{
				Vault:    params.Vault,
				Path:     missingNote.Path,
				PathHash: missingNote.PathHash,
			}
			note, err := noteSvc.Get(ctx, c.User.UID, getParams)
			if err != nil {
				h.App.Logger().Warn("websocket_router.note.NoteSync.noteSvc.Get",
					zap.String(logger.FieldTraceID, c.TraceID),
					zap.Int64(logger.FieldUID, c.User.UID),
					zap.String("path", missingNote.Path),
					zap.String("pathHash", missingNote.PathHash),
					zap.Error(err))
				continue
			}
			if note != nil && note.Action != "delete" {
				messageQueue = append(messageQueue, dto.WSQueuedMessage{
					Action: dto.NoteSyncModify,
					Data: dto.NoteSyncModifyMessage{
						Path:             note.Path,
						PathHash:         note.PathHash,
						Content:          note.Content,
						ContentHash:      note.ContentHash,
						Ctime:            note.Ctime,
						Mtime:            note.Mtime,
						UpdatedTimestamp: note.UpdatedTimestamp,
					},
				})
				needModifyCount++
				// 加入排除索引
				cDelNotesKeys[note.PathHash] = struct{}{}
			}
		}
	}

	for _, note := range list {
		// 如果该笔记是客户端刚才通过参数告知删除的，则跳过下发
		if _, ok := cDelNotesKeys[note.PathHash]; ok {
			continue
		}

		if note.UpdatedTimestamp >= lastTime {
			lastTime = note.UpdatedTimestamp
		}
		if note.Action == "delete" {
			// Client has it, server already deleted, notify client to delete
			// Client has it, server already deleted, notify client to delete
			// 客户端有,服务端已经删除, 通知客户端删除
			if _, ok := cNotes[note.PathHash]; ok {
				delete(cNotesKeys, note.PathHash)
				// 将消息添加到队列而非立即发送
				messageQueue = append(messageQueue, dto.WSQueuedMessage{
					Action: dto.NoteSyncDelete,
					Data: dto.NoteSyncDeleteMessage{
						Path:     note.Path,
						PathHash: note.PathHash,
						Ctime:    note.Ctime,
						Mtime:    note.Mtime,
						Size:     note.Size,
					},
				})
				needDeleteCount++
			}
		} else {
			// Check if client has it
			//检查客户端是否有
			if cNote, ok := cNotes[note.PathHash]; ok {

				delete(cNotesKeys, note.PathHash)

				if note.ContentHash == cNote.ContentHash && note.Mtime == cNote.Mtime {
					// Content and modification time match, skip
					//内容和修改时间一致, 跳过
					continue
				} else if note.ContentHash != cNote.ContentHash {
					// Content inconsistent
					// 内容不一致
					if cNote.Mtime < note.Mtime {

						switch c.OfflineSyncStrategy {
						// When ignore time and merge, register those needing merge, notify client to upload note
						//当忽略时间并合并时,登记需要合并的, 通知客户端上传笔记
						case "ignoreTimeMerge":

							c.DiffMergePathsMu.Lock()
							c.DiffMergePaths[note.Path] = pkgapp.DiffMergeEntry{CreatedAt: time.Now()}
							c.DiffMergePathsMu.Unlock()

							// Add message to queue instead of sending immediately
							// 将消息添加到队列而非立即发送
							messageQueue = append(messageQueue, dto.WSQueuedMessage{
								Action: dto.NoteSyncNeedPush,
								Data: dto.NoteSyncNeedPushMessage{
									Path:     note.Path,
									PathHash: note.PathHash,
								},
							})
							needUploadCount++
						// When only new notes are merged, since local note is older, server notifies client to override local with cloud note
						// Don't set, default override as well
						// 当设置新笔记才进行合并, 因为本地笔记比较老, 服务器通知客户端使用云端笔记覆盖本地
						// 不设置 默认也一样覆盖
						case "newTimeMerge", "":
							// 将消息添加到队列而非立即发送
							messageQueue = append(messageQueue, dto.WSQueuedMessage{
								Action: dto.NoteSyncModify,
								Data: dto.NoteSyncModifyMessage{
									Path:             note.Path,
									PathHash:         note.PathHash,
									Content:          note.Content,
									ContentHash:      note.ContentHash,
									Ctime:            note.Ctime,
									Mtime:            note.Mtime,
									UpdatedTimestamp: note.UpdatedTimestamp,
								},
							})
							needModifyCount++
						}
						// Server modification time is newer than client, notify client to update note
						// 服务端修改时间比客户端新, 通知客户端更新笔记

					} else {
						// Client note is newer than server, notify client to upload note
						// Offline sync strategy description:
						// - ignoreTimeMerge: ignore timestamp, always execute three-way merge, need to register to DiffMergePaths
						// - newTimeMerge: new time priority, register DiffMergePaths
						// 客户端笔记 比服务端笔记新, 通知客户端上传笔记
						// 离线同步策略说明：
						// - ignoreTimeMerge: 忽略时间戳，始终执行三方合并，需要登记到 DiffMergePaths
						// - newTimeMerge: 新时间优先, 登记 DiffMergePaths

						if c.OfflineSyncStrategy == "ignoreTimeMerge" || c.OfflineSyncStrategy == "newTimeMerge" {
							c.DiffMergePathsMu.Lock()
							c.DiffMergePaths[note.Path] = pkgapp.DiffMergeEntry{CreatedAt: time.Now()}
							c.DiffMergePathsMu.Unlock()
						}

						// Add message to queue instead of sending immediately
						// 将消息添加到队列而非立即发送
						messageQueue = append(messageQueue, dto.WSQueuedMessage{
							Action: dto.NoteSyncNeedPush,
							Data: dto.NoteSyncNeedPushMessage{
								Path:     note.Path,
								PathHash: note.PathHash,
							},
						})
						needUploadCount++
					}
				} else {
					// Content matches, but modification time differs, notify client to update note mtime
					// 内容一致, 但修改时间不一致, 通知客户端更新笔记修改时间
					// Add message to queue instead of sending immediately
					// 将消息添加到队列而非立即发送
					messageQueue = append(messageQueue, dto.WSQueuedMessage{
						Action: dto.NoteSyncMtime,
						Data: dto.NoteSyncMtimeMessage{
							Path:  note.Path,
							Ctime: note.Ctime,
							Mtime: note.Mtime,
						},
					})
					needSyncMtimeCount++
				}
			} else {
				// File client doesn't have, notify client to create file
				// 客户端没有的文件, 通知客户端创建文件
				// 将消息添加到队列而非立即发送
				messageQueue = append(messageQueue, dto.WSQueuedMessage{
					Action: dto.NoteSyncModify,
					Data: dto.NoteSyncModifyMessage{
						Path:             note.Path,
						PathHash:         note.PathHash,
						Content:          note.Content,
						ContentHash:      note.ContentHash,
						Ctime:            note.Ctime,
						Mtime:            note.Mtime,
						UpdatedTimestamp: note.UpdatedTimestamp,
					},
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

			// Add message to queue instead of sending immediately
			// 将消息添加到队列而非立即发送
			messageQueue = append(messageQueue, dto.WSQueuedMessage{
				Action: dto.NoteSyncNeedPush,
				Data: dto.NoteSyncNeedPushMessage{
					Path:     note.Path,
					PathHash: note.PathHash,
				},
			})

			needUploadCount++
		}
	}

	c.IsFirstSync = true

	// Send NoteSyncEnd message, containing all counts
	// 发送 NoteSyncEnd 消息，包含所有统计计数
	c.ToResponse(code.Success.WithData(
		dto.NoteSyncEndMessage{
			LastTime:           lastTime,
			NeedUploadCount:    needUploadCount,
			NeedModifyCount:    needModifyCount,
			NeedSyncMtimeCount: needSyncMtimeCount,
			NeedDeleteCount:    needDeleteCount,
		},
	).WithVault(params.Vault).WithContext(params.Context), dto.NoteSyncEnd)

	// After End message, send all queued messages individually
	// 在 End 消息后，逐条发送队列中的消息
	for _, item := range messageQueue {
		c.ToResponse(code.Success.WithData(item.Data).WithVault(params.Vault).WithContext(params.Context), item.Action)
	}
}

// UserInfo verifies and retrieves user info
// 函数名: UserInfo
// Function name: UserInfo
// usage: Retrieves user info from service layer and converts to UserSelectEntity structure needed by WebSocket (for WebSocket user verification).
// 函数使用说明: 从 service 层获取用户信息并转换成 WebSocket 需要的 UserSelectEntity 结构体（用于 WebSocket 用户验证）。
// Parameters:
//   - c *pkgapp.WebsocketClient: Current WebSocket client connection, including context and service factory (SF).
//
// 参数说明:
//   - c *pkgapp.WebsocketClient: 当前 WebSocket 客户端连接，包含上下文与服务工厂（SF）。
//   - uid int64: User ID to query.
//
// 参数说明:
//   - uid int64: 要查询的用户 ID。
//
// Return:
//   - *pkgapp.UserSelectEntity: If user found, returns converted user entity, otherwise nil.
//
// 返回值说明:
//   - *pkgapp.UserSelectEntity: 如果查询到用户则返回转换后的用户实体，否则返回 nil。
//   - error: Error during query (if any).
//
// 返回值说明:
//   - error: 查询过程中的错误（若有）。
func (h *NoteWSHandler) UserInfo(c *pkgapp.WebsocketClient, uid int64) (*pkgapp.UserSelectEntity, error) {

	// Use WebSocket connection's long-lived context
	// 使用 WebSocket 连接的长生命周期 context
	ctx := c.Context()
	user, err := h.App.UserService.GetInfo(ctx, uid)

	var userEntity *pkgapp.UserSelectEntity
	if user != nil {
		userEntity = convert.StructAssign(user, &pkgapp.UserSelectEntity{}).(*pkgapp.UserSelectEntity)
	}

	return userEntity, err

}
