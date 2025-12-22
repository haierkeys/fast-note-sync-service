package service

import (
	"fmt"
	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
)

// Setting 表示配置的完整数据结构。
type Setting struct {
	ID               int64      `json:"id" form:"id"`                     // 主键ID
	Action           string     `json:"-" form:"action"`                  // 操作：create/modify/delete
	Path             string     `json:"path" form:"path"`                 // 路径信息
	PathHash         string     `json:"pathHash" form:"pathHash"`         // 路径哈希值
	Content          string     `json:"content" form:"content"`           // 内容详情
	ContentHash      string     `json:"contentHash" form:"contentHash"`   // 内容哈希
	Ctime            int64      `json:"ctime" form:"ctime"`               // 创建时间戳
	Mtime            int64      `json:"mtime" form:"mtime"`               // 修改时间戳
	UpdatedTimestamp int64      `json:"lastTime" form:"updatedTimestamp"` // 记录更新时间戳
	UpdatedAt        timex.Time `json:"-"`                                // 更新时间
	CreatedAt        timex.Time `json:"-"`                                // 创建时间
}

// SettingUpdateCheckRequestParams 客户端检查更新请求参数。
type SettingUpdateCheckRequestParams struct {
	Vault       string `json:"vault" form:"vault"  binding:"required"`
	Path        string `json:"path" form:"path"  binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash"  binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash"`
	Ctime       int64  `json:"ctime" form:"ctime"  binding:"required"`
	Mtime       int64  `json:"mtime" form:"mtime"  binding:"required"`
}

// SettingModifyOrCreateRequestParams 修改或创建配置参数。
type SettingModifyOrCreateRequestParams struct {
	Vault       string `json:"vault" form:"vault"  binding:"required"`
	Path        string `json:"path" form:"path"  binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash"`
	Content     string `json:"content" form:"content"`
	ContentHash string `json:"contentHash" form:"contentHash"`
	Ctime       int64  `json:"ctime" form:"ctime"`
	Mtime       int64  `json:"mtime" form:"mtime"`
}

// SettingDeleteRequestParams 删除配置参数。
type SettingDeleteRequestParams struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash"`
}

// SettingSyncRequestParams 同步请求参数。
type SettingSyncRequestParams struct {
	Vault    string                          `json:"vault" form:"vault" binding:"required"`
	LastTime int64                           `json:"lastTime" form:"lastTime"`
	Cover    bool                            `json:"cover" form:"cover"`
	Settings []SettingSyncCheckRequestParams `json:"settings" form:"settings"`
}

// SettingSyncCheckRequestParams 单条同步检查参数。
type SettingSyncCheckRequestParams struct {
	Path        string `json:"path" form:"path"`
	PathHash    string `json:"pathHash" form:"pathHash"  binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash"`
	Mtime       int64  `json:"mtime" form:"mtime"  binding:"required"`
}

// SettingUpdateCheck 检查配置是否需要更新
func (svc *Service) SettingUpdateCheck(uid int64, params *SettingUpdateCheckRequestParams) (string, *Setting, error) {
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_Get_%d", uid), func() (any, error) {
		return svc.VaultIdGetByName(params.Vault, uid)
	})
	if err != nil {
		return "", nil, err
	}
	vaultID := vID.(int64)

	svc.SF.Do(fmt.Sprintf("Setting_%d", uid), func() (any, error) {
		return nil, svc.dao.SettingAutoMigrate(uid)
	})

	if setting, _ := svc.dao.SettingGetByPathHash(params.PathHash, vaultID, uid); setting != nil {
		settingSvc := convert.StructAssign(setting, &Setting{}).(*Setting)
		if setting.ContentHash == params.ContentHash {
			if params.Mtime < setting.Mtime {
				return "UpdateMtime", settingSvc, nil
			} else if params.Mtime > setting.Mtime {
				svc.dao.SettingUpdateMtime(params.Mtime, setting.ID, uid)
			}
			return "", settingSvc, nil
		} else {
			return "UpdateContent", settingSvc, nil
		}
	} else {
		return "Create", nil, nil
	}
}

// SettingModifyOrCreate 创建或修改配置
func (svc *Service) SettingModifyOrCreate(uid int64, params *SettingModifyOrCreateRequestParams, mtimeCheck bool) (bool, *Setting, error) {
	svc.SF.Do(fmt.Sprintf("Setting_%d", uid), func() (any, error) {
		return nil, svc.dao.SettingAutoMigrate(uid)
	})

	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_Get_%d", uid), func() (any, error) {
		return svc.VaultIdGetByName(params.Vault, uid)
	})
	if err != nil {
		return false, nil, err
	}
	vaultID := vID.(int64)

	settingSet := &dao.SettingSet{
		VaultID:     vaultID,
		Path:        params.Path,
		PathHash:    params.PathHash,
		Content:     params.Content,
		ContentHash: params.ContentHash,
		Size:        int64(len(params.Content)),
		Mtime:       params.Mtime,
		Ctime:       params.Ctime,
	}

	setting, _ := svc.dao.SettingGetByPathHash(params.PathHash, vaultID, uid)
	if setting != nil {
		if mtimeCheck && setting.Mtime == params.Mtime && setting.ContentHash == params.ContentHash {
			return false, nil, nil
		}
		if mtimeCheck && setting.Mtime < params.Mtime && setting.ContentHash == params.ContentHash {
			err := svc.dao.SettingUpdateMtime(params.Mtime, setting.ID, uid)
			if err != nil {
				return false, nil, err
			}
			setting.Mtime = params.Mtime
			return false, convert.StructAssign(setting, &Setting{}).(*Setting), nil
		}
		if setting.Action == "delete" {
			settingSet.Action = "create"
		} else {
			settingSet.Action = "modify"
		}
		res, err := svc.dao.SettingUpdate(settingSet, setting.ID, uid)
		if err != nil {
			return false, nil, err
		}
		return false, convert.StructAssign(res, &Setting{}).(*Setting), nil
	} else {
		settingSet.Action = "create"
		res, err := svc.dao.SettingCreate(settingSet, uid)
		if err != nil {
			return true, nil, err
		}
		return true, convert.StructAssign(res, &Setting{}).(*Setting), nil
	}
}

// SettingDelete 删除配置
func (svc *Service) SettingDelete(uid int64, params *SettingDeleteRequestParams) (*Setting, error) {
	svc.SF.Do(fmt.Sprintf("Setting_%d", uid), func() (any, error) {
		return nil, svc.dao.SettingAutoMigrate(uid)
	})

	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_Get_%d", uid), func() (any, error) {
		return svc.VaultIdGetByName(params.Vault, uid)
	})
	if err != nil {
		return nil, err
	}
	vaultID := vID.(int64)

	setting, err := svc.dao.SettingGetByPathHash(params.PathHash, vaultID, uid)
	if err != nil {
		return nil, err
	}
	settingSet := &dao.SettingSet{
		VaultID:     vaultID,
		Action:      "delete",
		Path:        setting.Path,
		PathHash:    setting.PathHash,
		Content:     "",
		ContentHash: "",
		Size:        0,
	}
	res, err := svc.dao.SettingUpdate(settingSet, setting.ID, uid)
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(res, &Setting{}).(*Setting), nil
}

// SettingListByLastTime 根据最后更新时间获取配置列表
func (svc *Service) SettingListByLastTime(uid int64, params *SettingSyncRequestParams) ([]*Setting, error) {
	svc.SF.Do(fmt.Sprintf("Setting_%d", uid), func() (any, error) {
		return nil, svc.dao.SettingAutoMigrate(uid)
	})

	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_Get_%d", uid), func() (any, error) {
		return svc.VaultIdGetByName(params.Vault, uid)
	})
	if err != nil {
		return nil, err
	}
	vaultID := vID.(int64)

	settings, err := svc.dao.SettingListByUpdatedTimestamp(params.LastTime, vaultID, uid)
	if err != nil {
		return nil, err
	}

	var results []*Setting
	var cacheList map[string]bool = make(map[string]bool)
	for _, s := range settings {
		if cacheList[s.PathHash] {
			continue
		}
		results = append(results, convert.StructAssign(s, &Setting{}).(*Setting))
		cacheList[s.PathHash] = true
	}
	return results, nil
}

// SettingCleanup 清理过期的软删除配置
func (svc *Service) SettingCleanup(uid int64) error {
	retentionTimeStr := global.Config.App.SoftDeleteRetentionTime
	if retentionTimeStr == "" || retentionTimeStr == "0" {
		return nil
	}

	retentionDuration, err := util.ParseDuration(retentionTimeStr)
	if err != nil {
		return err
	}

	if retentionDuration <= 0 {
		return nil
	}

	cutoffTime := time.Now().Add(-retentionDuration).UnixMilli()
	return svc.dao.SettingDeletePhysicalByTime(cutoffTime, uid)
}

// SettingCleanupAll 清理所有用户的过期软删除配置
func (svc *Service) SettingCleanupAll() error {
	uids, err := svc.dao.GetAllUserUIDs()
	if err != nil {
		return err
	}

	for _, uid := range uids {
		_ = svc.SettingCleanup(uid)
	}

	return nil
}
