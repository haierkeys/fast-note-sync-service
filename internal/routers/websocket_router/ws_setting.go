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
	Vault    string `json:"vault" form:"vault"`
	LastTime int64  `json:"lastTime" form:"lastTime"`
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
func SettingModify(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.SettingModifyOrCreateRequestParams{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("websocket_router.setting.SettingModify.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	checkParams := convert.StructAssign(params, &service.SettingUpdateCheckRequestParams{}).(*service.SettingUpdateCheckRequestParams)
	updateMode, settingCheck, err := svc.SettingUpdateCheck(c.User.UID, checkParams)
	if err != nil {
		c.ToResponse(code.ErrorSettingModifyOrCreateFailed.WithDetails(err.Error()))
		return
	}

	switch updateMode {
	case "UpdateContent", "Create":
		_, setting, err := svc.SettingModifyOrCreate(c.User.UID, params, true)
		if err != nil {
			c.ToResponse(code.ErrorSettingModifyOrCreateFailed.WithDetails(err.Error()))
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
		c.ToResponse(code.Success.Reset())
		c.BroadcastResponse(code.Success.Reset().WithData(settingMessage).WithVault(params.Vault), true, "SettingSyncModify")
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
		c.ToResponse(code.SuccessNoUpdate.Reset())
		return
	}
}

// SettingModifyCheck 检查配置修改必要性
func SettingModifyCheck(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.SettingUpdateCheckRequestParams{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("websocket_router.setting.SettingModifyCheck.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	updateMode, settingCheck, err := svc.SettingUpdateCheck(c.User.UID, params)
	if err != nil {
		c.ToResponse(code.ErrorSettingUpdateCheckFailed.WithDetails(err.Error()))
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
		c.ToResponse(code.SuccessNoUpdate.Reset())
		return
	}
}

// SettingDelete 处理配置删除消息
func SettingDelete(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.SettingDeleteRequestParams{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("websocket_router.setting.SettingDelete.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	setting, err := svc.SettingDelete(c.User.UID, params)
	if err != nil {
		c.ToResponse(code.ErrorSettingDeleteFailed.WithDetails(err.Error()))
		return
	}

	c.ToResponse(code.Success)
	c.BroadcastResponse(code.Success.WithData(setting).WithVault(params.Vault), true, "SettingSyncDelete")
}

// SettingSync 处理配置同步消息
func SettingSync(c *app.WebsocketClient, msg *app.WebSocketMessage) {
	params := &service.SettingSyncRequestParams{}
	valid, errs := c.BindAndValid(msg.Data, params)
	if !valid {
		global.Logger.Error("websocket_router.setting.SettingSync.BindAndValid errs: %v", zap.Error(errs))
		c.ToResponse(code.ErrorInvalidParams.WithDetails(errs.ErrorsToString()).WithData(errs.MapsToString()))
		return
	}

	svc := service.New(c.Ctx).WithSF(c.SF)
	svc.VaultGetOrCreate(params.Vault, c.User.UID)

	list, err := svc.SettingListByLastTime(c.User.UID, params)
	if err != nil {
		c.ToResponse(code.ErrorSettingListFailed.WithDetails(err.Error()))
		return
	}

	cSettings := make(map[string]service.SettingSyncCheckRequestParams)
	cSettingsKeys := make(map[string]struct{})
	for _, s := range params.Settings {
		cSettings[s.PathHash] = s
		cSettingsKeys[s.PathHash] = struct{}{}
	}

	var lastTime int64
	for _, s := range list {
		if s.UpdatedTimestamp >= lastTime {
			lastTime = s.UpdatedTimestamp
		}
		if s.Action == "delete" {

			if _, ok := cSettings[s.PathHash]; ok {
				delete(cSettingsKeys, s.PathHash)
				c.ToResponse(code.Success.WithData(&SettingDeleteMessage{Path: s.Path}), "SettingSyncDelete")
			}
		} else {
			if cSetting, ok := cSettings[s.PathHash]; ok {
				delete(cSettingsKeys, s.PathHash)
				if s.ContentHash == cSetting.ContentHash && s.Mtime == cSetting.Mtime {
					continue
				}
				// 强制覆盖连接端
				if params.Cover {
					c.ToResponse(code.Success.Reset().WithData(&SettingMessage{
						Path:             s.Path,
						PathHash:         s.PathHash,
						Content:          s.Content,
						ContentHash:      s.ContentHash,
						Ctime:            s.Ctime,
						Mtime:            s.Mtime,
						UpdatedTimestamp: s.UpdatedTimestamp,
					}), "SettingSyncModify")
					continue
				}
				// 链接端和服务端， 文件内容相同
				if s.ContentHash != cSetting.ContentHash {
					if s.Mtime >= cSetting.Mtime {
						// 服务端文件 mtime 大于链接端文件 mtime，则通知连接端更新
						c.ToResponse(code.Success.Reset().WithData(&SettingMessage{
							Path:             s.Path,
							PathHash:         s.PathHash,
							Content:          s.Content,
							ContentHash:      s.ContentHash,
							Ctime:            s.Ctime,
							Mtime:            s.Mtime,
							UpdatedTimestamp: s.UpdatedTimestamp,
						}), "SettingSyncModify")
					} else {
						// 服务端文件 mtime 小于链接端文件 mtime，则通知连接端更新
						c.ToResponse(code.Success.Reset().WithData(&SettingSyncNeedUploadMessage{
							Path: s.Path,
						}), "SettingSyncNeedUpload")
					}
				} else {
					// 链接端和服务端， 文件内容相同，文件 mtime 时间不同
					c.ToResponse(code.Success.WithData(&SettingSyncMtimeMessage{
						Path:  s.Path,
						Ctime: s.Ctime,
						Mtime: s.Mtime,
					}), "SettingSyncMtime")
				}
			} else {
				c.ToResponse(code.Success.WithData(&SettingMessage{
					Path:             s.Path,
					PathHash:         s.PathHash,
					Content:          s.Content,
					ContentHash:      s.ContentHash,
					Ctime:            s.Ctime,
					Mtime:            s.Mtime,
					UpdatedTimestamp: s.UpdatedTimestamp,
				}), "SettingSyncModify")
			}
		}
	}

	if list == nil {
		lastTime = timex.Now().UnixMilli()
	}
	for pathHash := range cSettingsKeys {
		s := cSettings[pathHash]
		c.ToResponse(code.Success.WithData(&SettingSyncNeedUploadMessage{Path: s.Path}), "SettingSyncNeedUpload")
	}

	c.ToResponse(code.Success.WithData(&SettingSyncEndMessage{
		Vault:    params.Vault,
		LastTime: lastTime,
	}), "SettingSyncEnd")
}
