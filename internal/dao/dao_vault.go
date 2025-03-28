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

type VaultSet struct {
	Vault     string     // 保险库名称或标识
	Action    string     // 操作类型
	NoteCount int64      // 笔记数量
	Size      int64      // 大小
}

// vault 访问器
func (d *Dao) vault(uid int64) *query.Query {
	return d.Use(
		func(g *gorm.DB) {
			model.AutoMigrate(g, "Vault")
		}, strconv.Itoa(int(uid)),
	)
}

// VaultGetByVault 根据保险库标识获取Vault记录
func (d *Dao) VaultGetByVault(vault string, uid int64) (*model.Vault, error) {
	u := d.vault(uid).Vault
	m, err := u.WithContext(d.ctx).Where(u.Vault.Eq(vault)).First()
	if err != nil {
		return nil, err
	}
	return m, nil
}

// VaultCreate 创建Vault记录
func (d *Dao) VaultCreate(params *VaultSet, uid int64) (*model.Vault, error) {
	u := d.vault(uid).Vault
	m := convert.StructAssign(params, &model.Vault{}).(*model.Vault)
	m.CreatedAt = timex.Now()
	m.UpdatedAt = timex.Now()
	err := u.WithContext(d.ctx).Create(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// VaultUpdate 更新Vault记录
func (d *Dao) VaultUpdate(params *VaultSet, id int64, uid int64) (*model.Vault, error) {
	u := d.vault(uid).Vault
	m := convert.StructAssign(params, &model.Vault{}).(*model.Vault)
	m.ID = id
	m.UpdatedAt = timex.Now()
	err := u.WithContext(d.ctx).Where(u.ID.Eq(id)).Save(m)

	if err != nil {
		return nil, err
	}
	return m, nil
}

// VaultList 获取Vault列表
func (d *Dao) VaultList(page int, pageSize int, uid int64) ([]*model.Vault, error) {
	u := d.vault(uid).Vault

	modelList, err := u.WithContext(d.ctx).
		Order(u.CreatedAt).
		Limit(pageSize).
		Offset(app.GetPageOffset(page, pageSize)).
		Find()

	if err != nil {
		return nil, err
	}

	return modelList, nil
}

// VaultListByUpdatedAt 获取Vault列表按更新时间过滤
func (d *Dao) VaultListByUpdatedAt(t timex.Time, uid int64) ([]*model.Vault, error) {
	u := d.vault(uid).Vault

	mList, err := u.WithContext(d.ctx).
		Where(u.UpdatedAt.Gt(t)).
		Order(u.UpdatedAt).
		Find()

	if err != nil {
		return nil, err
	}

	return mList, nil
}
