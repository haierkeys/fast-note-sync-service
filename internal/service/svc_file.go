package service

import (
	"fmt"
	"os"

	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
)

var (
	lastFileCleanupTime time.Time
)

// File 表示文件的完整数据结构。
type File struct {
	ID               int64      `json:"id" form:"id"`                         // 主键ID
	Action           string     `json:"-" form:"action"`                      // 操作：create/modify/delete
	Path             string     `json:"path" form:"path"`                     // 路径信息（文件路径）
	PathHash         string     `json:"pathHash" form:"pathHash"`             // 路径哈希值，用于快速查找
	ContentHash      string     `json:"contentHash" form:"contentHash"`       // 内容哈希，用于判定内容是否变更
	SavePath         string     `json:"savePath" form:"savePath"  binding:""` // 文件保存路径
	Size             int64      `json:"size" form:"size"`                     // 文件大小
	Ctime            int64      `json:"ctime" form:"ctime"`                   // 创建时间戳（秒）
	Mtime            int64      `json:"mtime" form:"mtime"`                   // 文件修改时间戳（秒）
	UpdatedTimestamp int64      `json:"lastTime" form:"updatedTimestamp"`     // 记录更新时间戳（用于同步）
	UpdatedAt        timex.Time `json:"-"`                                    // 更新时间字段（time 类型包装）
	CreatedAt        timex.Time `json:"-"`                                    // 创建时间字段（time 类型包装）
}

// FileUpdateCheckParams 客户端用于检查是否需要更新的请求参数。
type FileUpdateCheckParams struct {
	Vault       string `json:"vault" form:"vault"  binding:"required"`       // 仓库标识
	Path        string `json:"path" form:"path"  binding:"required"`         // 路径
	PathHash    string `json:"pathHash" form:"pathHash"  binding:"required"` // 路径哈希
	ContentHash string `json:"contentHash" form:"contentHash"  binding:""`   // 内容哈希（可选）
	Size        int64  `json:"size" form:"size" binding:""`                  // 大小
	Ctime       int64  `json:"ctime" form:"ctime"  binding:"required"`       // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime"  binding:"required"`       // 修改时间戳
}

// FileModifyOrCreateRequestParams 用于创建或修改文件的请求参数。
type FileUpdateParams struct {
	Vault       string `json:"vault" form:"vault"  binding:"required"`     // 仓库标识
	Path        string `json:"path" form:"path"  binding:"required"`       // 路径
	PathHash    string `json:"pathHash" form:"pathHash"`                   // 路径哈希
	ContentHash string `json:"contentHash" form:"contentHash"  binding:""` // 内容哈希（可选）
	SavePath    string `json:"savePath" form:"savePath"  binding:""`       // 文件保存路径
	Size        int64  `json:"size" form:"size" `                          // 文件大小
	Ctime       int64  `json:"ctime" form:"ctime" `                        // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime" `                        // 修改时间戳
}

// FileDeleteRequestParams 删除文件所需参数。
type FileDeleteParams struct {
	Vault    string `json:"vault" form:"vault" binding:"required"` // 仓库标识
	Path     string `json:"path" form:"path" binding:"required"`   // 路径
	PathHash string `json:"pathHash" form:"pathHash" `             // 路径哈希（可选）
}

// FileSyncCheckRequestParams 同步检查单条记录的参数。
type FileSyncCheckRequestParams struct {
	Path        string `json:"path" form:"path"`                             // 路径
	PathHash    string `json:"pathHash" form:"pathHash"  binding:"required"` // 路径哈希
	ContentHash string `json:"contentHash" form:"contentHash"  binding:""`   // 内容哈希（可选）
	Mtime       int64  `json:"mtime" form:"mtime"  binding:"required"`       // 修改时间戳
}

// FileSyncRequestParams 同步请求主体，包含 vault、上次同步时间和要检查的文件列表。
type FileSyncParams struct {
	Vault    string                       `json:"vault" form:"vault" binding:"required"` // 仓库标识
	LastTime int64                        `json:"lastTime" form:"lastTime"`              // 上次同步时间戳
	Files    []FileSyncCheckRequestParams `json:"files" form:"files"`                    // 客户端文件信息列表
}

type FileUploadCompleteParams struct {
	SessionID string `json:"sessionId" binding:"required"`
}

// FileGetRequestParams 用于获取单条文件的请求参数。
type FileGetParams struct {
	Vault    string `json:"vault" form:"vault" binding:"required"` // 仓库标识
	Path     string `json:"path" form:"path" binding:"required"`   // 路径
	PathHash string `json:"pathHash" form:"pathHash"`              // 路径哈希（可选）
}

// FileListRequestParams 获取文件列表的分页参数。
type FileListParams struct {
	Vault string `json:"vault" form:"vault" binding:"required"` // 仓库标识
}

// FileGet 获取单条文件
// 函数名: FileGet
// 函数使用说明: 根据 vault 与 pathHash 获取单条文件记录并转换为 service.File 返回。
// 参数说明:
//   - uid int64: 用户ID
//   - params *FileGetRequestParams: 包含 vault/path/pathHash 的请求参数
//
// 返回值说明:
//   - *File: 文件数据
//   - error: 出错时返回错误
func (svc *Service) FileGet(uid int64, params *FileGetParams) (*File, error) {
	var vaultID int64
	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_Get_%d", uid), func() (any, error) {
		return svc.VaultIdGetByName(params.Vault, uid)
	})
	if err != nil {
		return nil, err
	}
	vaultID = vID.(int64)
	file, err := svc.dao.FileGetByPathHash(params.PathHash, vaultID, uid)
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(file, &File{}).(*File), nil
}

// FileUpdateCheck 检查文件是否需要更新
// 根据客户端传来的路径哈希、内容哈希和 mtime 对比服务端记录，决定是否需要服务端/客户端更新。
// 参数说明:
//   - uid int64: 用户ID，用于多租户/隔离数据
//   - params *FileUpdateCheckParams: 请求参数，包含 vault、path、pathHash、contentHash、ctime、mtime
//
// 返回值说明:
//   - updateMode: 更新模式，可选值为 "Create"（创建）、"UpdateContent"（更新内容）、"UpdateMtime"（需要用户更新 mtime）
//   - *File: 若存在服务端记录，返回该记录（struct File）以便调用者使用；不存在则返回 nil
//   - error: 出错时返回错误
//

func (svc *Service) FileUpdateCheck(uid int64, params *FileUpdateCheckParams) (string, *File, error) {

	var vaultID int64

	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_Get_%d", uid), func() (any, error) {
		return svc.VaultIdGetByName(params.Vault, uid)
	})
	if err != nil {
		return "", nil, err
	}
	vaultID = vID.(int64)

	if file, _ := svc.dao.FileGetByPathHash(params.PathHash, vaultID, uid); file != nil {
		fileSvc := convert.StructAssign(file, &File{}).(*File)
		// 检查内容是否一致
		if file.ContentHash == params.ContentHash {
			// 当用户 mtime 小于服务端 mtime 时，通知用户更新mtime
			if params.Mtime < file.Mtime {
				return "UpdateMtime", fileSvc, nil
			} else if params.Mtime > file.Mtime {
				svc.dao.FileUpdateMtime(params.Mtime, file.ID, uid)
			}
			return "", fileSvc, nil
		} else {
			return "UpdateContent", fileSvc, nil
		}
	} else {
		return "Create", nil, nil
	}
}

/*
FileModifyOrCreate
函数名: FileModifyOrCreate
函数使用说明: 根据请求参数在数据库中创建或更新文件。
  - 若记录不存在，则创建（action = "create"）
  - 若记录存在，则根据内容/mtime 做相应更新：
  - 当 mtimeCheck 为 true 且内容哈希与 mtime 均相同：不做任何操作返回 nil
  - 当 mtimeCheck 为 true 且服务器记录的 mtime 小于请求 mtime 但内容哈希相同：只更新 mtime
  - 否则执行更新（并设置 action = "modify" 或从 delete 变为 create）

参数说明:
  - uid int64: 用户ID
  - params *FileModifyOrCreateRequestParams: 创建或修改所需参数（vault/path/pathHash/contentHash/savePath/size/ctime/mtime）
  - mtimeCheck bool: 是否启用 mtime 的特定检查逻辑（用于避免重复写入）

返回值说明:
  - *File: 更新或创建后的文件数据（转换为 service.File 结构）
  - error: 出错则返回错误
*/
func (svc *Service) FileUpdateOrCreate(uid int64, params *FileUpdateParams, mtimeCheck bool) (bool, *File, error) {

	var isNew bool
	var vaultID int64
	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_Get_%d", uid), func() (any, error) {
		return svc.VaultIdGetByName(params.Vault, uid)
	})
	if err != nil {
		return isNew, nil, err
	}
	vaultID = vID.(int64)

	fileSet := &dao.FileSet{
		VaultID:     vaultID,
		Path:        params.Path,
		PathHash:    params.PathHash,
		ContentHash: params.ContentHash,
		SavePath:    params.SavePath,
		Size:        params.Size,
		Mtime:       params.Mtime,
		Ctime:       params.Ctime,
	}

	file, _ := svc.dao.FileGetByPathHash(params.PathHash, vaultID, uid)
	if file != nil {
		isNew = false
		// 检查内容是否一致
		if mtimeCheck && file.Mtime == params.Mtime && file.ContentHash == params.ContentHash {
			return isNew, nil, nil
		}
		// 检查内容是否一致 但是修改时间不同 则只更新修改时间
		if mtimeCheck && file.Mtime < params.Mtime && file.ContentHash == params.ContentHash {
			err := svc.dao.FileUpdateMtime(params.Mtime, file.ID, uid)
			if err != nil {
				return isNew, nil, err
			}
			file.Mtime = params.Mtime
			rFile := convert.StructAssign(file, &File{}).(*File)
			return isNew, rFile, nil
		}
		if file.Action == "delete" {
			fileSet.Action = "create"
		} else {
			fileSet.Action = "modify"
		}

		fileDao, err := svc.dao.FileUpdate(fileSet, file.ID, uid)
		if err != nil {
			return isNew, nil, err
		}
		svc.FileCountSizeSum(vaultID, uid)
		rFile := convert.StructAssign(fileDao, &File{}).(*File)
		return isNew, rFile, nil
	} else {
		isNew = true
		fileSet.Action = "create"
		fileDao, err := svc.dao.FileCreate(fileSet, uid)
		if err != nil {
			return isNew, nil, err
		}
		svc.FileCountSizeSum(vaultID, uid)
		rFile := convert.StructAssign(fileDao, &File{}).(*File)
		return isNew, rFile, nil
	}

}

// FileDelete 删除文件
// 函数名: FileDelete
// 函数使用说明: 将指定路径的文件标记为删除（action = "delete"），并更新 vault 的统计信息。
// 参数说明:
//   - uid int64: 用户ID
//   - params *FileDeleteParams: 删除请求参数，包含 vault/path/pathHash
//
// 返回值说明:
//   - *File: 更新后的文件数据（service.File）
//   - error: 出错时返回错误
func (svc *Service) FileDelete(uid int64, params *FileDeleteParams) (*File, error) {
	var vaultID int64
	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_Get_%d", uid), func() (any, error) {
		return svc.VaultIdGetByName(params.Vault, uid)
	})
	if err != nil {
		return nil, err
	}
	vaultID = vID.(int64)

	file, err := svc.dao.FileGetByPathHash(params.PathHash, vaultID, uid)
	if err != nil {
		return nil, err
	}
	savePath := file.SavePath
	fileSet := &dao.FileSet{
		VaultID:     vaultID,
		Action:      "delete",
		Path:        file.Path,
		PathHash:    file.PathHash,
		ContentHash: "",
		SavePath:    "",
		Size:        0,
		Mtime:       0,
		Ctime:       0,
	}
	fileDao, err := svc.dao.FileUpdate(fileSet, file.ID, uid)
	if err != nil {
		return nil, err
	}
	// 删除文件
	err = os.Remove(savePath)
	if err != nil {
		return nil, err
	}
	svc.FileCountSizeSum(vaultID, uid)
	rFile := convert.StructAssign(fileDao, &File{}).(*File)

	return rFile, nil
}

// FileList 获取文件列表
// 函数名: FileList
// 函数使用说明: 返回 vault 的文件分页列表。
// 参数说明:
//   - uid int64: 用户ID
//   - params *FileListRequestParams: 包含 vault/page/pageSize 的请求参数
//
// 返回值说明:
//   - []*File: 文件切片
//   - error: 出错时返回错误
func (svc *Service) FileList(uid int64, params *FileListParams, pager *app.Pager) ([]*File, int, error) {
	var vaultID int64
	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_Get_%d", uid), func() (any, error) {
		return svc.VaultIdGetByName(params.Vault, uid)
	})
	if err != nil {
		return nil, 0, err
	}
	vaultID = vID.(int64)

	files, err := svc.dao.FileList(vaultID, pager.Page, pager.PageSize, uid)
	if err != nil {
		return nil, 0, err
	}

	count, err := svc.dao.FileListCount(vaultID, uid)
	if err != nil {
		return nil, 0, err
	}

	var result []*File
	for _, f := range files {
		result = append(result, convert.StructAssign(f, &File{}).(*File))
	}

	return result, int(count), nil
}

// FileListByLastTime 获取在 lastTime 之后更新的文件（去重 pathHash）
// 函数名: FileListByLastTime
// 函数使用说明: 返回 vault 中 updated_timestamp 大于给定 lastTime 的文件列表，结果去重（按 pathHash）。
// 参数说明:
//   - uid int64: 用户ID
//   - params *FileSyncRequestParams: 包含 vault 与 lastTime 的请求参数
//
// 返回值说明:
//   - []*File: 返回符合条件的文件切片
//   - error: 出错时返回错误
func (svc *Service) FileListByLastTime(uid int64, params *FileSyncParams) ([]*File, error) {
	var vaultID int64
	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_Get_%d", uid), func() (any, error) {
		return svc.VaultIdGetByName(params.Vault, uid)
	})
	if err != nil {
		return nil, err
	}
	vaultID = vID.(int64)

	files, err := svc.dao.FileListByUpdatedTimestamp(params.LastTime, vaultID, uid)

	var results []*File
	var cacheList map[string]bool = make(map[string]bool)
	for _, file := range files {
		if cacheList[file.PathHash] {
			continue
		}
		results = append(results, convert.StructAssign(file, &File{}).(*File))
		cacheList[file.PathHash] = true
	}

	if err != nil {
		return nil, err
	}
	return results, nil
}

// FileCountSizeSum 统计 vault 中文件总数与总大小并更新 vault 记录
// 函数名: FileCountSizeSum
// 函数使用说明: 调用 dao 层统计指定 vault 的文件数量与大小，并写回 Vault 表。
// 参数说明:
//   - vaultID int64: vault 的 ID
//   - uid int64: 用户 ID
//
// 返回值说明:
//   - error: 出错时返回错误
func (svc *Service) FileCountSizeSum(vaultID int64, uid int64) error {
	result, err := svc.dao.FileCountSizeSum(vaultID, uid)
	if err != nil {
		return err
	}
	return svc.dao.VaultUpdateFileCountSize(result.Size, result.Count, vaultID, uid)
}

// FileCleanup 清理过期的软删除文件
// 函数名: FileCleanup
// 函数使用说明: 根据配置的保留时间，物理删除过期的软删除文件。
// 参数说明:
//   - uid int64: 用户ID
//
// 返回值说明:
//   - error: 出错时返回错误
func (svc *Service) FileCleanup(uid int64) error {
	// 获取保留时间配置
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

	// 计算截止时间戳 (当前时间 - 保留时间)
	// 注意: UpdatedTimestamp 是毫秒级时间戳
	cutoffTime := time.Now().Add(-retentionDuration).UnixMilli()

	return svc.dao.FileDeletePhysicalByTime(cutoffTime, uid)
}

// FileCleanupAll 清理所有用户的过期软删除文件
func (svc *Service) FileCleanupAll() error {
	// 获取保留时间配置
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
	if time.Since(lastFileCleanupTime) < checkInterval {
		return nil
	}

	uids, err := svc.dao.GetAllUserUIDs()
	if err != nil {
		return err
	}

	for _, uid := range uids {
		// 忽略单个用户的清理错误，继续清理下一个
		_ = svc.FileCleanup(uid)
	}

	lastFileCleanupTime = time.Now()
	return nil
}
