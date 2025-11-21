package service

import (
	"fmt"

	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
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
	UpdatedAt        timex.Time `json:"updatedAt"`                                // 更新时间字段
	CreatedAt        timex.Time `json:"createdAt"`                                // 创建时间字段
}

type NoteNoContent struct {
	ID               int64      `json:"id" form:"id"`                             // 主键ID
	Action           string     `json:"action" form:"action"`                     // 操作
	Path             string     `json:"path" form:"path"`                         // 路径信息
	PathHash         string     `json:"pathHash" form:"pathHash"`                 // 路径哈希值
	Ctime            int64      `json:"ctime" form:"ctime"`                       // 创建时间戳
	Mtime            int64      `json:"mtime" form:"mtime"`                       // 修改时间戳
	UpdatedTimestamp int64      `json:"updatedTimestamp" form:"updatedTimestamp"` // 更新时间戳
	UpdatedAt        timex.Time `json:"updatedAt"`                                // 更新时间字段
	CreatedAt        timex.Time `json:"createdAt"`                                // 创建时间字段
}

type NoteUpdateCheckRequestParams struct {
	Vault       string `json:"vault" form:"vault"  binding:"required"`
	Path        string `json:"path" form:"path"  binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash"  binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash"  binding:""` // 内容哈希
	Ctime       int64  `json:"ctime" form:"ctime"  binding:"required"`     // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime"  binding:"required"`     // 修改时间戳
}

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

	note, _ := svc.dao.NoteGetByPathHash(params.PathHash, vaultID, uid)
	if note != nil {
		noteSvc := convert.StructAssign(note, &Note{}).(*Note)
		// 检查内容是否一致1
		if note.ContentHash == params.ContentHash {
			// 修改时间是否
			if params.Mtime > note.Mtime {
				return true, false, noteSvc, nil
			} else if params.Mtime < note.Mtime {
				return false, true, noteSvc, nil
			} else {
				return false, false, noteSvc, nil
			}
		}
		return true, false, noteSvc, nil
	} else {
		return true, false, nil, nil
	}
}

type NoteModifyOrCreateRequestParams struct {
	Vault       string `json:"vault" form:"vault"  binding:"required"`
	Path        string `json:"path" form:"path"  binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash"`
	Content     string `json:"content" form:"content"  binding:""`         // 内容详情
	ContentHash string `json:"contentHash" form:"contentHash"  binding:""` // 内容哈希
	Ctime       int64  `json:"ctime" form:"ctime" `                        // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime" `                        // 修改时间戳
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

	note, _ := svc.dao.NoteGetByPathHash(params.PathHash, vaultID, uid)
	if note != nil {
		// 检查内容是否一致1
		if mtimeCheck && note.Mtime == params.Mtime && note.ContentHash == params.ContentHash {
			return nil, nil
		}
		// 检查内容是否一致 但是修改时间不同 则只更新修改时间
		if mtimeCheck && note.Mtime < params.Mtime && note.ContentHash == params.ContentHash {
			err := svc.dao.NoteUpdateMtime(params.Mtime, note.ID, uid)
			if err != nil {
				return nil, err
			}
			note.Mtime = params.Mtime
			rNote := convert.StructAssign(note, &Note{}).(*Note)
			return rNote, nil
		}
		if note.Action == "delete" {
			noteSet.Action = "create"
		} else {
			noteSet.Action = "modify"
		}

		noteDao, err := svc.dao.NoteUpdate(noteSet, note.ID, uid)
		if err != nil {
			return nil, err
		}
		svc.NoteCountSizeSum(vaultID, uid)
		rNote := convert.StructAssign(noteDao, &Note{}).(*Note)
		return rNote, nil
	} else {

		noteSet.Action = "create"
		noteDao, err := svc.dao.NoteCreate(noteSet, uid)
		if err != nil {
			return nil, err
		}
		svc.NoteCountSizeSum(vaultID, uid)
		rNote := convert.StructAssign(noteDao, &Note{}).(*Note)
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
	PathHash string `json:"pathHash" form:"pathHash" `
}

// NoteDelete 删除笔记
func (svc *Service) NoteDelete(uid int64, params *NoteDeleteRequestParams) error {
	var vaultID int64
	// 单例模式获取VaultID
	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_%d", uid), func() (any, error) {
		return svc.VaultGetOrCreate(params.Vault, uid)
	})
	if err != nil {
		return err
	}
	vaultID = vID.(int64)

	svc.SF.Do(fmt.Sprintf("Note_%d", uid), func() (any, error) {
		return nil, svc.dao.Note(uid)
	})

	note, err := svc.dao.NoteGetByPathHash(params.PathHash, vaultID, uid)
	if err != nil {
		return err
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
	_, err = svc.dao.NoteUpdate(noteSet, note.ID, uid)
	if err != nil {
		return err
	}
	svc.NoteCountSizeSum(vaultID, uid)

	return nil
}

type NoteSyncCheckRequestParams struct {
	Path        string `json:"path" form:"path"` // 路径信息
	PathHash    string `json:"pathHash" form:"pathHash"  binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash"  binding:""` // 内容哈希
	Mtime       int64  `json:"mtime" form:"mtime"  binding:"required"`     // 修改时间戳
}
type NoteSyncRequestParams struct {
	Vault    string                       `json:"vault" form:"vault" binding:"required"`
	LastTime int64                        `json:"lastTime" form:"lastTime"`
	Notes    []NoteSyncCheckRequestParams `json:"notes" form:"notes"`
}

type NoteSyncEndMessage struct {
	Vault    string `json:"vault" form:"vault"`
	LastTime int64  `json:"lastTime" form:"lastTime"`
}

type NoteSyncNeedPushMessage struct {
	Path  string `json:"path" form:"path"`   // 路径信息
	Mtime int64  `json:"mtime" form:"mtime"` // 修改时间戳	UpdatedTimestamp int64  `json:"updatedTimestamp"
}

type NoteSyncMtimeMessage struct {
	Path  string `json:"path" form:"path"`   // 路径信息
	Ctime int64  `json:"ctime" form:"ctime"` // 创建时间戳
	Mtime int64  `json:"mtime" form:"mtime"` // 修改时间戳	UpdatedTimestamp int64  `json:"updatedTimestamp"
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

	notes, err := svc.dao.NoteListByMtime(params.Mtime, vaultID, uid)
	if err != nil {
		return nil, err
	}

	var results []*Note
	var cacheList map[string]bool = make(map[string]bool)
	for _, note := range notes {
		if cacheList[note.PathHash] {
			continue
		}
		results = append(results, convert.StructAssign(note, &Note{}).(*Note))
		cacheList[note.PathHash] = true
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

type NoteGetRequestParams struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash"`
}

// NoteGet 获取笔记
func (svc *Service) NoteGet(uid int64, params *NoteGetRequestParams) (*Note, error) {
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
	note, err := svc.dao.NoteGetByPathHash(params.PathHash, vaultID, uid)
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(note, &Note{}).(*Note), nil
}

type NoteListRequestParams struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
}

// NoteList 获取笔记列表
func (svc *Service) NoteList(uid int64, params *NoteListRequestParams) ([]*NoteNoContent, error) {
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

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	notes, err := svc.dao.NoteList(vaultID, params.Page, params.PageSize, uid)
	if err != nil {
		return nil, err
	}

	var result []*NoteNoContent
	for _, n := range notes {
		result = append(result, convert.StructAssign(n, &NoteNoContent{}).(*NoteNoContent))
	}

	return result, nil
}
