package dao

import (
	"strconv"

	"github.com/haierkeys/obsidian-better-sync-service/internal/model"
	"github.com/haierkeys/obsidian-better-sync-service/internal/query"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/app"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/convert"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/timex"
	"gorm.io/gorm"
)

type Note struct {
	ID          int64      `json:"id" form:"id"`                   // ID
	Vault       string     `json:"vault" form:"vault"`             // 保险库
	Action      string     `json:"action" form:"action"`           // 操作
	Path        string     `json:"path" form:"path"`               // 路径
	PathHash    string     `json:"pathHash" form:"pathHash"`       // 路径哈希
	Content     string     `json:"content" form:"content"`         // 内容
	ContentHash string     `json:"contentHash" form:"contentHash"` // 内容哈希
	Size        int64      `json:"size" form:"size"`               // 大小
	Mtime       timex.Time `json:"mtime" form:"mtime"`             // 修改时间
	CreatedAt   timex.Time `json:"createdAt" form:"createdAt"`     // 创建时间
	UpdatedAt   timex.Time `json:"updatedAt" form:"updatedAt"`     // 更新时间
}

type NoteSet struct {
	Vault       string     // 保险库名称或标识
	Action      string     // 操作类型
	Path        string     // 文件路径
	PathHash    string     // 文件路径的哈希值
	Content     string     // 文件内容
	ContentHash string     // 内容哈希
	Mtime       timex.Time // Mtime 是修改时间
	Size        int64      // 文件大小，不能为空
}

func (d *Dao) note(uid int64) *query.Query {
	return d.Use(
		func(g *gorm.DB) {
			model.AutoMigrate(g, "Note")
		}, strconv.Itoa(int(uid)),
	)
}

// NoteGetByPathHash 根据路径哈希获取笔记
func (d *Dao) NoteGetByPathHash(hash string, vault string, uid int64) (*Note, error) {
	u := d.note(uid).Note
	m, err := u.WithContext(d.ctx).Where(u.Vault.Eq(vault), u.PathHash.Eq(hash)).First()
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &Note{}).(*Note), nil
}

// NoteGetByPathHash 根据路径哈希获取笔记
func (d *Dao) NoteGetByPath(path string, vault string, uid int64) (*Note, error) {
	u := d.note(uid).Note
	m, err := u.WithContext(d.ctx).Where(u.Vault.Eq(vault), u.Path.Eq(path)).First()
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &Note{}).(*Note), nil
}

// 创建笔记
func (d *Dao) NoteCreate(params *NoteSet, uid int64) (*Note, error) {
	u := d.note(uid).Note
	m := convert.StructAssign(params, &model.Note{}).(*model.Note)
	m.CreatedAt = timex.Now()
	m.UpdatedAt = timex.Now()
	err := u.WithContext(d.ctx).Create(m)
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &Note{}).(*Note), nil
}

// 更新云存储配置
func (d *Dao) NoteUpdate(params *NoteSet, id int64, uid int64) (*Note, error) {
	u := d.note(uid).Note
	m := convert.StructAssign(params, &model.Note{}).(*model.Note)
	m.UpdatedAt = timex.Now()
	m.ID = id
	err := u.WithContext(d.ctx).Where(u.ID.Eq(id)).Save(m)

	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &Note{}).(*Note), nil
}

// NoteList 获取笔记列表
func (d *Dao) NoteList(vault string, page int, pageSize int, uid int64) ([]*Note, error) {

	u := d.note(uid).Note

	modelList, err := u.WithContext(d.ctx).Where(
		u.Vault.Eq(vault),
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

func (d *Dao) NoteListByUpdatedAt(t timex.Time, vault string, uid int64) ([]*Note, error) {

	u := d.note(uid).Note

	mList, err := u.WithContext(d.ctx).Where(
		u.Vault.Eq(vault),
		u.UpdatedAt.Gt(t),
	).Order(u.UpdatedAt).
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

func (d *Dao) NoteListByMtime(mt timex.Time, vault string, uid int64) ([]*Note, error) {

	u := d.note(uid).Note

	mList, err := u.WithContext(d.ctx).Where(
		u.Vault.Eq(vault),
		u.Mtime.Gte(mt),
	).Order(u.CreatedAt).
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
