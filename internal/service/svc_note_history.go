package service

import (
	"fmt"

	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// NoteHistoryMsg 笔记历史记录延时处理消息
type NoteHistoryMsg struct {
	NoteID int64
	UID    int64
}

var (
	// NoteHistoryChannel 延时任务通道，后台 task 会监听此通道
	NoteHistoryChannel = make(chan NoteHistoryMsg, 1000)
)

// NoteHistoryDelayPush 将笔记推送至延时处理队列
func NoteHistoryDelayPush(noteID int64, uid int64) {
	NoteHistoryChannel <- NoteHistoryMsg{NoteID: noteID, UID: uid}
}

// NoteHistory 笔记历史记录展示结构体
type NoteHistory struct {
	ID        int64  `json:"id" form:"id"`
	NoteID    int64  `json:"noteId" form:"noteId"`
	VaultId   int64  `json:"vaultId" form:"vaultId"`
	Path      string `json:"path" form:"path"`
	Content   string `json:"content" form:"content"`
	Client    string `json:"client" form:"client"`
	Version   int64  `json:"version" form:"version"`
	CreatedAt string `json:"createdAt" form:"createdAt"`
}

// NoteHistoryListRequestParams 笔记历史列表请求参数
type NoteHistoryListRequestParams struct {
	NoteID int64 `json:"noteId" form:"noteId" binding:"required"`
}

// NoteHistorySave 手动保存笔记历史版本（直接保存全量内容）
func (svc *Service) NoteHistorySave(uid int64, note *dao.Note, client string) error {
	svc.SF.Do(fmt.Sprintf("NoteHistory_%d", uid), func() (any, error) {
		return nil, svc.dao.NoteHistoryAutoMigrate(uid)
	})

	latestVersion, err := svc.dao.NoteHistoryGetLatestVersion(note.ID, uid)
	if err != nil {
		return err
	}

	params := &dao.NoteHistorySet{
		NoteID:  note.ID,
		VaultID: note.VaultID,
		Path:    note.Path,
		Content: note.Content,
		Client:  client,
		Version: latestVersion + 1,
	}

	_, err = svc.dao.NoteHistoryCreate(params, uid)
	return err
}

// NoteHistoryList 获取指定笔记的历史版本列表
func (svc *Service) NoteHistoryList(uid int64, params *NoteHistoryListRequestParams, pager *app.Pager) ([]*NoteHistory, int64, error) {
	svc.SF.Do(fmt.Sprintf("NoteHistory_%d", uid), func() (any, error) {
		return nil, svc.dao.NoteHistoryAutoMigrate(uid)
	})

	list, count, err := svc.dao.NoteHistoryListByNoteId(params.NoteID, pager.Page, pager.PageSize, uid)
	if err != nil {
		return nil, 0, err
	}

	var results []*NoteHistory
	for _, m := range list {
		results = append(results, convert.StructAssign(m, &NoteHistory{}).(*NoteHistory))
	}
	return results, count, nil
}

// NoteHistoryGet 获取指定 ID 的笔记历史详情
func (svc *Service) NoteHistoryGet(uid int64, id int64) (*NoteHistory, error) {
	svc.SF.Do(fmt.Sprintf("NoteHistory_%d", uid), func() (any, error) {
		return nil, svc.dao.NoteHistoryAutoMigrate(uid)
	})

	history, err := svc.dao.NoteHistoryGetById(id, uid)
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(history, &NoteHistory{}).(*NoteHistory), nil
}

// NoteHistoryProcessDelay 延时处理笔记历史（计算 diff 并保存补丁版本）
// 该方法会被后台任务 `NoteHistoryTask` 触发
func (svc *Service) NoteHistoryProcessDelay(noteID int64, uid int64) error {
	note, err := svc.dao.NoteGetById(noteID, uid)
	if err != nil {
		return err
	}

	if note.Content == note.ContentLastSnapshot {
		return nil
	}

	// 计算 diff
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(note.ContentLastSnapshot, note.Content, false)
	patchText := dmp.PatchToText(dmp.PatchMake(note.ContentLastSnapshot, diffs))

	latestVersion, err := svc.dao.NoteHistoryGetLatestVersion(note.ID, uid)
	if err != nil {
		return err
	}

	params := &dao.NoteHistorySet{
		NoteID:  note.ID,
		VaultID: note.VaultID,
		Path:    note.ContentLastSnapshot, // 复用 Path 存储快照内容
		Content: patchText,                // 复用 Content 存储补丁
		Client:  "system-task",
		Version: latestVersion + 1,
	}

	_, err = svc.dao.NoteHistoryCreate(params, uid)
	if err != nil {
		return err
	}

	// 更新 ContentLastSnapshot
	return svc.dao.NoteUpdateSnapshot(note.Content, note.ID, uid)
}
