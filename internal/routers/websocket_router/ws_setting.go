package websocket_router

import (
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
)

// SettingWSHandler WebSocket 配置处理器
// 使用 App Container 注入依赖
type SettingWSHandler struct {
	*WSHandler
}

// NewSettingWSHandler 创建 SettingWSHandler 实例
func NewSettingWSHandler(a *app.App) *SettingWSHandler {
	return &SettingWSHandler{
		WSHandler: NewWSHandler(a),
	}
}

type SettingMessage struct {
	Vault            string `json:"vault" form:"vault"`
	Path             string `json:"path" form:"path"`
	PathHash         string `json:"pathHash" form:"pathHash"`
	Content          string `json:"content" form:"content"`
	ContentHash      string `json:"contentHash" form:"contentHash"`
	Ctime            int64  `json:"ctime" form:"ctime"`
	Mtime            int64  `json:"mtime" form:"mtime"`
	UpdatedTimestamp int64  `json:"lastTime" form:"updatedTimestamp"`
}

type SettingSyncEndMessage struct {
	LastTime           int64           `json:"lastTime" form:"lastTime"`
	NeedUploadCount    int64           `json:"needUploadCount" form:"needUploadCount"`       // 需要上传的数量
	NeedModifyCount    int64           `json:"needModifyCount" form:"needModifyCount"`       // 需要修改的数量
	NeedSyncMtimeCount int64           `json:"needSyncMtimeCount" form:"needSyncMtimeCount"` // 需要同步修改时间的数量
	NeedDeleteCount    int64           `json:"needDeleteCount" form:"needDeleteCount"`       // 需要删除的数量
	Messages           []queuedMessage `json:"messages"`                                     // 合并的消息队列
}

type SettingSyncNeedUploadMessage struct {
	Path string `json:"path" form:"path"`
}

type SettingSyncMtimeMessage struct {
	Path  string `json:"path" form:"path"`
	Ctime int64  `json:"ctime" form:"ctime"`
	Mtime int64  `json:"mtime" form:"mtime"`
}

type SettingDeleteMessage struct {
	Path string `json:"path" form:"path"`
}

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

		settingMessage := &SettingMessage{
			Path:             setting.Path,
			PathHash:         setting.PathHash,
			Content:          setting.Content,
			ContentHash:      setting.ContentHash,
			Ctime:            setting.Ctime,
			Mtime:            setting.Mtime,
			UpdatedTimestamp: setting.UpdatedTimestamp,
		}
		c.ToResponse(code.Success)
		c.BroadcastResponse(code.Success.WithData(settingMessage).WithVault(params.Vault), true, "SettingSyncModify")
		return

	case "UpdateMtime":
		settingSyncMtimeMessage := &SettingSyncMtimeMessage{
			Path:  settingCheck.Path,
			Ctime: settingCheck.Ctime,
			Mtime: settingCheck.Mtime,
		}
		c.ToResponse(code.Success.WithData(settingSyncMtimeMessage), "SettingSyncMtime")
		return
	default:
		c.ToResponse(code.SuccessNoUpdate)
		return
	}
}

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
		settingSyncNeedPushMessage := &SettingSyncNeedUploadMessage{
			Path: settingCheck.Path,
		}
		c.ToResponse(code.Success.WithData(settingSyncNeedPushMessage), "SettingSyncNeedUpload")
		return
	case "UpdateMtime":
		settingSyncMtimeMessage := &SettingSyncMtimeMessage{
			Path:  settingCheck.Path,
			Ctime: settingCheck.Ctime,
			Mtime: settingCheck.Mtime,
		}
		c.ToResponse(code.Success.WithData(settingSyncMtimeMessage), "SettingSyncMtime")
		return
	default:
		c.ToResponse(code.SuccessNoUpdate)
		return
	}
}

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
	c.BroadcastResponse(code.Success.WithData(setting).WithVault(params.Vault), true, "SettingSyncDelete")
}

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

	// 创建消息队列，用于收集所有待发送的消息
	var messageQueue []queuedMessage

	var lastTime int64
	var needUploadCount int64
	var needModifyCount int64
	var needSyncMtimeCount int64
	var needDeleteCount int64

	for _, s := range list {
		if s.UpdatedTimestamp >= lastTime {
			lastTime = s.UpdatedTimestamp
		}
		if s.Action == "delete" {

			if _, ok := cSettings[s.PathHash]; ok {
				delete(cSettingsKeys, s.PathHash)
				// 将消息添加到队列而非立即发送
				messageQueue = append(messageQueue, queuedMessage{
					Action: "SettingSyncDelete",
					Data:   &SettingDeleteMessage{Path: s.Path},
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
					messageQueue = append(messageQueue, queuedMessage{
						Action: "SettingSyncModify",
						Data: &SettingMessage{
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
						// 服务端文件 mtime 大于链接端文件 mtime，则通知连接端更新
						// 将消息添加到队列而非立即发送
						messageQueue = append(messageQueue, queuedMessage{
							Action: "SettingSyncModify",
							Data: &SettingMessage{
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
						// 服务端文件 mtime 小于链接端文件 mtime，则通知连接端更新
						// 将消息添加到队列而非立即发送
						messageQueue = append(messageQueue, queuedMessage{
							Action: "SettingSyncNeedUpload",
							Data: &SettingSyncNeedUploadMessage{
								Path: s.Path,
							},
						})
						needUploadCount++
					}
				} else {
					// 链接端和服务端， 文件内容相同，文件 mtime 时间不同
					// 将消息添加到队列而非立即发送
					messageQueue = append(messageQueue, queuedMessage{
						Action: "SettingSyncMtime",
						Data: &SettingSyncMtimeMessage{
							Path:  s.Path,
							Ctime: s.Ctime,
							Mtime: s.Mtime,
						},
					})
					needSyncMtimeCount++
				}
			} else {
				// 将消息添加到队列而非立即发送
				messageQueue = append(messageQueue, queuedMessage{
					Action: "SettingSyncModify",
					Data: &SettingMessage{
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
		// 将消息添加到队列而非立即发送
		messageQueue = append(messageQueue, queuedMessage{
			Action: "SettingSyncNeedUpload",
			Data:   &SettingSyncNeedUploadMessage{Path: s.Path},
		})
		needUploadCount++
	}

	// 发送 SettingSyncEnd 消息，包含所有合并的消息
	message := &SettingSyncEndMessage{
		LastTime:           lastTime,
		NeedUploadCount:    needUploadCount,
		NeedModifyCount:    needModifyCount,
		NeedSyncMtimeCount: needSyncMtimeCount,
		NeedDeleteCount:    needDeleteCount,
		Messages:           messageQueue,
	}
	c.ToResponse(code.Success.WithData(message).WithVault(params.Vault), "SettingSyncEnd")
}
