package service

import (
	"errors"

	"github.com/haierkeys/fast-note-sync-service/internal/dao"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

type Vault struct {
	ID        int64      `json:"id"`        // ID 字段
	Vault     string     `json:"vault"`     // 保险库字段
	NoteCount int64      `json:"noteCount"` // 笔记数量字段
	Size      int64      `json:"size"`      // 大小字段
	UpdatedAt timex.Time `json:"updatedAt"` // 更新时间字段
	CreatedAt timex.Time `json:"createdAt"` // 创建时间字段
}

type VaultPostRequest struct {
	Vault string `json:"vault" form:"vault" binding:"required"`
	ID    int64  `json:"id" form:"id"`
}

type VaultGetRequest struct {
	ID int64 `form:"id" binding:"required,gte=1"`
}

// VaultCreate 创建保险库
func (svc *Service) VaultCreate(param *VaultPostRequest, uid int64) (*Vault, error) {

	vault, err := svc.dao.VaultGetByName(param.Vault, uid)

	if vault != nil {
		return nil, code.ErrorVaultExist
	}

	vault, err = svc.dao.VaultCreate(&dao.VaultSet{
		Vault: param.Vault,
	}, uid)
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(vault, &Vault{}).(*Vault), nil
}

// VaultUpdate 更新保险库信息
func (svc *Service) VaultUpdate(param *VaultPostRequest, uid int64) (*Vault, error) {
	vault, err := svc.dao.VaultUpdate(&dao.VaultSet{
		Vault: param.Vault,
	}, param.ID, uid)
	if err != nil {
		return nil, err
	}
	return convert.StructAssign(vault, &Vault{}).(*Vault), nil
}

// VaultGet 获取用户的保险库详情
func (svc *Service) VaultGet(param *VaultGetRequest, uid int64) (*Vault, error) {
	vault, err := svc.dao.VaultGet(param.ID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorVaultNotFound
		}
		return nil, code.ErrorDBQuery
	}
	return convert.StructAssign(vault, &Vault{}).(*Vault), nil
}

// VaultList 获取用户的保险库列表
func (svc *Service) VaultList(uid int64) ([]*Vault, error) {
	vaults, err := svc.dao.VaultList(uid)
	if err != nil {
		return nil, code.ErrorDBQuery
	}
	var results []*Vault
	for _, vault := range vaults {
		results = append(results, convert.StructAssign(vault, &Vault{}).(*Vault))
	}
	return results, nil
}

// VaultDelete 删除保险库
func (svc *Service) VaultDelete(param *VaultGetRequest, uid int64) error {

	// 调用数据访问层的删除方法
	return svc.dao.VaultDelete(param.ID, uid)
}

func (svc *Service) VaultGetOrCreate(name string, uid int64) (int64, error) {

	vault, err := svc.dao.VaultGetByName(name, uid)

	if vault == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		vault, err = svc.dao.VaultCreate(&dao.VaultSet{
			Vault: name,
		}, uid)
		if err != nil {
			return 0, err
		}
	}
	return vault.ID, nil

}
