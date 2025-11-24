package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
)

var (
	lastCleanupTime time.Time
	cleanupMutex    sync.Mutex
)

// Note 表示笔记的完整数据结构（包含内容）。
type Note struct {
	ID               int64      `json:"id" form:"id"`                             // 主键ID
	Action           string     `json:"action" form:"action"`                     // 操作：create/modify/delete
	Path             string     `json:"path" form:"path"`                         // 路径信息（文件路径）
	PathHash         string     `json:"pathHash" form:"pathHash"`                 // 路径哈希值，用于快速查找
	Content          string     `json:"content" form:"content"`                   // 内容详情（完整文本）
	ContentHash      string     `json:"contentHash" form:"contentHash"`           // 内容哈希，用于判定内容是否变更
	Ctime            int64      `json:"ctime" form:"ctime"`                       // 创建时间戳（秒）
	Mtime            int64      `json:"mtime" form:"mtime"`                       // 文件修改时间戳（秒）
	UpdatedTimestamp int64      `json:"updatedTimestamp" form:"updatedTimestamp"` // 记录更新时间戳（用于同步）
	UpdatedAt        timex.Time `json:"updatedAt"`                                // 更新时间字段（time 类型包装）
	CreatedAt        timex.Time `json:"createdAt"`                                // 创建时间字段（time 类型包装）
}

// NoteNoContent 表示不包含 content 字段的笔记数据结构，适合列表查询返回简洁信息。
type NoteNoContent struct {
	ID               int64      `json:"id" form:"id"`                             // 主键ID
	Action           string     `json:"action" form:"action"`                     // 操作：create/modify/delete
	Path             string     `json:"path" form:"path"`                         // 路径信息（文件路径）
	PathHash         string     `json:"pathHash" form:"pathHash"`                 // 路径哈希值，用于快速查找
	Ctime            int64      `json:"ctime" form:"ctime"`                       // 创建时间戳（秒）
	Mtime            int64      `json:"mtime" form:"mtime"`                       // 文件修改时间戳（秒）
	UpdatedTimestamp int64      `json:"updatedTimestamp" form:"updatedTimestamp"` // 记录更新时间戳（用于同步）
	UpdatedAt        timex.Time `json:"updatedAt"`                                // 更新时间字段（time 类型包装）
	CreatedAt        timex.Time `json:"createdAt"`                                // 创建时间字段（time 类型包装）
}

// NoteUpdateCheckRequestParams 客户端用于检查是否需要更新的请求参数。
type NoteUpdateCheckRequestParams struct {
	Vault       string `json:"vault" form:"vault"  binding:"required"`       // 仓库标识
	Path        string `json:"path" form:"path"  binding:"required"`         // 路径
	PathHash    string `json:"pathHash" form:"pathHash"  binding:"required"` // 路径哈希
	ContentHash string `json:"contentHash" form:"contentHash"  binding:""`   // 内容哈希（可选）
	Ctime       int64  `json:"ctime" form:"ctime"  binding:"required"`       // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime"  binding:"required"`       // 修改时间戳
}

// NoteModifyOrCreateRequestParams 用于创建或修改笔记的请求参数。
type NoteModifyOrCreateRequestParams struct {
	Vault       string `json:"vault" form:"vault"  binding:"required"`     // 仓库标识
	Path        string `json:"path" form:"path"  binding:"required"`       // 路径
	PathHash    string `json:"pathHash" form:"pathHash"`                   // 路径哈希
	SrcPath     string `json:"srcPath" form:"srcPath" `                    // 源路径用于删除笔记
	SrcPathHash string `json:"srcPathHash" form:"srcPathHash"`             // 源路径哈希用于删除笔记
	Content     string `json:"content" form:"content"  binding:""`         // 内容详情（可选）
	ContentHash string `json:"contentHash" form:"contentHash"  binding:""` // 内容哈希（可选）
	Ctime       int64  `json:"ctime" form:"ctime" `                        // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime" `                        // 修改时间戳
}

// ContentModifyRequestParams 专用于只修改内容的请求参数（必填 content）。
type ContentModifyRequestParams struct {
	Vault       string `json:"vault" form:"vault"  binding:"required"`             // 仓库标识
	Path        string `json:"path" form:"path"  binding:"required"`               // 路径
	PathHash    string `json:"pathHash" form:"pathHash"  binding:"required"`       // 路径哈希
	Content     string `json:"content" form:"content"  binding:"required"`         // 新内容（必填）
	ContentHash string `json:"contentHash" form:"contentHash"  binding:"required"` // 内容哈希（必填）
	Ctime       int64  `json:"ctime" form:"ctime"  binding:"required"`             // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime"  binding:"required"`             // 修改时间戳
}

// NoteDeleteRequestParams 删除笔记所需参数。
type NoteDeleteRequestParams struct {
	Vault    string `json:"vault" form:"vault" binding:"required"` // 仓库标识
	Path     string `json:"path" form:"path" binding:"required"`   // 路径
	PathHash string `json:"pathHash" form:"pathHash" `             // 路径哈希（可选）
}

// NoteSyncCheckRequestParams 同步检查单条记录的参数。
type NoteSyncCheckRequestParams struct {
	Path        string `json:"path" form:"path"`                             // 路径
	PathHash    string `json:"pathHash" form:"pathHash"  binding:"required"` // 路径哈希
	ContentHash string `json:"contentHash" form:"contentHash"  binding:""`   // 内容哈希（可选）
	Mtime       int64  `json:"mtime" form:"mtime"  binding:"required"`       // 修改时间戳
}

// NoteSyncRequestParams 同步请求主体，包含 vault、上次同步时间和要检查的笔记列表。
type NoteSyncRequestParams struct {
	Vault    string                       `json:"vault" form:"vault" binding:"required"` // 仓库标识
	LastTime int64                        `json:"lastTime" form:"lastTime"`              // 上次同步时间戳
	Notes    []NoteSyncCheckRequestParams `json:"notes" form:"notes"`                    // 客户端笔记信息列表
}

// NoteSyncEndMessage 同步结束时返回的信息结构。
type NoteSyncEndMessage struct {
	Vault    string `json:"vault" form:"vault"`       // 仓库标识
	LastTime int64  `json:"lastTime" form:"lastTime"` // 本次同步更新时间
}

// NoteSyncNeedPushMessage 服务端告知客户端需要推送的文件信息。
type NoteSyncNeedPushMessage struct {
	Path  string `json:"path" form:"path"`   // 路径
	Mtime int64  `json:"mtime" form:"mtime"` // 修改时间戳
}

// NoteSyncMtimeMessage 同步时用于更新 mtime 的消息结构。
type NoteSyncMtimeMessage struct {
	Path  string `json:"path" form:"path"`   // 路径
	Ctime int64  `json:"ctime" form:"ctime"` // 创建时间戳
	Mtime int64  `json:"mtime" form:"mtime"` // 修改时间戳
}

// ModifyMtimeFilesRequestParams 用于按 mtime 查询修改文件。
type ModifyMtimeFilesRequestParams struct {
	Vault string `json:"vault" form:"vault"  binding:"required"` // 仓库标识
	Mtime int64  `json:"mtime" form:"mtime"`                     // mtime 阈值
}

// NoteGetRequestParams 用于获取单条笔记的请求参数。
type NoteGetRequestParams struct {
	Vault    string `json:"vault" form:"vault" binding:"required"` // 仓库标识
	Path     string `json:"path" form:"path" binding:"required"`   // 路径
	PathHash string `json:"pathHash" form:"pathHash"`              // 路径哈希（可选）
}

// NoteListRequestParams 获取笔记列表的分页参数。
type NoteListRequestParams struct {
	Vault string `json:"vault" form:"vault" binding:"required"` // 仓库标识
}

// NoteGet 获取单条笔记
// 函数名: NoteGet
// 函数使用说明: 根据 vault 与 pathHash 获取单条笔记记录并转换为 service.Note 返回。
// 参数说明:
//   - uid int64: 用户ID
//   - params *NoteGetRequestParams: 包含 vault/path/pathHash 的请求参数
//
// 返回值说明:
//   - *Note: 笔记数据（包含 content）
//   - error: 出错时返回错误
func (svc *Service) NoteGet(uid int64, params *NoteGetRequestParams) (*Note, error) {
	var vaultID int64
	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_%d", uid), func() (any, error) {
		return svc.VaultIdGetByName(params.Vault, uid)
	})
	if err != nil {
		return nil, err
	}
	vaultID = vID.(int64)
	svc.SF.Do(fmt.Sprintf("Note_%d", uid), func() (any, error) {
		return nil, svc.dao.Note(uid)
	})
	note, err := svc.dao.NoteGetByPathHash(params.PathHash, vaultID, uid)
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(note, &Note{}).(*Note), nil
}

// NoteUpdateCheck 检查文件是否需要更新
// 函数名: NoteUpdateCheck
// 函数使用说明: 根据客户端传来的路径哈希、内容哈希和 mtime 对比服务端记录，决定是否需要服务端/客户端更新。
// 参数说明:
//   - uid int64: 用户ID，用于多租户/隔离数据
//   - params *NoteUpdateCheckRequestParams: 请求参数，包含 vault、path、pathHash、contentHash、ctime、mtime
//
// 返回值说明:
//   - bool: 是否新建
//   - bool: 是否需要服务端更新（即客户端应该 push）
//   - bool: 是否需要客户端修改其本地 mtime（即服务端比客户端新）
//   - *Note: 若存在服务端记录，返回该记录（struct Note）以便调用者使用；不存在则返回 nil
//   - error: 出错时返回错误
func (svc *Service) NoteUpdateCheck(uid int64, params *NoteUpdateCheckRequestParams) (bool, bool, bool, *Note, error) {

	var isNew bool
	var isUpdate bool
	var isUpdateOnlyMtime bool
	var vaultID int64
	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_%d", uid), func() (any, error) {
		return svc.VaultIdGetByName(params.Vault, uid)
	})
	if err != nil {
		return isNew, false, false, nil, err
	}
	vaultID = vID.(int64)

	// 检查数据表是否存在
	svc.SF.Do(fmt.Sprintf("Note_%d", uid), func() (any, error) {
		return nil, svc.dao.Note(uid)
	})

	note, _ := svc.dao.NoteGetByPathHash(params.PathHash, vaultID, uid)
	if note != nil {
		noteSvc := convert.StructAssign(note, &Note{}).(*Note)
		// 检查内容是否一致1
		if note.ContentHash == params.ContentHash {
			// 修改时间是否
			if params.Mtime < note.Mtime {
				isUpdate, isUpdateOnlyMtime = false, true
			}
		} else {
			isUpdate, isUpdateOnlyMtime = true, false
		}
		return isNew, isUpdate, isUpdateOnlyMtime, noteSvc, nil
	} else {
		isNew = true
		return isNew, isUpdate, isUpdateOnlyMtime, nil, nil
	}
}

/*
NoteModifyOrCreate
函数名: NoteModifyOrCreate
函数使用说明: 根据请求参数在数据库中创建或更新笔记。
  - 若记录不存在，则创建（action = "create"）
  - 若记录存在，则根据内容/mtime 做相应更新：
  - 当 mtimeCheck 为 true 且内容哈希与 mtime 均相同：不做任何操作返回 nil
  - 当 mtimeCheck 为 true 且服务器记录的 mtime 小于请求 mtime 但内容哈希相同：只更新 mtime
  - 否则执行更新（并设置 action = "modify" 或从 delete 变为 create）

参数说明:
  - uid int64: 用户ID
  - params *NoteModifyOrCreateRequestParams: 创建或修改所需参数（vault/path/pathHash/content/contentHash/ctime/mtime）
  - mtimeCheck bool: 是否启用 mtime 的特定检查逻辑（用于避免重复写入）

返回值说明:
  - *Note: 更新或创建后的笔记数据（转换为 service.Note 结构）
  - error: 出错则返回错误
*/
func (svc *Service) NoteModifyOrCreate(uid int64, params *NoteModifyOrCreateRequestParams, mtimeCheck bool) (bool, *Note, error) {

	var isNew bool
	var vaultID int64
	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_%d", uid), func() (any, error) {
		return svc.VaultIdGetByName(params.Vault, uid)
	})
	if err != nil {
		return isNew, nil, err
	}
	vaultID = vID.(int64)

	noteSet := &dao.NoteSet{
		VaultID:     vaultID,
		Path:        params.Path,
		PathHash:    params.PathHash,
		Content:     params.Content,
		ContentHash: params.ContentHash,
		Size:        int64(len(params.Content)),
		Mtime:       params.Mtime,
		Ctime:       params.Ctime,
	}

	svc.SF.Do(fmt.Sprintf("Note_%d", uid), func() (any, error) {
		return nil, svc.dao.Note(uid)
	})

	note, _ := svc.dao.NoteGetByPathHash(params.PathHash, vaultID, uid)
	if note != nil {
		isNew = false
		// 检查内容是否一致1
		if mtimeCheck && note.Mtime == params.Mtime && note.ContentHash == params.ContentHash {
			return isNew, nil, nil
		}
		// 检查内容是否一致 但是修改时间不同 则只更新修改时间
		if mtimeCheck && note.Mtime < params.Mtime && note.ContentHash == params.ContentHash {
			err := svc.dao.NoteUpdateMtime(params.Mtime, note.ID, uid)
			if err != nil {
				return isNew, nil, err
			}
			note.Mtime = params.Mtime
			rNote := convert.StructAssign(note, &Note{}).(*Note)
			return isNew, rNote, nil
		}
		if note.Action == "delete" {
			noteSet.Action = "create"
		} else {
			noteSet.Action = "modify"
		}

		noteDao, err := svc.dao.NoteUpdate(noteSet, note.ID, uid)
		if err != nil {
			return isNew, nil, err
		}
		svc.NoteCountSizeSum(vaultID, uid)
		rNote := convert.StructAssign(noteDao, &Note{}).(*Note)
		return isNew, rNote, nil
	} else {
		isNew = true
		noteSet.Action = "create"
		noteDao, err := svc.dao.NoteCreate(noteSet, uid)
		if err != nil {
			return isNew, nil, err
		}
		svc.NoteCountSizeSum(vaultID, uid)
		rNote := convert.StructAssign(noteDao, &Note{}).(*Note)
		return isNew, rNote, nil
	}

}

// NoteDelete 删除笔记
// 函数名: NoteDelete
// 函数使用说明: 将指定路径的笔记标记为删除（action = "delete"），并更新 vault 的统计信息。
// 参数说明:
//   - uid int64: 用户ID
//   - params *NoteDeleteRequestParams: 删除请求参数，包含 vault/path/pathHash
//
// 返回值说明:
//   - *Note: 更新后的笔记数据（service.Note）
//   - error: 出错时返回错误
func (svc *Service) NoteDelete(uid int64, params *NoteDeleteRequestParams) (*Note, error) {
	var vaultID int64
	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_%d", uid), func() (any, error) {
		return svc.VaultIdGetByName(params.Vault, uid)
	})
	if err != nil {
		return nil, err
	}
	vaultID = vID.(int64)

	svc.SF.Do(fmt.Sprintf("Note_%d", uid), func() (any, error) {
		return nil, svc.dao.Note(uid)
	})

	note, err := svc.dao.NoteGetByPathHash(params.PathHash, vaultID, uid)
	if err != nil {
		return nil, err
	}
	noteSet := &dao.NoteSet{
		VaultID:     vaultID,
		Action:      "delete",
		Path:        note.Path,
		PathHash:    note.PathHash,
		Content:     "",
		ContentHash: "",
		Size:        0,
	}
	noteDao, err := svc.dao.NoteUpdate(noteSet, note.ID, uid)
	if err != nil {
		return nil, err
	}
	svc.NoteCountSizeSum(vaultID, uid)
	rNote := convert.StructAssign(noteDao, &Note{}).(*Note)

	return rNote, nil
}

// NoteList 获取笔记列表（不包含 content）
// 函数名: NoteList
// 函数使用说明: 返回 vault 的笔记分页列表，但不包含 content 字段（用于列表展示）。
// 参数说明:
//   - uid int64: 用户ID
//   - params *NoteListRequestParams: 包含 vault/page/pageSize 的请求参数
//
// 返回值说明:
//   - []*NoteNoContent: 不包含 content 的笔记切片
//   - error: 出错时返回错误
func (svc *Service) NoteList(uid int64, params *NoteListRequestParams, pager *app.Pager) ([]*NoteNoContent, int, error) {
	var vaultID int64
	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_%d", uid), func() (any, error) {
		return svc.VaultIdGetByName(params.Vault, uid)
	})
	if err != nil {
		return nil, 0, err
	}
	vaultID = vID.(int64)

	svc.SF.Do(fmt.Sprintf("Note_%d", uid), func() (any, error) {
		return nil, svc.dao.Note(uid)
	})

	notes, err := svc.dao.NoteList(vaultID, pager.Page, pager.PageSize, uid)
	if err != nil {
		return nil, 0, err
	}

	count, err := svc.dao.NoteListCount(vaultID, uid)
	if err != nil {
		return nil, 0, err
	}

	var result []*NoteNoContent
	for _, n := range notes {
		result = append(result, convert.StructAssign(n, &NoteNoContent{}).(*NoteNoContent))
	}

	return result, int(count), nil
}

// NoteListByLastTime 获取在 lastTime 之后更新的笔记（去重 pathHash）
// 函数名: NoteListByLastTime
// 函数使用说明: 返回 vault 中 updated_timestamp 大于给定 lastTime 的笔记列表，结果去重（按 pathHash）。
// 参数说明:
//   - uid int64: 用户ID
//   - params *NoteSyncRequestParams: 包含 vault 与 lastTime 的请求参数
//
// 返回值说明:
//   - []*Note: 返回符合条件的笔记切片（包含 content）
//   - error: 出错时返回错误
func (svc *Service) NoteListByLastTime(uid int64, params *NoteSyncRequestParams) ([]*Note, error) {
	var vaultID int64
	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_%d", uid), func() (any, error) {
		return svc.VaultIdGetByName(params.Vault, uid)
	})
	if err != nil {
		return nil, err
	}
	vaultID = vID.(int64)

	svc.SF.Do(fmt.Sprintf("Note_%d", uid), func() (any, error) {
		return nil, svc.dao.Note(uid)
	})

	notes, err := svc.dao.NoteListByUpdatedTimestamp(params.LastTime, vaultID, uid)

	var results []*Note
	var cacheList map[string]bool = make(map[string]bool)
	for _, note := range notes {
		if cacheList[note.PathHash] {
			continue
		}
		results = append(results, convert.StructAssign(note, &Note{}).(*Note))
		cacheList[note.PathHash] = true
	}

	if err != nil {
		return nil, err
	}
	return results, nil
}

// NoteCountSizeSum 统计 vault 中笔记总数与总大小并更新 vault 记录
// 函数名: NoteCountSizeSum
// 函数使用说明: 调用 dao 层统计指定 vault 的笔记数量与大小，并写回 Vault 表。
// 参数说明:
//   - vaultID int64: vault 的 ID
//   - uid int64: 用户 ID
//
// 返回值说明:
//   - error: 出错时返回错误
func (svc *Service) NoteCountSizeSum(vaultID int64, uid int64) error {
	result, err := svc.dao.NoteCountSizeSum(vaultID, uid)
	if err != nil {
		return err
	}
	return svc.dao.VaultUpdateNoteCountSize(result.Size, result.Count, vaultID, uid)
}

// NoteCleanup 清理过期的软删除笔记
// 函数名: NoteCleanup
// 函数使用说明: 根据配置的保留时间，物理删除过期的软删除笔记。
// 参数说明:
//   - uid int64: 用户ID
//
// 返回值说明:
//   - error: 出错时返回错误
func (svc *Service) NoteCleanup(uid int64) error {
	// 获取保留时间配置
	retentionTimeStr := global.Config.App.DeleteNoteRetentionTime
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

	// 计算截止时间戳 (当前时间 - 保留时间)
	// 注意: UpdatedTimestamp 是毫秒级时间戳
	cutoffTime := time.Now().Add(-retentionDuration).UnixMilli()
	//转换成 2006-01-02 15:04:05.000
	cutoffTimeStr := time.UnixMilli(cutoffTime).Format("2006-01-02 15:04:05.000")

	global.Logger.Info("note cleanup", zap.Int64("uid", uid), zap.String("retention_time", retentionTimeStr), zap.String("cutoff_time", cutoffTimeStr))

	svc.SF.Do(fmt.Sprintf("Note_%d", uid), func() (any, error) {
		return nil, svc.dao.Note(uid)
	})

	return svc.dao.NoteDeletePhysicalByTime(cutoffTime, uid)
}

// NoteCleanupAll 清理所有用户的过期软删除笔记
func (svc *Service) NoteCleanupAll() error {
	// 获取保留时间配置
	retentionTimeStr := global.Config.App.DeleteNoteRetentionTime
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

	cleanupMutex.Lock()
	defer cleanupMutex.Unlock()

	// 动态计算检查间隔
	// 如果保留时间很短（< 1小时），则最小间隔为 1 分钟
	// 如果保留时间较长（>= 1小时），则间隔为保留时间的 1/10，但最大不超过 1 小时
	var checkInterval time.Duration
	if retentionDuration < time.Hour {
		checkInterval = time.Minute
	} else {
		checkInterval = retentionDuration / 10
		if checkInterval > time.Hour {
			checkInterval = time.Hour
		}
		if checkInterval < time.Minute {
			checkInterval = time.Minute
		}
	}

	// 如果距离上次清理时间不足检查间隔，则跳过
	if time.Since(lastCleanupTime) < checkInterval {
		return nil
	}

	uids, err := svc.dao.GetAllUserUIDs()
	if err != nil {
		return err
	}

	for _, uid := range uids {
		// 忽略单个用户的清理错误，继续清理下一个
		_ = svc.NoteCleanup(uid)
	}

	lastCleanupTime = time.Now()
	return nil
}
