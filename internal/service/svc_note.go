package service

import (
	"fmt"

	"github.com/haierkeys/obsidian-better-sync-service/internal/dao"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/convert"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/timex"
)

type Note struct {
	ID               int64      `json:"id" form:"id"`                             // 主键ID
	Action           string     `json:"action" form:"action"`                     // 操作
	Path             string     `json:"path" form:"path"`                         // 路径信息
	PathHash         string     `json:"pathHash" form:"pathHash"`                 // 路径哈希值
	Content          string     `json:"content" form:"content"`                   // 内容详情
	ContentHash      string     `json:"contentHash" form:"contentHash"`           // 内容哈希
	Ctime            int64      `json:"ctime" form:"ctime"`                       // 创建时间戳
	Mtime            int64      `json:"mtime" form:"mtime"`                       // 修改时间戳
	UpdatedTimestamp int64      `json:"updatedTimestamp" form:"updatedTimestamp"` // 更新时间戳
	CreatedAt        timex.Time `json:"createdAt" form:"createdAt"`               // 创建时间，自动填充当前时间
	UpdatedAt        timex.Time `json:"updatedAt" form:"updatedAt"`               // 更新时间，自动填充当前时间
}

type NoteUpdateCheckRequestParams struct {
	Vault       string `json:"vault" form:"vault"  binding:"required"`
	Path        string `json:"path" form:"path"  binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash"  binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash"  binding:""` // 内容哈希
	Ctime       int64  `json:"ctime" form:"ctime"  binding:"required"`     // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime"  binding:"required"`     // 修改时间戳
}

/**
* NoteUpdateCheck
* @Description        修改文件
* @Create             HaierKeys 2025-03-01 17:30
* @Param              params  *NoteModifyRequestParams  文件修改请求参数
* @Return             error  错误信息
@Return             error  错误信息
@Return             error  错误信息
*/

/*
* NoteUpdateCheck 检查文件是否需要更新
* @Return             bool  是否需要更新
* @Return             bool  是否需要客户端修改更新修改时间
* @Return             error  错误信息
 */
func (svc *Service) NoteUpdateCheck(uid int64, params *NoteUpdateCheckRequestParams) (bool, bool, *Note, error) {

	var vaultID int64
	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_%d", uid), func() (any, error) {
		return svc.VaultGetOrCreate(params.Vault, uid)
	})
	if err != nil {
		return false, false, nil, err
	}
	vaultID = vID.(int64)

	// 检查数据表是否存在
	svc.SF.Do(fmt.Sprintf("Note_%d", uid), func() (any, error) {
		return nil, svc.dao.Note(uid)
	})

	node, _ := svc.dao.NoteGetByPathHash(params.PathHash, vaultID, uid)
	if node != nil {
		nodeSvc := convert.StructAssign(node, &Note{}).(*Note)
		// 检查内容是否一致1
		if node.ContentHash == params.ContentHash {
			// 修改时间是否
			if params.Mtime > node.Mtime {
				return true, false, nodeSvc, nil
			} else if params.Mtime < node.Mtime {
				return false, true, nodeSvc, nil
			} else {
				return false, false, nodeSvc, nil
			}
		}
		return true, false, nodeSvc, nil
	} else {
		return true, false, nil, nil
	}
}

type NoteModifyOrCreateRequestParams struct {
	Vault       string `json:"vault" form:"vault"  binding:"required"`
	Path        string `json:"path" form:"path"  binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash"  binding:"required"`
	Content     string `json:"content" form:"content"  binding:""`         // 内容详情
	ContentHash string `json:"contentHash" form:"contentHash"  binding:""` // 内容哈希
	Ctime       int64  `json:"ctime" form:"ctime"  binding:"required"`     // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime"  binding:"required"`     // 修改时间戳
}

/**
* NoteModify
* @Description        修改文件
* @Create             HaierKeys 2025-03-01 17:30
* @Param              params  *NoteModifyRequestParams  文件修改请求参数
* @Return             error  错误信息
 */

func (svc *Service) NoteModifyOrCreate(uid int64, params *NoteModifyOrCreateRequestParams, mtimeCheck bool) (*Note, error) {

	var vaultID int64
	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_%d", uid), func() (any, error) {
		return svc.VaultGetOrCreate(params.Vault, uid)
	})
	if err != nil {
		return nil, err
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

	node, _ := svc.dao.NoteGetByPathHash(params.PathHash, vaultID, uid)
	if node != nil {
		// 检查内容是否一致1
		if mtimeCheck && node.Mtime == params.Mtime && node.ContentHash == params.ContentHash {
			return nil, nil
		}
		// 检查内容是否一致 但是修改时间不同 则只更新修改时间
		if mtimeCheck && node.Mtime < params.Mtime && node.ContentHash == params.ContentHash {
			err := svc.dao.NoteUpdateMtime(params.Mtime, node.ID, uid)
			if err != nil {
				return nil, err
			}
			node.Mtime = params.Mtime
			rNote := convert.StructAssign(node, &Note{}).(*Note)
			return rNote, nil
		}
		if node.Action == "delete" {
			noteSet.Action = "create"
		} else {
			noteSet.Action = "modify"
		}

		nodeDao, err := svc.dao.NoteUpdate(noteSet, node.ID, uid)
		if err != nil {
			return nil, err
		}
		svc.NoteCountSizeSum(vaultID, uid)
		rNote := convert.StructAssign(nodeDao, &Note{}).(*Note)
		return rNote, nil
	} else {

		noteSet.Action = "create"
		nodeDao, err := svc.dao.NoteCreate(noteSet, uid)
		if err != nil {
			return nil, err
		}
		svc.NoteCountSizeSum(vaultID, uid)
		rNote := convert.StructAssign(nodeDao, &Note{}).(*Note)
		return rNote, nil
	}

}

type ContentModifyRequestParams struct {
	Vault       string `json:"vault" form:"vault"  binding:"required"`
	Path        string `json:"path" form:"path"  binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash"  binding:"required"`
	Content     string `json:"content" form:"content"  binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash"  binding:"required"` // 内容哈希
	Ctime       int64  `json:"ctime" form:"ctime"  binding:"required"`             // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime"  binding:"required"`             // 修改时间戳
}

type NoteDeleteRequestParams struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash" binding:"required"`
}

// NoteDelete 删除笔记
func (svc *Service) NoteDelete(uid int64, params *NoteDeleteRequestParams) (*Note, error) {
	var vaultID int64
	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_%d", uid), func() (any, error) {
		return svc.VaultGetOrCreate(params.Vault, uid)
	})
	if err != nil {
		return nil, err
	}
	vaultID = vID.(int64)

	svc.SF.Do(fmt.Sprintf("Note_%d", uid), func() (any, error) {
		return nil, svc.dao.Note(uid)
	})

	node, err := svc.dao.NoteGetByPathHash(params.PathHash, vaultID, uid)
	if err != nil {
		return nil, err
	}
	note := &dao.NoteSet{
		VaultID:     vaultID,
		Action:      "delete",
		Path:        node.Path,
		PathHash:    node.PathHash,
		Content:     "",
		ContentHash: "",
		Size:        0,
	}
	nodeDao, err := svc.dao.NoteUpdate(note, node.ID, uid)
	if err != nil {
		return nil, err
	}
	svc.NoteCountSizeSum(vaultID, uid)
	rNote := convert.StructAssign(nodeDao, &Note{}).(*Note)

	return rNote, nil
}

type NoteSyncRequestParams struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	LastTime int64  `json:"lastTime" form:"lastTime"`
}

type NoteSyncEndMessage struct {
	Vault    string `json:"vault" form:"vault"`
	LastTime int64  `json:"lastTime" form:"lastTime"`
}

// ModifyFiles 获取修改的文件列表
func (svc *Service) NoteListByLastTime(uid int64, params *NoteSyncRequestParams) ([]*Note, error) {
	var vaultID int64
	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_%d", uid), func() (any, error) {
		return svc.VaultGetOrCreate(params.Vault, uid)
	})
	if err != nil {
		return nil, err
	}
	vaultID = vID.(int64)

	svc.SF.Do(fmt.Sprintf("Note_%d", uid), func() (any, error) {
		return nil, svc.dao.Note(uid)
	})

	nodes, err := svc.dao.NoteListByUpdatedTimestamp(params.LastTime, vaultID, uid)

	var results []*Note
	for _, node := range nodes {
		results = append(results, convert.StructAssign(node, &Note{}).(*Note))
	}

	if err != nil {
		return nil, err
	}
	return results, nil
}

type ModifyMtimeFilesRequestParams struct {
	Vault string `json:"vault" form:"vault"  binding:"required"`
	Mtime int64  `json:"mtime" form:"mtime"`
}

// ModifyFiles 获取修改的文件列表
func (svc *Service) NoteListByMtime(uid int64, params *ModifyMtimeFilesRequestParams) ([]*Note, error) {
	var vaultID int64
	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_%d", uid), func() (any, error) {
		return svc.VaultGetOrCreate(params.Vault, uid)
	})
	if err != nil {
		return nil, err
	}
	vaultID = vID.(int64)

	svc.SF.Do(fmt.Sprintf("Note_%d", uid), func() (any, error) {
		return nil, svc.dao.Note(uid)
	})

	nodes, err := svc.dao.NoteListByMtime(params.Mtime, vaultID, uid)
	if err != nil {
		return nil, err
	}

	var results []*Note
	for _, node := range nodes {
		results = append(results, convert.StructAssign(node, &Note{}).(*Note))
	}

	return results, nil
}

func (svc *Service) NoteCountSizeSum(vaultID int64, uid int64) error {
	result, err := svc.dao.NoteCountSizeSum(vaultID, uid)
	if err != nil {
		return err
	}
	return svc.dao.VaultUpdateNoteCountSize(result.Size, result.Count, vaultID, uid)
}
