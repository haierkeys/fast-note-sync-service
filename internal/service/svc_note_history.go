package service

import (
	"fmt"

	"github.com/haierkeys/fast-note-sync-service/global"
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
	ID         int64  `json:"id" form:"id"`
	NoteID     int64  `json:"noteId" form:"noteId"`
	VaultID    int64  `json:"vaultId" form:"vaultId"`
	Path       string `json:"path" form:"path"`
	DiffPatch  string `json:"diffPatch" form:"diffPatch"`
	Content    string `json:"content" form:"content"`
	ClientName string `json:"clientName" form:"clientName"`
	Version    int64  `json:"version" form:"version"`
	CreatedAt  string `json:"createdAt" form:"createdAt"`
}

type NoteHistoryNoContent struct {
	ID         int64  `json:"id" form:"id"`
	NoteID     int64  `json:"noteId" form:"noteId"`
	VaultID    int64  `json:"vaultId" form:"vaultId"`
	Path       string `json:"path" form:"path"`
	DiffPatch  string `json:"diffPatch" form:"diffPatch"`
	ClientName string `json:"clientName" form:"clientName"`
	Version    int64  `json:"version" form:"version"`
	CreatedAt  string `json:"createdAt" form:"createdAt"`
}

// NoteHistoryListRequestParams 笔记历史列表请求参数
type NoteHistoryListRequestParams struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash"`
}

// NoteHistoryList 获取指定笔记的历史版本列表
func (svc *Service) NoteHistoryList(uid int64, params *NoteHistoryListRequestParams, pager *app.Pager) ([]*NoteHistoryNoContent, int64, error) {

	// get note id
	note, err := svc.NoteGet(uid, &NoteGetRequestParams{
		Vault:    params.Vault,
		Path:     params.Path,
		PathHash: params.PathHash,
	})
	if err != nil {
		return nil, 0, err
	}

	if note == nil {
		return nil, 0, fmt.Errorf("note not found")
	}

	list, count, err := svc.dao.NoteHistoryListByNoteId(note.ID, pager.Page, pager.PageSize, uid)
	if err != nil {
		return nil, 0, err
	}
	global.Dump(list)

	var results []*NoteHistoryNoContent
	for _, m := range list {
		results = append(results, convert.StructAssign(m, &NoteHistoryNoContent{}).(*NoteHistoryNoContent))
	}
	return results, count, nil
}

// NoteHistoryGet 获取指定 ID 的笔记历史详情
func (svc *Service) NoteHistoryGet(uid int64, id int64) (*NoteHistory, error) {

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
		return fmt.Errorf("svc.dao.NoteGetById: %w", err)
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
		return fmt.Errorf("svc.dao.NoteHistoryGetLatestVersion: %w", err)
	}

	params := &dao.NoteHistorySet{
		NoteID:     note.ID,
		VaultID:    note.VaultID,
		Path:       note.Path, // 存储diff补丁
		DiffPatch:  patchText,
		Content:    note.ContentLastSnapshot, // 快照
		ClientName: note.ClientName,
		Version:    latestVersion + 1,
		CreatedAt:  note.UpdatedAt,
		UpdatedAt:  note.UpdatedAt,
	}

	_, err = svc.dao.NoteHistoryCreate(params, uid)
	if err != nil {
		return fmt.Errorf("svc.dao.NoteHistoryCreate: %w", err)
	}

	// 更新 ContentLastSnapshot
	return svc.dao.NoteUpdateSnapshot(note.Content, latestVersion+1, note.ID, uid)
}

// NoteHistoryMigrate 处理笔记历史迁移
func (svc *Service) NoteHistoryMigrate(oldNoteID, newNoteID int64, uid int64) error {

	// 3. 迁移历史记录
	return svc.dao.NoteHistoryMigrate(oldNoteID, newNoteID, uid)
}
