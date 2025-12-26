package dao

import (
	"strconv"

	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"gorm.io/gorm"
)

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
	UpdatedAt  string `json:"updatedAt" form:"updatedAt"`
}

type NoteHistorySet struct {
	NoteID     int64  `json:"noteId" form:"noteId"`
	VaultID    int64  `json:"vaultId" form:"vaultId"`
	Path       string `json:"path" form:"path"`
	DiffPatch  string `json:"diffPatch" form:"diffPatch"`
	Content    string `json:"content" form:"content"`
	ClientName string `json:"clientName" form:"clientName"`
	Version    int64  `json:"version" form:"version"`
}

func (d *Dao) noteHistory(uid int64) *query.Query {
	key := "user_" + strconv.FormatInt(uid, 10)
	return d.UseQuery(key)
}

func (d *Dao) NoteHistoryCreate(params *NoteHistorySet, uid int64) (*NoteHistory, error) {
	u := d.noteHistory(uid).NoteHistory
	m := convert.StructAssign(params, &model.NoteHistory{}).(*model.NoteHistory)
	err := u.WithContext(d.ctx).Create(m)
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &NoteHistory{}).(*NoteHistory), nil
}

func (d *Dao) NoteHistoryGetById(id int64, uid int64) (*NoteHistory, error) {
	u := d.noteHistory(uid).NoteHistory
	m, err := u.WithContext(d.ctx).Where(u.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &NoteHistory{}).(*NoteHistory), nil
}

func (d *Dao) NoteHistoryListByNoteId(noteId int64, page, pageSize int, uid int64) ([]*NoteHistory, int64, error) {
	u := d.noteHistory(uid).NoteHistory
	q := u.WithContext(d.ctx).Where(u.NoteID.Eq(noteId))

	count, err := q.Count()
	if err != nil {
		return nil, 0, err
	}

	modelList, err := q.Order(u.Version.Desc()).
		Limit(pageSize).
		Offset(app.GetPageOffset(page, pageSize)).
		Find()
	if err != nil {
		return nil, 0, err
	}

	var list []*NoteHistory
	for _, m := range modelList {
		list = append(list, convert.StructAssign(m, &NoteHistory{}).(*NoteHistory))
	}
	return list, count, nil
}

func (d *Dao) NoteHistoryGetLatestVersion(noteId int64, uid int64) (int64, error) {
	u := d.noteHistory(uid).NoteHistory
	m, err := u.WithContext(d.ctx).Where(u.NoteID.Eq(noteId)).Order(u.Version.Desc()).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return m.Version, nil
}

func (d *Dao) NoteHistoryMigrate(oldNoteID, newNoteID int64, uid int64) error {
	u := d.noteHistory(uid).NoteHistory
	_, err := u.WithContext(d.ctx).Where(u.NoteID.Eq(oldNoteID)).Update(u.NoteID, newNoteID)
	return err
}
