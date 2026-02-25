package websocket_router

import (
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
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.setting.SettingModify.BindAndValid")
		return
	}

	ctx := c.Context()

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "SettingModify", params.Path, params.Vault)

	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	checkParams := convert.StructAssign(params, &dto.SettingUpdateCheckRequest{}).(*dto.SettingUpdateCheckRequest)
	updateMode, settingCheck, err := h.App.SettingService.UpdateCheck(ctx, c.User.UID, checkParams)
	if err != nil {
		h.respondError(c, code.ErrorSettingModifyOrCreateFailed, err, "websocket_router.setting.SettingModify.UpdateCheck")
		return
	}

	switch updateMode {
	case "UpdateContent", "Create":
		_, setting, err := h.App.SettingService.ModifyOrCreate(ctx, c.User.UID, params, true)
		if err != nil {
			h.respondError(c, code.ErrorSettingModifyOrCreateFailed, err, "websocket_router.setting.SettingModify.ModifyOrCreate")
			return
		}

		c.ToResponse(code.Success)
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
		).WithVault(params.Vault), true, dto.SettingSyncModify)
		return

	case "UpdateMtime":
		c.ToResponse(code.Success.WithData(
			dto.SettingSyncMtimeMessage{
				Path:  settingCheck.Path,
				Ctime: settingCheck.Ctime,
				Mtime: settingCheck.Mtime,
			},
		).WithVault(params.Vault), dto.SettingSyncMtime)
		return
	default:
		c.ToResponse(code.SuccessNoUpdate)
		return
	}
}

// SettingModifyCheck checks the necessity of setting modification
// SettingModifyCheck 检查配置修改必要性
func (h *SettingWSHandler) SettingModifyCheck(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.SettingUpdateCheckRequest{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.setting.SettingModifyCheck.BindAndValid")
		return
	}

	ctx := c.Context()

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "SettingModifyCheck", params.Path, params.Vault)

	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	updateMode, settingCheck, err := h.App.SettingService.UpdateCheck(ctx, c.User.UID, params)
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
		).WithVault(params.Vault), dto.SettingSyncNeedUpload)
		return
	case "UpdateMtime":
		c.ToResponse(code.Success.WithData(
			dto.SettingSyncMtimeMessage{
				Path:  settingCheck.Path,
				Ctime: settingCheck.Ctime,
				Mtime: settingCheck.Mtime,
			},
		).WithVault(params.Vault), dto.SettingSyncMtime)
		return
	default:
		c.ToResponse(code.SuccessNoUpdate)
		return
	}
}

// SettingDelete handles setting deletion messages
// SettingDelete 处理配置删除消息
func (h *SettingWSHandler) SettingDelete(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.SettingDeleteRequest{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.setting.SettingDelete.BindAndValid")
		return
	}

	ctx := c.Context()

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "SettingDelete", params.Path, params.Vault)

	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	setting, err := h.App.SettingService.Delete(ctx, c.User.UID, params)
	if err != nil {
		h.respondError(c, code.ErrorSettingDeleteFailed, err, "websocket_router.setting.SettingDelete.Delete")
		return
	}

	c.ToResponse(code.Success)
	c.BroadcastResponse(code.Success.WithData(
		dto.SettingSyncDeleteMessage{
			Path: setting.Path,
		},
	).WithVault(params.Vault), true, dto.SettingSyncDelete)
}

// SettingSync handles setting synchronization messages
// SettingSync 处理配置同步消息
func (h *SettingWSHandler) SettingSync(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.SettingSyncRequest{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.setting.SettingSync.BindAndValid")
		return
	}

	ctx := c.Context()

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "SettingSync", "", params.Vault)

	h.App.VaultService.GetOrCreate(ctx, c.User.UID, params.Vault)

	list, err := h.App.SettingService.Sync(ctx, c.User.UID, params)
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
	var messageQueue []dto.WSQueuedMessage

	var lastTime int64
	var needUploadCount int64
	var needModifyCount int64
	var needSyncMtimeCount int64
	var needDeleteCount int64

	var cDelSettingsKeys map[string]struct{} = make(map[string]struct{}, 0)

	// Handle settings deleted by client
	// 处理客户端删除的配置
	if len(params.DelSettings) > 0 {
		for _, delSetting := range params.DelSettings {

			// Check if setting exists before deleting
			getCheckParams := &dto.SettingGetRequest{
				Vault:    params.Vault,
				PathHash: delSetting.PathHash,
			}
			checkSetting, err := h.App.SettingService.Get(ctx, c.User.UID, getCheckParams)

			if err == nil && checkSetting != nil && checkSetting.Action != "delete" {
				delParams := &dto.SettingDeleteRequest{
					Vault:    params.Vault,
					Path:     delSetting.Path,
					PathHash: delSetting.PathHash,
				}
				setting, err := h.App.SettingService.Delete(ctx, c.User.UID, delParams)
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
				).WithVault(params.Vault), true, dto.SettingSyncDelete)
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
				).WithVault(params.Vault), true, dto.SettingSyncDelete)
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
			setting, err := h.App.SettingService.Get(ctx, c.User.UID, getParams)
			if err != nil {
				h.App.Logger().Warn("websocket_router.setting.SettingSync.SettingService.Get",
					zap.String(logger.FieldTraceID, c.TraceID),
					zap.String("pathHash", missingSetting.PathHash),
					zap.Error(err))
				continue
			}

			if setting != nil && setting.Action != "delete" {

				messageQueue = append(messageQueue, dto.WSQueuedMessage{
					Action: dto.SettingSyncModify,
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

		if s.UpdatedTimestamp >= lastTime {
			lastTime = s.UpdatedTimestamp
		}
		if s.Action == "delete" {

			if _, ok := cSettings[s.PathHash]; ok {
				delete(cSettingsKeys, s.PathHash)
				// 将消息添加到队列而非立即发送
				messageQueue = append(messageQueue, dto.WSQueuedMessage{
					Action: dto.SettingSyncDelete,
					Data:   dto.SettingSyncDeleteMessage{Path: s.Path},
				})
				needDeleteCount++
			}
		} else {
			if cSetting, ok := cSettings[s.PathHash]; ok {
				delete(cSettingsKeys, s.PathHash)
				if s.ContentHash == cSetting.ContentHash && s.Mtime == cSetting.Mtime {
					continue
				}
				// 强制覆盖连接端
				if params.Cover {
					// 将消息添加到队列而非立即发送
					messageQueue = append(messageQueue, dto.WSQueuedMessage{
						Action: dto.SettingSyncModify,
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
						messageQueue = append(messageQueue, dto.WSQueuedMessage{
							Action: dto.SettingSyncModify,
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
						messageQueue = append(messageQueue, dto.WSQueuedMessage{
							Action: dto.SettingSyncNeedUpload,
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
					messageQueue = append(messageQueue, dto.WSQueuedMessage{
						Action: dto.SettingSyncMtime,
						Data: dto.SettingSyncMtimeMessage{
							Path:  s.Path,
							Ctime: s.Ctime,
							Mtime: s.Mtime,
						},
					})
					needSyncMtimeCount++
				}
			} else {
				// 将消息添加到队列而非立即发送
				messageQueue = append(messageQueue, dto.WSQueuedMessage{
					Action: dto.SettingSyncModify,
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

	if list == nil {
		lastTime = timex.Now().UnixMilli()
	}
	for pathHash := range cSettingsKeys {
		s := cSettings[pathHash]
		// Add message to queue instead of sending immediately
		// 将消息添加到队列而非立即发送
		messageQueue = append(messageQueue, dto.WSQueuedMessage{
			Action: dto.SettingSyncNeedUpload,
			Data:   dto.SettingSyncNeedUploadMessage{Path: s.Path},
		})
		needUploadCount++
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
	).WithVault(params.Vault).WithContext(params.Context), dto.SettingSyncEnd)

	// Send queued messages individually
	// 逐条发送队列中的消息
	for _, item := range messageQueue {
		c.ToResponse(code.Success.WithData(item.Data).WithVault(params.Vault).WithContext(params.Context), item.Action)
	}
}

// SettingClear handles clear all settings messages
// SettingClear 处理清理所有配置消息
func (h *SettingWSHandler) SettingClear(c *pkgapp.WebsocketClient, msg *pkgapp.WebSocketMessage) {
	params := &dto.SettingClearRequest{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		h.respondErrorWithData(c, code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()), errs, errs.MapsToString(), "websocket_router.setting.SettingClear.BindAndValid")
		return
	}

	ctx := c.Context()

	pkgapp.NoteModifyLog(c.TraceID, c.User.UID, "SettingClear", "", params.Vault)

	err := h.App.SettingService.ClearByVault(ctx, c.User.UID, params.Vault)
	if err != nil {
		h.respondError(c, code.ErrorSettingDeleteFailed, err, "websocket_router.setting.SettingClear.ClearByVault")
		return
	}

	// Broadcast clearing to other clients with vault info
	// 将清除消息广播给其他客户端，带上笔记本信息
	c.BroadcastResponse(code.Success.WithData(nil).WithVault(params.Vault), false, dto.SettingSyncClear)
}
