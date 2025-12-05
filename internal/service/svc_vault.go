package service

import (
	"errors"
	"fmt"

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
	NoteSize  int64      `json:"noteSize"`  // 笔记总大小字段
	FileCount int64      `json:"fileCount"` // 文件数量字段
	FileSize  int64      `json:"fileSize"`  // 文件总大小字段
	Size      int64      `json:"size"`      // 总大小字段(noteSize + fileSize)
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

	vID, err, _ := svc.SF.Do(fmt.Sprintf("Vault_GetOrCreate_%d", uid), func() (any, error) {
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
	})
	if err != nil {
		return 0, err
	}

	return vID.(int64), nil

}

func (svc *Service) VaultIdGetByName(name string, uid int64) (int64, error) {

	vault, err := svc.dao.VaultGetByName(name, uid)

	if vault == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, code.ErrorVaultNotFound
	}
	return vault.ID, nil

}

func (svc *Service) VaultGetByName(name string, uid int64) (*Vault, error) {

	vault, err := svc.dao.VaultGetByName(name, uid)

	if vault == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, code.ErrorVaultNotFound
	}
	return convert.StructAssign(vault, &Vault{}).(*Vault), nil

}

// VaultMigrateAll 迁移所有用户的保险库
func (svc *Service) VaultMigrateQuery(sql string, options ...bool) error {

	isAutoMigrate := false
	if len(options) > 0 {
		isAutoMigrate = options[0]
	}

	uids, err := svc.dao.GetAllUserUIDs()
	if err != nil {
		return err
	}

	for _, uid := range uids {
		// 忽略单个用户的清理错误，继续清理下一个
		if isAutoMigrate {
			_ = svc.dao.VaultAutoMigrate(uid)
		}
		db := svc.dao.Query(uid)
		db.Exec(sql)
	}

	return nil
}
