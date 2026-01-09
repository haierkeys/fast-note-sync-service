// Package service 实现业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

// VaultService 定义 Vault 业务服务接口
// 提供 Vault 获取和创建的核心业务逻辑
type VaultService interface {
	// GetByName 根据名称获取 Vault
	GetByName(ctx context.Context, uid int64, name string) (*domain.Vault, error)

	// GetOrCreate 获取或创建 Vault，使用 Singleflight 合并并发请求
	GetOrCreate(ctx context.Context, uid int64, name string) (*domain.Vault, error)

	// MustGetID 获取 Vault ID，如果不存在则返回错误
	// 使用 Singleflight 合并并发请求
	MustGetID(ctx context.Context, uid int64, name string) (int64, error)

	// Create 创建 Vault
	Create(ctx context.Context, uid int64, name string) (*VaultDTO, error)

	// Update 更新 Vault
	Update(ctx context.Context, uid int64, id int64, name string) (*VaultDTO, error)

	// Get 根据 ID 获取 Vault
	Get(ctx context.Context, uid int64, id int64) (*VaultDTO, error)

	// List 获取用户的 Vault 列表
	List(ctx context.Context, uid int64) ([]*VaultDTO, error)

	// Delete 删除 Vault
	Delete(ctx context.Context, uid int64, id int64) error

	// UpdateNoteStats 更新 Vault 的笔记统计信息
	UpdateNoteStats(ctx context.Context, noteSize, noteCount, vaultID, uid int64) error

	// UpdateFileStats 更新 Vault 的文件统计信息
	UpdateFileStats(ctx context.Context, fileSize, fileCount, vaultID, uid int64) error
}

// VaultDTO Vault 数据传输对象
type VaultDTO struct {
	ID        int64  `json:"id"`
	Name      string `json:"vault"`
	NoteCount int64  `json:"noteCount"`
	NoteSize  int64  `json:"noteSize"`
	FileCount int64  `json:"fileCount"`
	FileSize  int64  `json:"fileSize"`
	Size      int64  `json:"size"`
}

// vaultService 实现 VaultService 接口
type vaultService struct {
	repo domain.VaultRepository
	sf   *singleflight.Group
}

// NewVaultService 创建 VaultService 实例
func NewVaultService(repo domain.VaultRepository) VaultService {
	return &vaultService{
		repo: repo,
		sf:   &singleflight.Group{},
	}
}

// GetByName 根据名称获取 Vault
func (s *vaultService) GetByName(ctx context.Context, uid int64, name string) (*domain.Vault, error) {
	vault, err := s.repo.GetByName(ctx, name, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorVaultNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	return vault, nil
}

// GetOrCreate 获取或创建 Vault
// 使用 Singleflight 合并并发请求，避免重复创建问题
func (s *vaultService) GetOrCreate(ctx context.Context, uid int64, name string) (*domain.Vault, error) {
	key := fmt.Sprintf("vault_get_or_create_%d_%s", uid, name)

	result, err, _ := s.sf.Do(key, func() (interface{}, error) {
		// 先尝试获取
		vault, err := s.repo.GetByName(ctx, name, uid)
		if err == nil {
			return vault, nil
		}

		// 如果不存在，则创建
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newVault := &domain.Vault{
				Name: name,
			}
			created, err := s.repo.Create(ctx, newVault, uid)
			if err != nil {
				return nil, code.ErrorDBQuery.WithDetails(err.Error())
			}
			return created, nil
		}

		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	})

	if err != nil {
		return nil, err
	}
	return result.(*domain.Vault), nil
}

// MustGetID 获取 Vault ID，如果不存在则返回错误
// 使用 Singleflight 合并并发请求
func (s *vaultService) MustGetID(ctx context.Context, uid int64, name string) (int64, error) {
	key := fmt.Sprintf("vault_must_get_id_%d_%s", uid, name)

	result, err, _ := s.sf.Do(key, func() (interface{}, error) {
		vault, err := s.repo.GetByName(ctx, name, uid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, code.ErrorVaultNotFound
			}
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}
		return vault.ID, nil
	})

	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}

// UpdateNoteStats 更新 Vault 的笔记统计信息
func (s *vaultService) UpdateNoteStats(ctx context.Context, noteSize, noteCount, vaultID, uid int64) error {
	return s.repo.UpdateNoteCountSize(ctx, noteSize, noteCount, vaultID, uid)
}

// UpdateFileStats 更新 Vault 的文件统计信息
func (s *vaultService) UpdateFileStats(ctx context.Context, fileSize, fileCount, vaultID, uid int64) error {
	return s.repo.UpdateFileCountSize(ctx, fileSize, fileCount, vaultID, uid)
}

// 确保 vaultService 实现了 VaultService 接口
var _ VaultService = (*vaultService)(nil)

// domainToDTO 将领域模型转换为 DTO
func (s *vaultService) domainToDTO(vault *domain.Vault) *VaultDTO {
	if vault == nil {
		return nil
	}
	return &VaultDTO{
		ID:        vault.ID,
		Name:      vault.Name,
		NoteCount: vault.NoteCount,
		NoteSize:  vault.NoteSize,
		FileCount: vault.FileCount,
		FileSize:  vault.FileSize,
		Size:      vault.NoteSize + vault.FileSize,
	}
}

// Create 创建 Vault
func (s *vaultService) Create(ctx context.Context, uid int64, name string) (*VaultDTO, error) {
	// 检查是否已存在
	existing, err := s.repo.GetByName(ctx, name, uid)
	if err == nil && existing != nil {
		return nil, code.ErrorVaultExist
	}

	// 创建新 Vault
	newVault := &domain.Vault{
		Name: name,
	}
	created, err := s.repo.Create(ctx, newVault, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	return s.domainToDTO(created), nil
}

// Get 根据 ID 获取 Vault
func (s *vaultService) Get(ctx context.Context, uid int64, id int64) (*VaultDTO, error) {
	vault, err := s.repo.GetByID(ctx, id, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorVaultNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	return s.domainToDTO(vault), nil
}

// List 获取用户的 Vault 列表
func (s *vaultService) List(ctx context.Context, uid int64) ([]*VaultDTO, error) {
	vaults, err := s.repo.List(ctx, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var results []*VaultDTO
	for _, vault := range vaults {
		results = append(results, s.domainToDTO(vault))
	}
	return results, nil
}

// Delete 删除 Vault
func (s *vaultService) Delete(ctx context.Context, uid int64, id int64) error {
	return s.repo.Delete(ctx, id, uid)
}

// Update 更新 Vault
func (s *vaultService) Update(ctx context.Context, uid int64, id int64, name string) (*VaultDTO, error) {
	// 获取现有 Vault
	vault, err := s.repo.GetByID(ctx, id, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorVaultNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 更新名称
	vault.Name = name
	err = s.repo.Update(ctx, vault, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 重新获取更新后的 Vault
	updated, err := s.repo.GetByID(ctx, id, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	return s.domainToDTO(updated), nil
}
