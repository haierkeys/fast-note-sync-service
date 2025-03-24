package dao

import (
	"strconv"

	"github.com/haierkeys/obsidian-better-sync-service/internal/model"
	"github.com/haierkeys/obsidian-better-sync-service/internal/query"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/convert"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/timex"
)

type Note struct {
	ID        int64      // 唯一标识符，主键
	Vault     string     // 保险库名称或标识
	Action    string     // 操作类型
	Path      string     // 文件路径
	PathHash  string     // 文件路径的哈希值
	Content   string     // 文件内容
	Size      int64      // 文件大小，不能为空
	IsDeleted int64      // 是否已删除，不能为空
	CreatedAt timex.Time // 创建时间，自动填充创建时间
	UpdatedAt timex.Time // 更新时间，自动填充更新时间
	DeletedAt timex.Time // 删除时间，默认为NULL
}

type NoteSet struct {
	Vault    string // 保险库名称或标识
	Action   string // 操作类型
	Path     string // 文件路径
	PathHash string // 文件路径的哈希值
	Content  string // 文件内容
	Size     int64  // 文件大小，不能为空
}

func (d *Dao) note(uid int64) *query.Query {
	return d.Use(strconv.Itoa(int(uid)))
}

// 创建笔记
func (d *Dao) Create(params *NoteSet, uid int64) (*Note, error) {
	u := d.note(uid).Note
	m := convert.StructAssign(params, &model.Note{}).(*model.Note)
	err := u.WithContext(d.ctx).Create(m)
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &Note{}).(*Note), nil
}

// 更新云存储配置
func (d *Dao) Update(params *NoteSet, id int64, uid int64) error {

	u := d.note(uid).Note

	m := convert.StructAssign(params, &model.Note{}).(*model.Note)
	m.ID = id
	err := u.WithContext(d.ctx).Where(u.ID.Eq(id), u.IsDeleted.Eq(0)).Save(m)

	return err
}

// 删除笔记
func (d *Dao) Delete(id int64, uid int64) error {
	u := d.note(uid).Note
	_, err := u.WithContext(d.ctx).Where(u.ID.Eq(id), u.IsDeleted.Eq(0)).UpdateSimple(u.IsDeleted.Value(1))
	// 如果发生错误，返回 nil 和错误
	return err
}
