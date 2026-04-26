// Package service implements the business logic layer
// Package service 实现业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

// VaultService defines the business service interface for Vault
// Provides core business logic for Vault retrieval and creation
// VaultService 定义 Vault 业务服务接口
// 提供 Vault 获取和创建的核心业务逻辑
type VaultService interface {
	// GetByName retrieves Vault by name
	// GetByName 根据名称获取 Vault
	GetByName(ctx context.Context, uid int64, name string) (*domain.Vault, error)

	// GetOrCreate retrieves or creates Vault, using Singleflight to merge concurrent requests
	// GetOrCreate 获取或创建 Vault，使用 Singleflight 合并并发请求
	GetOrCreate(ctx context.Context, uid int64, name string) (*domain.Vault, error)

	// MustGetID retrieves Vault ID, returns error if not exists
	// Uses Singleflight to merge concurrent requests
	// MustGetID 获取 Vault ID，如果不存在则返回错误
	// 使用 Singleflight 合并并发请求
	MustGetID(ctx context.Context, uid int64, name string) (int64, error)

	// Create creates Vault
	// Create 创建 Vault
	Create(ctx context.Context, uid int64, name string) (*dto.VaultDTO, error)

	// Update updates Vault
	// Update 更新 Vault
	Update(ctx context.Context, uid int64, id int64, name string) (*dto.VaultDTO, error)

	// Get retrieves Vault by ID
	// Get 根据 ID 获取 Vault
	Get(ctx context.Context, uid int64, id int64) (*dto.VaultDTO, error)

	// List retrieves Vault list for current user
	// List 获取用户的 Vault 列表
	List(ctx context.Context, uid int64) ([]*dto.VaultDTO, error)

	// Delete deletes Vault
	// Delete 删除 Vault
	Delete(ctx context.Context, uid int64, id int64) error

	// UpdateNoteStats updates note statistics for a Vault
	// UpdateNoteStats 更新 Vault 的笔记统计信息
	UpdateNoteStats(ctx context.Context, noteSize, noteCount, vaultID, uid int64) error

	// UpdateFileStats updates file statistics for a Vault
	// UpdateFileStats 更新 Vault 的文件统计信息
	UpdateFileStats(ctx context.Context, fileSize, fileCount, vaultID, uid int64) error
}

// vaultService implementation of VaultService interface
// vaultService 实现 VaultService 接口
type vaultService struct {
	repo        domain.VaultRepository
	noteRepo    domain.NoteRepository
	fileRepo    domain.FileRepository
	folderRepo  domain.FolderRepository
	logRepo     domain.SyncLogRepository
	historyRepo domain.NoteHistoryRepository
	linkRepo    domain.NoteLinkRepository
	settingRepo domain.SettingRepository
	ftsRepo     domain.NoteFTSRepository
	shareRepo   domain.UserShareRepository
	gitRepo     domain.GitSyncRepository
	backupRepo  domain.BackupRepository
	logger      *zap.Logger
	sf          *singleflight.Group
}

// NewVaultService creates VaultService instance
// NewVaultService 创建 VaultService 实例
func NewVaultService(
	repo domain.VaultRepository,
	noteRepo domain.NoteRepository,
	fileRepo domain.FileRepository,
	folderRepo domain.FolderRepository,
	logRepo domain.SyncLogRepository,
	historyRepo domain.NoteHistoryRepository,
	linkRepo domain.NoteLinkRepository,
	settingRepo domain.SettingRepository,
	ftsRepo domain.NoteFTSRepository,
	shareRepo domain.UserShareRepository,
	gitRepo domain.GitSyncRepository,
	backupRepo domain.BackupRepository,
	logger *zap.Logger,
) VaultService {
	return &vaultService{
		repo:        repo,
		noteRepo:    noteRepo,
		fileRepo:    fileRepo,
		folderRepo:  folderRepo,
		logRepo:     logRepo,
		historyRepo: historyRepo,
		linkRepo:    linkRepo,
		settingRepo: settingRepo,
		ftsRepo:     ftsRepo,
		shareRepo:   shareRepo,
		gitRepo:     gitRepo,
		backupRepo:  backupRepo,
		logger:      logger,
		sf:          &singleflight.Group{},
	}
}

// GetByName retrieves Vault by name
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

// GetOrCreate retrieves or creates Vault
// Uses Singleflight to merge concurrent requests, avoiding duplicate creation issues
// GetOrCreate 获取或创建 Vault
// 使用 Singleflight 合并并发请求，避免重复创建问题
func (s *vaultService) GetOrCreate(ctx context.Context, uid int64, name string) (*domain.Vault, error) {
	key := fmt.Sprintf("vault_get_or_create_%d_%s", uid, name)

	result, err, _ := s.sf.Do(key, func() (interface{}, error) {
		// Attempt to retrieve first
		// 先尝试获取
		vault, err := s.repo.GetByName(ctx, name, uid)
		if err == nil {
			return vault, nil
		}

		// Create if not exists
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

// MustGetID retrieves Vault ID, returns error if not exists
// Uses Singleflight to merge concurrent requests
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

// UpdateNoteStats updates note statistics for a Vault
// UpdateNoteStats 更新 Vault 的笔记统计信息
func (s *vaultService) UpdateNoteStats(ctx context.Context, noteSize, noteCount, vaultID, uid int64) error {
	err := s.repo.UpdateNoteCountSize(ctx, noteSize, noteCount, vaultID, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}
	return nil
}

// UpdateFileStats updates file statistics for a Vault
// UpdateFileStats 更新 Vault 的文件统计信息
func (s *vaultService) UpdateFileStats(ctx context.Context, fileSize, fileCount, vaultID, uid int64) error {
	err := s.repo.UpdateFileCountSize(ctx, fileSize, fileCount, vaultID, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}
	return nil
}

// Verify vaultService implements VaultService interface
// 确保 vaultService 实现了 VaultService 接口
var _ VaultService = (*vaultService)(nil)

// domainToDTO converts domain model to DTO
// domainToDTO 将领域模型转换为 DTO
func (s *vaultService) domainToDTO(vault *domain.Vault) *dto.VaultDTO {
	if vault == nil {
		return nil
	}
	return &dto.VaultDTO{
		ID:        vault.ID,
		Name:      vault.Name,
		NoteCount: vault.NoteCount,
		NoteSize:  vault.NoteSize,
		FileCount: vault.FileCount,
		FileSize:  vault.FileSize,
		Size:      vault.NoteSize + vault.FileSize,
		CreatedAt: vault.CreatedAt.Format("2006-01-02 15:04"),
		UpdatedAt: vault.UpdatedAt.Format("2006-01-02 15:04"),
	}
}

// Create creates Vault
// Create 创建 Vault
func (s *vaultService) Create(ctx context.Context, uid int64, name string) (*dto.VaultDTO, error) {
	// Check if already exists
	// 检查是否已存在
	existing, err := s.repo.GetByName(ctx, name, uid)
	if err == nil && existing != nil {
		return nil, code.ErrorVaultExist
	}

	// Create new Vault
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

// Get retrieves Vault by ID
// Get 根据 ID 获取 Vault
func (s *vaultService) Get(ctx context.Context, uid int64, id int64) (*dto.VaultDTO, error) {
	vault, err := s.repo.GetByID(ctx, id, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorVaultNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	return s.domainToDTO(vault), nil
}

// List retrieves Vault list for current user
// List 获取用户的 Vault 列表
func (s *vaultService) List(ctx context.Context, uid int64) ([]*dto.VaultDTO, error) {
	vaults, err := s.repo.List(ctx, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var results []*dto.VaultDTO
	for _, vault := range vaults {
		results = append(results, s.domainToDTO(vault))
	}
	return results, nil
}

// Delete deletes Vault and all its associated resources
// Delete 删除 Vault 及其所有关联资源
func (s *vaultService) Delete(ctx context.Context, uid int64, id int64) error {
	// 1. 清理笔记及物理内容
	if err := s.noteRepo.DeleteByVaultID(ctx, id, uid); err != nil {
		s.logger.Warn("failed to cleanup notes when deleting vault", zap.Int64("vaultID", id), zap.Error(err))
	}

	// 2. 清理文件及物理内容
	if err := s.fileRepo.DeleteByVaultID(ctx, id, uid); err != nil {
		s.logger.Warn("failed to cleanup files when deleting vault", zap.Int64("vaultID", id), zap.Error(err))
	}

	// 3. 清理文件夹记录
	if err := s.folderRepo.DeleteByVaultID(ctx, id, uid); err != nil {
		s.logger.Warn("failed to cleanup folders when deleting vault", zap.Int64("vaultID", id), zap.Error(err))
	}

	// 4. 清理同步日志
	if err := s.logRepo.DeleteByVaultID(ctx, id, uid); err != nil {
		s.logger.Warn("failed to cleanup sync logs when deleting vault", zap.Int64("vaultID", id), zap.Error(err))
	}

	// 5. 清理历史记录及物理内容
	if err := s.historyRepo.DeleteByVaultID(ctx, id, uid); err != nil {
		s.logger.Warn("failed to cleanup history when deleting vault", zap.Int64("vaultID", id), zap.Error(err))
	}

	// 6. 清理笔记链接
	if err := s.linkRepo.DeleteByVaultID(ctx, id, uid); err != nil {
		s.logger.Warn("failed to cleanup links when deleting vault", zap.Int64("vaultID", id), zap.Error(err))
	}

	// 7. 清理全文搜索索引
	if err := s.ftsRepo.DeleteByVaultID(ctx, id, uid); err != nil {
		s.logger.Warn("failed to cleanup FTS index when deleting vault", zap.Int64("vaultID", id), zap.Error(err))
	}

	// 8. 清理分享记录
	if err := s.shareRepo.DeleteByVaultID(ctx, id, uid); err != nil {
		s.logger.Warn("failed to cleanup shares when deleting vault", zap.Int64("vaultID", id), zap.Error(err))
	}

	// 9. 禁用 Git 同步
	if err := s.gitRepo.DisableByVaultID(ctx, id, uid); err != nil {
		s.logger.Warn("failed to disable git sync when deleting vault", zap.Int64("vaultID", id), zap.Error(err))
	}

	// 10. 禁用备份任务
	if err := s.backupRepo.DisableByVaultID(ctx, id, uid); err != nil {
		s.logger.Warn("failed to disable backup when deleting vault", zap.Int64("vaultID", id), zap.Error(err))
	}

	// 11. 清理配置
	if err := s.settingRepo.DeleteByVaultID(ctx, id, uid); err != nil {
		s.logger.Warn("failed to cleanup settings when deleting vault", zap.Int64("vaultID", id), zap.Error(err))
	}

	// 最后删除仓库本身
	err := s.repo.Delete(ctx, id, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}
	return nil
}

// Update updates Vault
// Update 更新 Vault
func (s *vaultService) Update(ctx context.Context, uid int64, id int64, name string) (*dto.VaultDTO, error) {
	// Get existing Vault
	// 获取现有 Vault
	vault, err := s.repo.GetByID(ctx, id, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorVaultNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Update name
	// 更新名称
	vault.Name = name
	err = s.repo.Update(ctx, vault, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Re-fetch updated Vault
	// 重新获取更新后的 Vault
	updated, err := s.repo.GetByID(ctx, id, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	return s.domainToDTO(updated), nil
}
