package dao

import (
	"strconv"

	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

type Vault struct {
	ID        int64      `json:"id" form:"id"`               // ID 字段
	Vault     string     `json:"vault" form:"vault"`         // 保险库字段
	NoteCount int64      `json:"noteCount" form:"noteCount"` // 笔记数量字段
	NoteSize  int64      `json:"noteSize" form:"noteSize"`   // 笔记总大小字段
	FileCount int64      `json:"fileCount" form:"fileCount"` // 文件数量字段
	FileSize  int64      `json:"fileSize" form:"fileSize"`   // 文件总大小字段
	Size      int64      `json:"size" form:"size"`           // 总大小字段(noteSize + fileSize)
	CreatedAt timex.Time `json:"createdAt" form:"createdAt"` // 创建时间字段
	UpdatedAt timex.Time `json:"updatedAt" form:"updatedAt"` // 更新时间字段
}

type VaultSet struct {
	Vault string // 保险库名称或标识
}

// vault 获取保险库查询对象
// 函数名: vault
// 函数使用说明: 获取指定用户的保险库表查询对象,内部方法。
// 参数说明:
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *query.Query: 查询对象
func (d *Dao) vault(uid int64) *query.Query {
	key := "note_" + strconv.FormatInt(uid, 10)
	return d.Use(
		func(g *gorm.DB) {
			model.AutoMigrate(g, "Vault")
		}, key,
	)
}

func (d *Dao) VaultAutoMigrate(uid int64) error {
	key := "note_" + strconv.FormatInt(uid, 10)
	b := d.UseKey(key)
	return model.AutoMigrate(b, "Vault")
}

func (d *Dao) Query(uid int64) *gorm.DB {
	key := "note_" + strconv.FormatInt(uid, 10)
	b := d.UseKey(key)
	return b
}

// VaultGet 根据ID获取保险库记录
// 函数名: VaultGet
// 函数使用说明: 根据保险库ID查询保险库记录。
// 参数说明:
//   - id int64: 保险库ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *Vault: 保险库数据
//   - error: 出错时返回错误
func (d *Dao) VaultGet(id int64, uid int64) (*Vault, error) {
	u := d.vault(uid).Vault
	m, err := u.WithContext(d.ctx).Where(u.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	data := convert.StructAssign(m, &Vault{}).(*Vault)
	data.Size = data.NoteSize + data.FileSize
	return data, nil
}

// VaultGetByName 根据保险库名称获取保险库记录
// 函数名: VaultGetByName
// 函数使用说明: 根据保险库名称查询保险库记录。
// 参数说明:
//   - name string: 保险库名称
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *Vault: 保险库数据
//   - error: 出错时返回错误
func (d *Dao) VaultGetByName(name string, uid int64) (*Vault, error) {
	u := d.vault(uid).Vault
	m, err := u.WithContext(d.ctx).Where(u.Vault.Eq(name)).First()
	if err != nil {
		return nil, err
	}
	data := convert.StructAssign(m, &Vault{}).(*Vault)
	data.Size = data.NoteSize + data.FileSize
	return data, nil
}

// VaultCreate 创建保险库记录
// 函数名: VaultCreate
// 函数使用说明: 在数据库中创建新的保险库记录,自动设置创建时间和更新时间。
// 参数说明:
//   - params *VaultSet: 保险库创建参数
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *Vault: 创建后的保险库数据
//   - error: 出错时返回错误
func (d *Dao) VaultCreate(params *VaultSet, uid int64) (*Vault, error) {
	u := d.vault(uid).Vault
	m := convert.StructAssign(params, &model.Vault{}).(*model.Vault)
	m.CreatedAt = timex.Now()
	m.UpdatedAt = timex.Now()
	err := u.WithContext(d.ctx).Create(m)
	if err != nil {
		return nil, err
	}
	data := convert.StructAssign(m, &Vault{}).(*Vault)
	data.Size = data.NoteSize + data.FileSize
	return data, nil
}

// VaultUpdate 更新保险库记录
// 函数名: VaultUpdate
// 函数使用说明: 根据保险库ID更新保险库记录,自动更新更新时间。
// 参数说明:
//   - params *VaultSet: 保险库更新参数
//   - id int64: 保险库ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - *Vault: 更新后的保险库数据
//   - error: 出错时返回错误
func (d *Dao) VaultUpdate(params *VaultSet, id int64, uid int64) (*Vault, error) {
	u := d.vault(uid).Vault
	m := convert.StructAssign(params, &model.Vault{}).(*model.Vault)
	m.ID = id
	m.UpdatedAt = timex.Now()
	err := u.WithContext(d.ctx).Where(u.ID.Eq(id)).Save(m)

	if err != nil {
		return nil, err
	}
	data := convert.StructAssign(m, &Vault{}).(*Vault)
	data.Size = data.NoteSize + data.FileSize
	return data, nil
}

// VaultList 获取保险库列表
// 函数名: VaultList
// 函数使用说明: 查询指定用户的所有保险库列表,按更新时间排序,最多返回100条。
// 参数说明:
//   - uid int64: 用户ID
//
// 返回值说明:
//   - []*Vault: 保险库列表
//   - error: 出错时返回错误
func (d *Dao) VaultList(uid int64) ([]*Vault, error) {
	u := d.vault(uid).Vault

	modelList, err := u.WithContext(d.ctx).
		Order(u.CreatedAt).
		Limit(100).
		Order(u.UpdatedAt).
		Find()

	if err != nil {
		return nil, err
	}
	var list []*Vault

	for _, m := range modelList {
		data := convert.StructAssign(m, &Vault{}).(*Vault)
		data.Size = data.NoteSize + data.FileSize
		list = append(list, data)
	}
	return list, nil
}

// VaultDelete 删除保险库
// 函数名: VaultDelete
// 函数使用说明: 根据保险库ID物理删除保险库记录。
// 参数说明:
//   - id int64: 保险库ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - error: 出错时返回错误
func (d *Dao) VaultDelete(id int64, uid int64) error {
	u := d.vault(uid).Vault

	_, err := u.WithContext(d.ctx).Where(
		u.ID.Eq(id),
	).Delete()
	return err
}

// VaultUpdateNoteCountSize 更新保险库的笔记数量和大小
// 函数名: VaultUpdateNoteCountSize
// 函数使用说明: 更新指定保险库的笔记数量和总大小统计信息。
// 参数说明:
//   - noteSize int64: 笔记总大小
//   - noteCount int64: 笔记数量
//   - id int64: 保险库ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - error: 出错时返回错误
func (d *Dao) VaultUpdateNoteCountSize(noteSize int64, noteCount int64, id int64, uid int64) error {
	u := d.vault(uid).Vault

	_, err := u.WithContext(d.ctx).Where(
		u.ID.Eq(id),
	).UpdateSimple(
		u.NoteSize.Value(noteSize),
		u.NoteCount.Value(noteCount),
		u.UpdatedAt.Value(timex.Now()),
	)
	return err
}

// VaultUpdateFileCountSize 更新保险库的文件数量和大小
// 函数名: VaultUpdateFileCountSize
// 函数使用说明: 更新指定保险库的文件数量和总大小统计信息。
// 参数说明:
//   - fileSize int64: 文件总大小
//   - fileCount int64: 文件数量
//   - id int64: 保险库ID
//   - uid int64: 用户ID
//
// 返回值说明:
//   - error: 出错时返回错误
func (d *Dao) VaultUpdateFileCountSize(fileSize int64, fileCount int64, id int64, uid int64) error {
	u := d.vault(uid).Vault

	_, err := u.WithContext(d.ctx).Where(
		u.ID.Eq(id),
	).UpdateSimple(
		u.FileSize.Value(fileSize),
		u.FileCount.Value(fileCount),
		u.UpdatedAt.Value(timex.Now()),
	)
	return err
}
