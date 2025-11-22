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
	Size      int64      `json:"size" form:"size"`           // 大小字段
	CreatedAt timex.Time `json:"createdAt" form:"createdAt"` // 创建时间字段
	UpdatedAt timex.Time `json:"updatedAt" form:"updatedAt"` // 更新时间字段
}

type VaultSet struct {
	Vault     string // 保险库名称或标识
	NoteCount int64  // 笔记数量
	Size      int64  // 大小
}

// vault 访问器
func (d *Dao) vault(uid int64) *query.Query {
	key := "note_" + strconv.FormatInt(uid, 10)
	return d.Use(
		func(g *gorm.DB) {
			model.AutoMigrate(g, "Vault")
		}, key,
	)
}

// VaultGetByVault 根据保险库标识获取Vault记录
func (d *Dao) VaultGet(id int64, uid int64) (*Vault, error) {
	u := d.vault(uid).Vault
	m, err := u.WithContext(d.ctx).Where(u.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &Vault{}).(*Vault), nil
}

// VaultGetByName 根据保险库标识获取Vault记录
func (d *Dao) VaultGetByName(name string, uid int64) (*Vault, error) {
	u := d.vault(uid).Vault
	m, err := u.WithContext(d.ctx).Where(u.Vault.Eq(name)).First()
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &Vault{}).(*Vault), nil
}

// VaultCreate 创建Vault记录
func (d *Dao) VaultCreate(params *VaultSet, uid int64) (*Vault, error) {
	u := d.vault(uid).Vault
	m := convert.StructAssign(params, &model.Vault{}).(*model.Vault)
	m.CreatedAt = timex.Now()
	m.UpdatedAt = timex.Now()
	err := u.WithContext(d.ctx).Create(m)
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &Vault{}).(*Vault), nil
}

// VaultUpdate 更新Vault记录
func (d *Dao) VaultUpdate(params *VaultSet, id int64, uid int64) (*Vault, error) {
	u := d.vault(uid).Vault
	m := convert.StructAssign(params, &model.Vault{}).(*model.Vault)
	m.ID = id
	m.UpdatedAt = timex.Now()
	err := u.WithContext(d.ctx).Where(u.ID.Eq(id)).Save(m)

	if err != nil {
		return nil, err
	}
	return convert.StructAssign(m, &Vault{}).(*Vault), nil
}

// VaultList 获取Vault列表
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
		list = append(list, convert.StructAssign(m, &Vault{}).(*Vault))
	}
	return list, nil
}

// VaultDelete 删除Vault
func (d *Dao) VaultDelete(id int64, uid int64) error {
	u := d.vault(uid).Vault

	_, err := u.WithContext(d.ctx).Where(
		u.ID.Eq(id),
	).Delete()
	return err
}

func (d *Dao) VaultUpdateNoteCountSize(size int64, count int64, id int64, uid int64) error {
	u := d.vault(uid).Vault

	_, err := u.WithContext(d.ctx).Where(
		u.ID.Eq(id),
	).UpdateSimple(
		u.Size.Value(size),
		u.NoteCount.Value(count),
		u.UpdatedAt.Value(timex.Now()),
	)
	return err
}
