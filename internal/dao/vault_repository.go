// Package dao 实现数据访问层
package dao

import (
	"context"
	"strconv"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

// vaultRepository 实现 domain.VaultRepository 接口
type vaultRepository struct {
	dao *Dao
}

// NewVaultRepository 创建 VaultRepository 实例
func NewVaultRepository(dao *Dao) domain.VaultRepository {
	return &vaultRepository{dao: dao}
}

// vault 获取保险库查询对象
func (r *vaultRepository) vault(uid int64) *query.Query {
	key := "user_vault_" + strconv.FormatInt(uid, 10)
	return r.dao.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "Vault")
	}, key+"#vault", key)
}

// toDomain 将数据库模型转换为领域模型
func (r *vaultRepository) toDomain(m *model.Vault) *domain.Vault {
	if m == nil {
		return nil
	}
	return &domain.Vault{
		ID:        m.ID,
		UID:       0, // 模型中没有 UID 字段，由上下文提供
		Name:      m.Vault,
		NoteCount: m.NoteCount,
		NoteSize:  m.NoteSize,
		FileCount: m.FileCount,
		FileSize:  m.FileSize,
		IsDeleted: m.IsDeleted == 1,
		CreatedAt: time.Time(m.CreatedAt),
		UpdatedAt: time.Time(m.UpdatedAt),
	}
}

// toModel 将领域模型转换为数据库模型
func (r *vaultRepository) toModel(vault *domain.Vault) *model.Vault {
	if vault == nil {
		return nil
	}
	isDeleted := int64(0)
	if vault.IsDeleted {
		isDeleted = 1
	}
	return &model.Vault{
		ID:        vault.ID,
		Vault:     vault.Name,
		NoteCount: vault.NoteCount,
		NoteSize:  vault.NoteSize,
		FileCount: vault.FileCount,
		FileSize:  vault.FileSize,
		IsDeleted: isDeleted,
		CreatedAt: timex.Time(vault.CreatedAt),
		UpdatedAt: timex.Time(vault.UpdatedAt),
	}
}

// GetByID 根据ID获取仓库
func (r *vaultRepository) GetByID(ctx context.Context, id, uid int64) (*domain.Vault, error) {
	u := r.vault(uid).Vault
	m, err := u.WithContext(ctx).Where(u.ID.Eq(id), u.IsDeleted.Eq(0)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

// GetByName 根据名称获取仓库
func (r *vaultRepository) GetByName(ctx context.Context, name string, uid int64) (*domain.Vault, error) {
	u := r.vault(uid).Vault
	m, err := u.WithContext(ctx).Where(u.Vault.Eq(name), u.IsDeleted.Eq(0)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

// Create 创建仓库
func (r *vaultRepository) Create(ctx context.Context, vault *domain.Vault, uid int64) (*domain.Vault, error) {
	u := r.vault(uid).Vault
	m := &model.Vault{
		Vault:     vault.Name,
		NoteCount: vault.NoteCount,
		NoteSize:  vault.NoteSize,
		FileCount: vault.FileCount,
		FileSize:  vault.FileSize,
		IsDeleted: 0,
		CreatedAt: timex.Now(),
		UpdatedAt: timex.Now(),
	}

	err := u.WithContext(ctx).Create(m)
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

// Update 更新仓库
func (r *vaultRepository) Update(ctx context.Context, vault *domain.Vault, uid int64) error {
	u := r.vault(uid).Vault
	m := r.toModel(vault)
	m.UpdatedAt = timex.Now()

	return u.WithContext(ctx).Where(u.ID.Eq(vault.ID)).Save(m)
}

// UpdateNoteCountSize 更新仓库的笔记数量和大小
func (r *vaultRepository) UpdateNoteCountSize(ctx context.Context, noteSize, noteCount, vaultID, uid int64) error {
	u := r.vault(uid).Vault

	_, err := u.WithContext(ctx).Where(
		u.ID.Eq(vaultID),
	).UpdateSimple(
		u.NoteSize.Value(noteSize),
		u.NoteCount.Value(noteCount),
		u.UpdatedAt.Value(timex.Now()),
	)
	return err
}

// UpdateFileCountSize 更新仓库的文件数量和大小
func (r *vaultRepository) UpdateFileCountSize(ctx context.Context, fileSize, fileCount, vaultID, uid int64) error {
	u := r.vault(uid).Vault

	_, err := u.WithContext(ctx).Where(
		u.ID.Eq(vaultID),
	).UpdateSimple(
		u.FileSize.Value(fileSize),
		u.FileCount.Value(fileCount),
		u.UpdatedAt.Value(timex.Now()),
	)
	return err
}

// List 获取仓库列表
func (r *vaultRepository) List(ctx context.Context, uid int64) ([]*domain.Vault, error) {
	u := r.vault(uid).Vault

	modelList, err := u.WithContext(ctx).
		Where(u.IsDeleted.Eq(0)).
		Order(u.CreatedAt).
		Limit(100).
		Order(u.UpdatedAt).
		Find()

	if err != nil {
		return nil, err
	}

	var list []*domain.Vault
	for _, m := range modelList {
		list = append(list, r.toDomain(m))
	}
	return list, nil
}

// Delete 删除仓库（软删除）
func (r *vaultRepository) Delete(ctx context.Context, id, uid int64) error {
	u := r.vault(uid).Vault

	_, err := u.WithContext(ctx).Where(
		u.ID.Eq(id),
	).UpdateSimple(
		u.IsDeleted.Value(1),
	)
	return err
}

// 确保 vaultRepository 实现了 domain.VaultRepository 接口
var _ domain.VaultRepository = (*vaultRepository)(nil)
