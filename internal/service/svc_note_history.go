package service

import (
	"fmt"

	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
)

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

type NoteHistoryListRequestParams struct {
	NoteID int64 `json:"noteId" form:"noteId" binding:"required"`
}

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

func (svc *Service) NoteHistoryRollback(uid int64, historyId int64) (*Note, error) {
	history, err := svc.dao.NoteHistoryGetById(historyId, uid)
	if err != nil {
		return nil, err
	}

	// 更新 note
	noteParams := &dao.NoteSet{
		VaultID:     history.VaultID,
		Path:        history.Path,
		Content:     history.Content,
		ContentHash: "", // 让后续逻辑自动计算或由调用方负责
		Action:      "modify",
	}

	note, err := svc.dao.NoteUpdate(noteParams, history.NoteID, uid)
	if err != nil {
		return nil, err
	}

	return convert.StructAssign(note, &Note{}).(*Note), nil
}
