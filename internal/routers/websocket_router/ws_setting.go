package websocket_router

import (
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"go.uber.org/zap"
)

// SettingWSHandler WebSocket setting handler
// SettingWSHandler WebSocket 配置处理器
// Uses App Container to inject dependencies
// 使用 App Container 注入依赖
type SettingWSHandler struct {
	*WSHandler
}

// NewSettingWSHandler creates SettingWSHandler instance
// NewSettingWSHandler 创建 SettingWSHandler 实例
func NewSettingWSHandler(a *app.App) *SettingWSHandler {
	return &SettingWSHandler{
		WSHandler: NewWSHandler(a),
	}
}

// SettingModify handles setting modification messages
// SettingModify 处理配置修改消息
func (h *SettingWSHandler) SettingModify(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.SettingModifyOrCreateRequest{}
	valid, errs := c.BindAndValidWithAction(msg.Type, msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.setting.SettingModify.BindAndValid")
		return
	}

	ctx := c.Context()

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "SettingModify", params.Path, params.Vault)

	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	settingSvc := h.App.GetSettingService(c.ClientType, c.ClientName, c.ClientVersion)

	checkParams := convert.StructAssign(params, &dto.SettingUpdateCheckRequest{}).(*dto.SettingUpdateCheckRequest)
	updateMode, settingCheck, err := settingSvc.UpdateCheck(ctx, c.User.UID, checkParams)
	if err != nil {
		h.respondError(c, code.ErrorSettingModifyOrCreateFailed, err, "websocket_router.setting.SettingModify.UpdateCheck")
		return
	}

	switch updateMode {
	case "UpdateContent", "Create":
		_, setting, err := settingSvc.ModifyOrCreate(ctx, c.User.UID, params, true)
		if err != nil {
			h.respondError(c, code.ErrorSettingModifyOrCreateFailed, err, "websocket_router.setting.SettingModify.ModifyOrCreate")
			return
		}

		c.ToResponse(code.Success.WithData(dto.SettingModifyAckMessage{
			LastTime: setting.UpdatedTimestamp,
			Path:     setting.Path,
			PathHash: setting.PathHash,
		}).WithVault(params.Vault), string(SettingModifyAck))
		c.BroadcastResponse(code.Success.WithData(
			dto.SettingSyncModifyMessage{
				Vault:            params.Vault,
				Path:             setting.Path,
				PathHash:         setting.PathHash,
				Content:          setting.Content,
				ContentHash:      setting.ContentHash,
				Ctime:            setting.Ctime,
				Mtime:            setting.Mtime,
				UpdatedTimestamp: setting.UpdatedTimestamp,
			},
		).WithVault(params.Vault), true, SettingSyncModify)
		return

	case "UpdateMtime":
		c.ToResponse(code.Success.WithData(dto.SettingModifyAckMessage{
			LastTime: settingCheck.UpdatedTimestamp,
			Path:     settingCheck.Path,
			PathHash: settingCheck.PathHash,
		}).WithVault(params.Vault), string(SettingModifyAck))
		return
	default:
		c.ToResponse(code.SuccessNoUpdate.WithData(dto.SettingModifyAckMessage{
			LastTime: params.Mtime,
			Path:     params.Path,
			PathHash: params.PathHash,
		}).WithVault(params.Vault), string(SettingModifyAck))
		return
	}
}

// SettingModifyCheck checks the necessity of setting modification
// SettingModifyCheck 检查配置修改必要性
func (h *SettingWSHandler) SettingModifyCheck(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.SettingUpdateCheckRequest{}
	valid, errs := c.BindAndValidWithAction(msg.Type, msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.setting.SettingModifyCheck.BindAndValid")
		return
	}

	ctx := c.Context()

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "SettingModifyCheck", params.Path, params.Vault)

	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	settingSvc := h.App.GetSettingService(c.ClientType, c.ClientName, c.ClientVersion)

	updateMode, settingCheck, err := settingSvc.UpdateCheck(ctx, c.User.UID, params)
	if err != nil {
		h.respondError(c, code.ErrorSettingUpdateCheckFailed, err, "websocket_router.setting.SettingModifyCheck.UpdateCheck")
		return
	}

	switch updateMode {
	case "UpdateContent", "Create":
		c.ToResponse(code.Success.WithData(
			dto.SettingSyncNeedUploadMessage{
				Path: settingCheck.Path,
			},
		), SettingSyncNeedUpload)
		return
	case "UpdateMtime":
		c.ToResponse(code.Success.WithData(
			dto.SettingSyncMtimeMessage{
				Path:             settingCheck.Path,
				Ctime:            settingCheck.Ctime,
				Mtime:            settingCheck.Mtime,
				UpdatedTimestamp: settingCheck.UpdatedTimestamp,
			},
		), SettingSyncMtime)
		return
	default:
		c.ToResponse(code.SuccessNoUpdate.WithVault(params.Vault))
		return
	}
}

// SettingDelete handles setting deletion messages
// SettingDelete 处理配置删除消息
func (h *SettingWSHandler) SettingDelete(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.SettingDeleteRequest{}
	valid, errs := c.BindAndValidWithAction(msg.Type, msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.setting.SettingDelete.BindAndValid")
		return
	}

	ctx := c.Context()

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "SettingDelete", params.Path, params.Vault)

	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	settingSvc := h.App.GetSettingService(c.ClientType, c.ClientName, c.ClientVersion)

	setting, err := settingSvc.Delete(ctx, c.User.UID, params)
	if err != nil {
		h.respondError(c, code.ErrorSettingDeleteFailed, err, "websocket_router.setting.SettingDelete.Delete")
		return
	}

	c.ToResponse(code.Success.WithData(dto.SettingDeleteAckMessage{
		LastTime: setting.UpdatedTimestamp,
		Path:     setting.Path,
		PathHash: setting.PathHash,
	}).WithVault(params.Vault), string(SettingDeleteAck))
	c.BroadcastResponse(code.Success.WithData(
		dto.SettingSyncDeleteMessage{
			Path:             setting.Path,
			PathHash:         setting.PathHash,
			Ctime:            setting.Ctime,
			Mtime:            setting.Mtime,
			UpdatedTimestamp: setting.UpdatedTimestamp,
		},
	).WithVault(params.Vault), true, SettingSyncDelete)
}

// SettingSync handles setting synchronization messages
// SettingSync 处理配置同步消息
func (h *SettingWSHandler) SettingSync(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.SettingSyncRequest{}
	valid, errs := c.BindAndValidWithAction(msg.Type, msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.setting.SettingSync.BindAndValid")
		return
	}

	// 分批协议快速路径：totalBatches <= 1 时直接执行，无需缓存归集
	// Fast path: totalBatches <= 1 means single batch, skip cache accumulation
	if params.TotalBatches > 1 {
		entry := syncBatchGetOrCreate(params.Context, "setting", params.TotalBatches)

		entry.mu.Lock()
		for _, s := range params.Settings {
			entry.Items = append(entry.Items, s)
		}
		entry.ReceivedCount++
		// 元数据每批次覆盖写（最后一批的值覆盖前批），保证重传时的幂等性
		// Overwrite meta every batch; last batch's value wins for idempotence on retransmit
		entry.DelItems = params.DelSettings
		entry.MissingItems = params.MissingSettings
		entry.UpdatedAt = time.Now()
		received := entry.ReceivedCount
		total := entry.TotalBatches
		entry.mu.Unlock()

		if received < total {
			// 未集齐：回送 BatchAck，通知客户端可以发送下一批
			// Not all batches received: send BatchAck to signal client to send next batch
			c.ToResponse(code.Success.WithData(map[string]interface{}{
				"context":    params.Context,
				"batchIndex": params.BatchIndex,
			}).WithVault(params.Vault), SettingSyncBatchAck)
			return
		}

		// 全部批次到达：从缓存中提取数据，清理缓存，执行差量比对
		// All batches received: extract from cache, delete cache, run differential sync
		syncBatchDelete(params.Context, "setting")
		allSettings := make([]dto.SettingSyncCheckRequest, 0, len(entry.Items))
		for _, item := range entry.Items {
			allSettings = append(allSettings, item.(dto.SettingSyncCheckRequest))
		}
		params.Settings = allSettings
		params.DelSettings, _ = entry.DelItems.([]dto.SettingSyncDelSetting)
		params.MissingSettings, _ = entry.MissingItems.([]dto.SettingSyncDelSetting)
	}

	h.doSettingSync(c, params)
}

func (h *SettingWSHandler) doSettingSync(c *pkgapp.WebsocketClient, params *dto.SettingSyncRequest) {
	ctx := c.Context()

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "SettingSync", "", params.Vault)

	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	settingSvc := h.App.GetSettingService(c.ClientType, c.ClientName, c.ClientVersion)

	// Record sync start time before querying to avoid missing writes that occur during query processing.
	// 查询前记录同步开始时间，防止查询处理期间的写入被遗漏（经典增量同步快照时间戳方案）。
	syncStartTime := timex.Now().UnixMilli()

	list, err := settingSvc.Sync(ctx, c.User.UID, params)
	if err != nil {
		h.respondError(c, code.ErrorSettingListFailed, err, "websocket_router.setting.SettingSync.Sync")
		return
	}

	cSettings := make(map[string]dto.SettingSyncCheckRequest)
	cSettingsKeys := make(map[string]struct{})
	for _, s := range params.Settings {
		cSettings[s.PathHash] = s
		cSettingsKeys[s.PathHash] = struct{}{}
	}

	// Create message queue for collecting all messages to be sent
	// 创建消息队列，用于收集所有待发送的消息
	// Check and create vault, internally uses SF to merge concurrent requests, avoiding duplicate creation issues
	// 检查并创建仓库，内部使用SF合并并发请求, 避免重复创建问题
	var messageQueue []WSQueuedMessage

	var lastTime int64
	var needUploadCount int64
	var needModifyCount int64
	var needSyncMtimeCount int64
	var needDeleteCount int64

	var cDelSettingsKeys map[string]struct{} = make(map[string]struct{}, 0)

	// Handle settings deleted by client
	// 处理客户端删除的配置
	if len(params.DelSettings) > 0 {
		hasWritePermission := pkgapp.VerifyPermissions(c.Scope, "ws", c.ClientType, "config_w")

		for _, delSetting := range params.DelSettings {

			// Check if setting exists before deleting
			getCheckParams := &dto.SettingGetRequest{
				Vault:    params.Vault,
				PathHash: delSetting.PathHash,
			}
			checkSetting, err := settingSvc.Get(ctx, c.User.UID, getCheckParams)

			if err == nil && checkSetting != nil && checkSetting.Action != "delete" {
				if !hasWritePermission {
					h.App.Logger().Warn("websocket_router.setting.SettingSync: permission denied for deletion",
						zap.String(logger.FieldTraceID, c.TraceID),
						zap.Int64(logger.FieldUID, c.User.UID),
						zap.String(logger.FieldPath, delSetting.Path))
					continue
				}

				delParams := &dto.SettingDeleteRequest{
					Vault:    params.Vault,
					Path:     delSetting.Path,
					PathHash: delSetting.PathHash,
				}
				setting, err := settingSvc.Delete(ctx, c.User.UID, delParams)
				if err != nil {
					h.App.Logger().Error("websocket_router.setting.SettingSync.SettingService.Delete",
						zap.String(logger.FieldTraceID, c.TraceID),
						zap.Int64(logger.FieldUID, c.User.UID),
						zap.String(logger.FieldPath, delSetting.Path),
						zap.Error(err))
					continue
				}

				// 记录客户端已主动删除的 PathHash,避免重复下发
				cDelSettingsKeys[delSetting.PathHash] = struct{}{}
				// Broadcast deletion to other clients
				// 将删除消息广播给其他客户端
				c.BroadcastResponse(code.Success.WithData(
					dto.SettingSyncDeleteMessage{
						Path: setting.Path,
					},
				).WithVault(params.Vault), true, SettingSyncDelete)
			} else {
				h.App.Logger().Debug("websocket_router.setting.SettingSync.SettingService.Get check failed (not found or already deleted), broadcasting delete anyway",
					zap.String(logger.FieldTraceID, c.TraceID),
					zap.String("pathHash", delSetting.PathHash))

				// Record PathHash
				// 记录 PathHash
				cDelSettingsKeys[delSetting.PathHash] = struct{}{}

				// Broadcast deletion with available info (Path)
				// 使用现有信息(Path)广播删除
				c.BroadcastResponse(code.Success.WithData(
					dto.SettingSyncDeleteMessage{
						Path: delSetting.Path,
					},
				).WithVault(params.Vault), true, SettingSyncDelete)
			}

		}
	}

	// Handle settings missing on client (only for incremental sync)
	// 处理客户端缺失的配置（仅限增量同步）
	if params.LastTime > 0 && len(params.MissingSettings) > 0 {
		for _, missingSetting := range params.MissingSettings {
			getParams := &dto.SettingGetRequest{
				Vault:    params.Vault,
				PathHash: missingSetting.PathHash,
			}
			setting, err := settingSvc.Get(ctx, c.User.UID, getParams)
			if err != nil {
				h.App.Logger().Warn("websocket_router.setting.SettingSync.SettingService.Get",
					zap.String(logger.FieldTraceID, c.TraceID),
					zap.String("pathHash", missingSetting.PathHash),
					zap.Error(err))
				continue
			}

			if setting != nil && setting.Action != "delete" {

				messageQueue = append(messageQueue, WSQueuedMessage{
					Action: SettingSyncModify,
					Data: dto.SettingSyncModifyMessage{
						Vault:            params.Vault,
						Path:             setting.Path,
						PathHash:         setting.PathHash,
						Content:          setting.Content,
						ContentHash:      setting.ContentHash,
						Ctime:            setting.Ctime,
						Mtime:            setting.Mtime,
						UpdatedTimestamp: setting.UpdatedTimestamp,
					},
				})
				needModifyCount++
				// 加入排除索引
				cDelSettingsKeys[setting.PathHash] = struct{}{}
			}
		}
	}

	for _, s := range list {
		// 如果该配置是客户端刚才通过参数告知删除的,则跳过下发
		if _, ok := cDelSettingsKeys[s.PathHash]; ok {
			continue
		}

		if s.Action == "delete" {
			// Server already deleted, notify client to delete (regardless of whether client has it)
			// 服务端已经删除，通知客户端删除（不再检查客户端是否存在）
			if _, ok := cSettings[s.PathHash]; ok {
				delete(cSettingsKeys, s.PathHash)
			}
			// 将消息添加到队列
			messageQueue = append(messageQueue, WSQueuedMessage{
				Action: SettingSyncDelete,
				Data: dto.SettingSyncDeleteMessage{
					Path:             s.Path,
					PathHash:         s.PathHash,
					Ctime:            s.Ctime,
					Mtime:            s.Mtime,
					UpdatedTimestamp: s.UpdatedTimestamp,
				},
			})
			needDeleteCount++
		} else {
			if cSetting, ok := cSettings[s.PathHash]; ok {
				delete(cSettingsKeys, s.PathHash)
				if s.ContentHash == cSetting.ContentHash && s.Mtime == cSetting.Mtime {
					continue
				}
				// 强制覆盖连接端
				if params.Cover {
					// 将消息添加到队列而非立即发送
					messageQueue = append(messageQueue, WSQueuedMessage{
						Action: SettingSyncModify,
						Data: dto.SettingSyncModifyMessage{
							Vault:            params.Vault,
							Path:             s.Path,
							PathHash:         s.PathHash,
							Content:          s.Content,
							ContentHash:      s.ContentHash,
							Ctime:            s.Ctime,
							Mtime:            s.Mtime,
							UpdatedTimestamp: s.UpdatedTimestamp,
						},
					})
					needModifyCount++
					continue
				}
				// 链接端和服务端， 文件内容相同
				if s.ContentHash != cSetting.ContentHash {
					if s.Mtime >= cSetting.Mtime {
						// Server file mtime is greater than client file mtime, notify client to update
						// 将消息添加到队列而非立即发送
						// 服务端文件 mtime 大于链接端文件 mtime，则通知连接端更新
						// 将消息添加到队列而非立即发送
						messageQueue = append(messageQueue, WSQueuedMessage{
							Action: SettingSyncModify,
							Data: dto.SettingSyncModifyMessage{
								Vault:            params.Vault,
								Path:             s.Path,
								PathHash:         s.PathHash,
								Content:          s.Content,
								ContentHash:      s.ContentHash,
								Ctime:            s.Ctime,
								Mtime:            s.Mtime,
								UpdatedTimestamp: s.UpdatedTimestamp,
							},
						})
						needModifyCount++
					} else {
						// Server file mtime is less than client file mtime, notify client to update
						// 将消息添加到队列而非立即发送
						// 服务端文件 mtime 小于链接端文件 mtime，则通知连接端更新
						// 将消息添加到队列而非立即发送
						messageQueue = append(messageQueue, WSQueuedMessage{
							Action: SettingSyncNeedUpload,
							Data: dto.SettingSyncNeedUploadMessage{
								Path: s.Path,
							},
						})
						needUploadCount++
					}
				} else {
					// Client and server have same content, but different mtime
					// 将消息添加到队列而非立即发送
					// 链接端和服务端， 文件内容相同，文件 mtime 时间不同
					// 将消息添加到队列而非立即发送
					messageQueue = append(messageQueue, WSQueuedMessage{
						Action: SettingSyncMtime,
						Data: dto.SettingSyncMtimeMessage{
							Path:             s.Path,
							Ctime:            s.Ctime,
							Mtime:            s.Mtime,
							UpdatedTimestamp: s.UpdatedTimestamp,
						},
					})
					needSyncMtimeCount++
				}
			} else {
				// 将消息添加到队列而非立即发送
				messageQueue = append(messageQueue, WSQueuedMessage{
					Action: SettingSyncModify,
					Data: dto.SettingSyncModifyMessage{
						Vault:            params.Vault,
						Path:             s.Path,
						PathHash:         s.PathHash,
						Content:          s.Content,
						ContentHash:      s.ContentHash,
						Ctime:            s.Ctime,
						Mtime:            s.Mtime,
						UpdatedTimestamp: s.UpdatedTimestamp,
					},
				})
				needModifyCount++
			}
		}
	}

	// Use syncStartTime (recorded before query) as lastTime to prevent writes that occurred
	// during query processing from being permanently missed on the next incremental sync.
	// 使用查询前记录的 syncStartTime 作为 lastTime，防止查询处理期间的写入在下次增量同步时被永久遗漏。
	lastTime = syncStartTime
	hasWritePermission := pkgapp.VerifyPermissions(c.Scope, "ws", c.ClientType, "config_w")
	for pathHash := range cSettingsKeys {
		s := cSettings[pathHash]
		// Add message to queue instead of sending immediately
		// 将消息添加到队列而非立即发送
		if hasWritePermission {
			messageQueue = append(messageQueue, WSQueuedMessage{
				Action: SettingSyncNeedUpload,
				Data:   dto.SettingSyncNeedUploadMessage{Path: s.Path},
			})
			needUploadCount++
		} else {
			h.App.Logger().Warn("websocket_router.setting.SettingSync: permission denied for upload",
				zap.String(logger.FieldTraceID, c.TraceID),
				zap.Int64(logger.FieldUID, c.User.UID),
				zap.String(logger.FieldPath, s.Path))
		}
	}

	// Send SettingSyncEnd message
	// 发送 SettingSyncEnd 消息
	c.ToResponse(code.Success.WithData(
		dto.SettingSyncEndMessage{
			LastTime:           lastTime,
			NeedUploadCount:    needUploadCount,
			NeedModifyCount:    needModifyCount,
			NeedSyncMtimeCount: needSyncMtimeCount,
			NeedDeleteCount:    needDeleteCount,
		},
	).WithVault(params.Vault).WithContext(params.Context), SettingSyncEnd)

	// 在 End 消息后，启动受控分页发送流程
	if len(messageQueue) > 0 {
		pageSize := h.App.Config().App.SyncDownloadChunkNum
		if pageSize <= 0 {
			pageSize = 50 // 默认值防呆
		}
		entry := &syncDownloadEntry{
			Context:      params.Context,
			TypeName:     "setting",
			Vault:        params.Vault,
			MessageQueue: messageQueue,
			PageSize:     pageSize,
			CurrentPage:  0,
		}
		syncDownloadStore(params.Context, "setting", entry)
		sendSyncPage(c, entry)
	}
}

// SettingSyncPageAck handles WebSocket messages for client page ACK
// SettingSyncPageAck 处理客户端发来的分页下载 ACK 消息
func (h *SettingWSHandler) SettingSyncPageAck(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.SyncPageAckRequest{}
	valid, errs := c.BindAndValidWithAction(msg.Type, msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.setting.SettingSyncPageAck.BindAndValid")
		return
	}

	entry, ok := syncDownloadGet(params.Context, "setting")
	if !ok {
		h.App.Logger().Warn("SettingSyncPageAck: sync download entry not found",
			zap.String(logger.FieldTraceID, c.TraceID),
			zap.String("context", params.Context))
		return
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	if params.PageIndex != entry.CurrentPage {
		h.App.Logger().Warn("SettingSyncPageAck: page index mismatch",
			zap.String(logger.FieldTraceID, c.TraceID),
			zap.Int("expected", entry.CurrentPage),
			zap.Int("got", params.PageIndex))
		return
	}

	start := entry.CurrentPage * entry.PageSize
	end := start + entry.PageSize
	if end >= len(entry.MessageQueue) {
		syncDownloadDelete(params.Context, "setting")
		h.App.Logger().Info("SettingSyncPageAck: sync finished, cache cleared",
			zap.String(logger.FieldTraceID, c.TraceID),
			zap.String("context", params.Context))
		return
	}

	entry.CurrentPage++
	entry.UpdatedAt = time.Now()
	sendSyncPage(c, entry)
}


// SettingClear handles clear all settings messages
// SettingClear 处理清理所有配置消息
func (h *SettingWSHandler) SettingClear(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.SettingClearRequest{}
	valid, errs := c.BindAndValidWithAction(msg.Type, msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.setting.SettingClear.BindAndValid")
		return
	}

	ctx := c.Context()

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "SettingClear", "", params.Vault)

	err := h.App.GetSettingService(c.ClientType, c.ClientName, c.ClientVersion).ClearByVault(ctx, c.User.UID, params.Vault)
	if err != nil {
		h.respondError(c, code.ErrorSettingDeleteFailed, err, "websocket_router.setting.SettingClear.ClearByVault")
		return
	}

	// Broadcast clearing to other clients with vault info
	// 将清除消息广播给其他客户端，带上笔记本信息
	c.BroadcastResponse(code.Success.WithData(nil).WithVault(params.Vault), false, SettingSyncClear)
}

// SettingRePush handles setting missing pull request
// SettingRePush 处理配置缺失请求拉取
func (h *SettingWSHandler) SettingRePush(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.SettingGetRequest{}
	valid, errs := c.BindAndValidWithAction(msg.Type, msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.setting.SettingRePush.BindAndValid")
		return
	}

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "SettingRePush", params.Path, params.Vault)

	ctx := c.Context()
	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	setting, err := h.App.GetSettingService(c.ClientType, c.ClientName, c.ClientVersion).Get(ctx, c.User.UID, params)
	if err != nil {
		h.App.Logger().Debug("websocket_router.setting.SettingRePush.Get: record not found or error, proceeding to send delete",
			zap.String(logger.FieldTraceID, c.TraceID),
			zap.Error(err))
	}

	if setting != nil && setting.Action != "delete" {
		c.ToResponse(code.Success.WithData(
			dto.SettingSyncModifyMessage{
				Vault:            params.Vault,
				Path:             setting.Path,
				PathHash:         setting.PathHash,
				Content:          setting.Content,
				ContentHash:      setting.ContentHash,
				Ctime:            setting.Ctime,
				Mtime:            setting.Mtime,
				UpdatedTimestamp: setting.UpdatedTimestamp,
			},
		).WithVault(params.Vault), SettingSyncModify)
	} else {
		// If setting not found, send delete message to client to clean up local unauthorized creation
		// 如果未找到配置，则向客户端发送删除消息，以清理本地未授权的创建
		c.ToResponse(code.Success.WithData(
			dto.SettingSyncDeleteMessage{
				Path: params.Path,
			},
		).WithVault(params.Vault), string(SettingSyncDelete))
	}
}
