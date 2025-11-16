package dao

import (
	"strconv"

	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

type Note struct {
	ID               int64      `json:"id" form:"id"`                             // ID
	VaultID          int64      `json:"vaultId" form:"vaultId"`                   // 保险库ID
	Action           string     `json:"action" form:"action"`                     // 操作
	Path             string     `json:"path" form:"path"`                         // 路径
	PathHash         string     `json:"pathHash" form:"pathHash"`                 // 路径哈希
	Content          string     `json:"content" form:"content"`                   // 内容
	ContentHash      string     `json:"contentHash" form:"contentHash"`           // 内容哈希
	Size             int64      `json:"size" form:"size"`                         // 大小
	Ctime            int64      `json:"ctime" form:"ctime"`                       // 创建时间戳
	Mtime            int64      `json:"mtime" form:"mtime"`                       // 修改时间戳
	UpdatedTimestamp int64      `json:"updatedTimestamp" form:"updatedTimestamp"` // 更新时间戳
	CreatedAt        timex.Time `json:"createdAt" form:"createdAt"`               // 创建时间
	UpdatedAt        timex.Time `json:"updatedAt" form:"updatedAt"`               // 更新时间

}

type NoteSet struct {
	VaultID     int64  `json:"vaultId" form:"vaultId"`         // 保险库ID
	Action      string `json:"action" form:"action"`           // 操作
	Path        string `json:"path" form:"path"`               // 路径
	PathHash    string `json:"pathHash" form:"pathHash"`       // 路径哈希
	Content     string `json:"content" form:"content"`         // 内容
	ContentHash string `json:"contentHash" form:"contentHash"` // 内容哈希
	Size        int64  `json:"size" form:"size"`               // 大小
	Ctime       int64  `json:"ctime" form:"ctime"`             // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime"`             // 修改时间戳
}

func (d *Dao) Note(uid int64) error {
	key := "note_" + strconv.FormatInt(uid, 10)
	b := d.UseKey(key)
	return model.AutoMigrate(b, "Note")
}

func (d *Dao) note(uid int64) *query.Query {
	key := "note_" + strconv.FormatInt(uid, 10)
	return d.Use(
		func(g *gorm.DB) {
			model.AutoMigrate(g, "Note")
		}, key,
	)
}

// NoteGetByPathHash 根据路径哈希获取笔记
func (d *Dao) NoteGetByPathHash(hash string, vaultID int64, uid int64) (*Note, error) {
	u := d.note(uid).Note
	m, err := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.PathHash.Eq(hash),
	).First()
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &Note{}).(*Note), nil
}

// NoteGetByPath 根据路径获取笔记
func (d *Dao) NoteGetByPath(path string, vaultID int64, uid int64) (*Note, error) {
	u := d.note(uid).Note
	m, err := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Path.Eq(path),
	).First()
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &Note{}).(*Note), nil
}

// NoteCreate 创建笔记
func (d *Dao) NoteCreate(params *NoteSet, uid int64) (*Note, error) {
	u := d.note(uid).Note
	m := convert.StructAssign(params, &model.Note{}).(*model.Note)

	m.UpdatedTimestamp = timex.Now().UnixMilli()
	m.CreatedAt = timex.Now()
	m.UpdatedAt = timex.Now()
	err := u.WithContext(d.ctx).Create(m)
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &Note{}).(*Note), nil
}

// NoteUpdate 更新笔记
func (d *Dao) NoteUpdate(params *NoteSet, id int64, uid int64) (*Note, error) {
	u := d.note(uid).Note
	m := convert.StructAssign(params, &model.Note{}).(*model.Note)
	m.UpdatedTimestamp = timex.Now().UnixMilli()
	m.UpdatedAt = timex.Now()
	m.ID = id
	err := u.WithContext(d.ctx).Where(
		u.ID.Eq(id),
	).Save(m)

	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &Note{}).(*Note), nil
}

// NoteUpdateMtime 更新笔记的mtime
func (d *Dao) NoteUpdateMtime(mtime int64, id int64, uid int64) error {
	u := d.note(uid).Note

	_, err := u.WithContext(d.ctx).Where(
		u.ID.Eq(id),
	).UpdateSimple(u.Mtime.Value(mtime))

	return err
}

// NoteList 获取笔记列表
func (d *Dao) NoteList(vaultID int64, page int, pageSize int, uid int64) ([]*Note, error) {
	u := d.note(uid).Note
	modelList, err := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
	).Order(u.CreatedAt).
		Limit(pageSize).
		Offset(app.GetPageOffset(page, pageSize)).
		Find()

	if err != nil {
		return nil, err
	}

	var list []*Note
	for _, m := range modelList {
		list = append(list, convert.StructAssign(m, &Note{}).(*Note))
	}
	return list, nil
}

// NoteListByUpdatedTimestamp 获取笔记列表
func (d *Dao) NoteListByUpdatedTimestamp(timestamp int64, vaultID int64, uid int64) ([]*Note, error) {
	u := d.note(uid).Note
	mList, err := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.UpdatedTimestamp.Gt(timestamp),
	).Order(u.UpdatedTimestamp.Desc()).
		Find()

	if err != nil {
		return nil, err
	}

	var list []*Note
	for _, m := range mList {
		list = append(list, convert.StructAssign(m, &Note{}).(*Note))
	}
	return list, nil
}

// NoteListByMtime 获取笔记列表
func (d *Dao) NoteListByMtime(timestamp int64, vaultID int64, uid int64) ([]*Note, error) {

	u := d.note(uid).Note

	mList, err := u.WithContext(d.ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Mtime.Gt(timestamp),
	).Order(u.UpdatedTimestamp.Desc()).
		Find()

	if err != nil {
		return nil, err
	}

	var list []*Note
	for _, m := range mList {
		list = append(list, convert.StructAssign(m, &Note{}).(*Note))
	}
	return list, nil
}

type NoteCountSizeSum struct {
	Size  int64
	Count int64
}

// NoteCount 获取笔记列表
func (d *Dao) NoteCountSizeSum(vaultID int64, uid int64) (*NoteCountSizeSum, error) {
	u := d.note(uid).Note

	result := &NoteCountSizeSum{}

	err := u.WithContext(d.ctx).Select(u.Size.Sum().As("size"), u.Size.Count().As("count")).Where(
		u.VaultID.Eq(vaultID),
		u.Action.Neq("delete"),
	).Scan(result)

	if err != nil {
		return nil, err
	}
	return result, nil
}
