// Package service implements the business logic layer
// Package service 实现业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

// SettingService defines the configuration business service interface
// SettingService 定义配置业务服务接口
type SettingService interface {
	// UpdateCheck checks if configuration needs updating
	// UpdateCheck 检查配置是否需要更新
	UpdateCheck(ctx context.Context, uid int64, params *dto.SettingUpdateCheckRequest) (string, *dto.SettingDTO, error)

	// ModifyCheck checks configuration modification (alias for UpdateCheck)
	// ModifyCheck 检查配置修改（UpdateCheck 的别名）
	ModifyCheck(ctx context.Context, uid int64, params *dto.SettingUpdateCheckRequest) (string, *dto.SettingDTO, error)

	// ModifyOrCreate creates or modifies configuration
	// ModifyOrCreate 创建或修改配置
	ModifyOrCreate(ctx context.Context, uid int64, params *dto.SettingModifyOrCreateRequest, mtimeCheck bool) (bool, *dto.SettingDTO, error)

	// Modify modifies configuration (alias for ModifyOrCreate)
	// Modify 修改配置（ModifyOrCreate 的别名）
	Modify(ctx context.Context, uid int64, params *dto.SettingModifyOrCreateRequest) (bool, *dto.SettingDTO, error)

	// Delete deletes configuration
	// Delete 删除配置
	Delete(ctx context.Context, uid int64, params *dto.SettingDeleteRequest) (*dto.SettingDTO, error)

	// Get retrieves a single configuration
	// Get 获取单条配置
	Get(ctx context.Context, uid int64, params *dto.SettingGetRequest) (*dto.SettingDTO, error)

	// ListByLastTime retrieves configurations updated after lastTime
	// ListByLastTime 获取在 lastTime 之后更新的配置
	ListByLastTime(ctx context.Context, uid int64, params *dto.SettingSyncRequest) ([]*dto.SettingDTO, error)

	// CleanDuplicateSettings cleans up duplicate configuration records
	// CleanDuplicateSettings 清理重复的配置记录
	CleanDuplicateSettings(ctx context.Context, uid int64, vaultID int64) error

	// Sync synchronizes configuration (alias for ListByLastTime)
	// Sync 同步配置（ListByLastTime 的别名）
	Sync(ctx context.Context, uid int64, params *dto.SettingSyncRequest) ([]*dto.SettingDTO, error)

	// List retrieves configurations with pagination
	// List 分页获取配置列表
	List(ctx context.Context, uid int64, params *dto.SettingListRequest, pager *pkgapp.Pager) ([]*dto.SettingDTO, int64, error)

	// Rename renames a configuration
	// Rename 重命名配置
	Rename(ctx context.Context, uid int64, params *dto.SettingRenameRequest) (*dto.SettingDTO, error)

	// Cleanup cleans up expired soft-deleted configurations
	// Cleanup 清理过期的软删除配置
	Cleanup(ctx context.Context, uid int64) error

	// CleanupByTime cleans up expired soft-deleted configurations for all users by cutoff time
	// CleanupByTime 按截止时间清理所有用户的过期软删除配置
	CleanupByTime(ctx context.Context, cutoffTime int64) error

	// ClearByVault clears all settings for a specific vault of a user
	// ClearByVault 清除用户指定笔记本的所有配置
	ClearByVault(ctx context.Context, uid int64, vaultName string) error

	// WithClient sets client info
	// WithClient 设置客户端信息
	WithClient(clientType, name, version string) SettingService
}

// settingService implementation of SettingService interface
// settingService 实现 SettingService 接口
type settingService struct {
	settingRepo    domain.SettingRepository // Setting repository // 配置仓库
	vaultService   VaultService             // Vault service // 仓库服务
	syncLogService SyncLogService           // Sync log service // 同步日志服务
	sf             *singleflight.Group      // Singleflight group // 并发请求合并组
	clientType     string                   // Client type // 客户端类型
	clientName     string                   // Client name // 客户端名称
	clientVer      string                   // Client version // 客户端版本
	config         *ServiceConfig           // Service configuration // 服务配置
}

// NewSettingService creates SettingService instance
// NewSettingService 创建 SettingService 实例
func NewSettingService(settingRepo domain.SettingRepository, vaultSvc VaultService, syncLogSvc SyncLogService, config *ServiceConfig) SettingService {
	return &settingService{
		settingRepo:    settingRepo,
		vaultService:   vaultSvc,
		syncLogService: syncLogSvc,
		sf:             &singleflight.Group{},
		config:         config,
	}
}

// WithClient sets client info, returns new SettingService instance
// WithClient 设置客户端信息，返回新 SettingService 实例
func (s *settingService) WithClient(clientType, name, version string) SettingService {
	return &settingService{
		settingRepo:    s.settingRepo,
		vaultService:   s.vaultService,
		syncLogService: s.syncLogService,
		sf:             s.sf,
		clientType:     clientType,
		clientName:     name,
		clientVer:      version,
		config:         s.config,
	}
}

// domainToDTO converts domain model to DTO
// domainToDTO 将领域模型转换为 DTO
func (s *settingService) domainToDTO(setting *domain.Setting) *dto.SettingDTO {
	if setting == nil {
		return nil
	}
	return &dto.SettingDTO{
		ID:               setting.ID,
		Action:           string(setting.Action),
		Path:             setting.Path,
		PathHash:         setting.PathHash,
		Content:          setting.Content,
		ContentHash:      setting.ContentHash,
		Ctime:            setting.Ctime,
		Mtime:            setting.Mtime,
		UpdatedTimestamp: setting.UpdatedTimestamp,
		UpdatedAt:        timex.Time(setting.UpdatedAt),
		CreatedAt:        timex.Time(setting.CreatedAt),
	}
}

// UpdateCheck checks if configuration needs updating
// UpdateCheck 检查配置是否需要更新
func (s *settingService) UpdateCheck(ctx context.Context, uid int64, params *dto.SettingUpdateCheckRequest) (string, *dto.SettingDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return "", nil, err
	}

	setting, _ := s.settingRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if setting != nil {
		settingDTO := s.domainToDTO(setting)

		// Check if setting is deleted
		// 检查设置是否已删除
		if setting.Action == domain.SettingActionDelete {
			return "Create", nil, nil
		}

		// Check if content is consistent
		// 检查内容是否一致
		if setting.ContentHash == params.ContentHash {
			// Notify user to update mtime when user mtime is less than server mtime
			// 当用户 mtime 小于服务端 mtime 时，通知用户更新 mtime
			if params.Mtime < setting.Mtime {
				return "UpdateMtime", settingDTO, nil
			} else if params.Mtime > setting.Mtime {
				if err := s.settingRepo.UpdateMtime(ctx, params.Mtime, setting.ID, uid); err != nil {
					// Non-critical update failed, log warning but do not block flow
					// 非关键更新失败，记录警告日志但不阻断流程
					zap.L().Warn("UpdateMtime failed for setting",
						zap.Int64(logger.FieldUID, uid),
						zap.Int64("settingId", setting.ID),
						zap.Int64("mtime", params.Mtime),
						zap.String(logger.FieldMethod, "SettingService.UpdateCheck"),
						zap.Error(err),
					)
				}
			}
			return "", settingDTO, nil
		}
		return "UpdateContent", settingDTO, nil
	}
	return "Create", nil, nil
}

// ModifyCheck checks configuration modification (alias for UpdateCheck)
// ModifyCheck 检查配置修改（UpdateCheck 的别名）
func (s *settingService) ModifyCheck(ctx context.Context, uid int64, params *dto.SettingUpdateCheckRequest) (string, *dto.SettingDTO, error) {
	return s.UpdateCheck(ctx, uid, params)
}

// ModifyOrCreate creates or modifies configuration
// ModifyOrCreate 创建或修改配置
func (s *settingService) ModifyOrCreate(ctx context.Context, uid int64, params *dto.SettingModifyOrCreateRequest, mtimeCheck bool) (bool, *dto.SettingDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return false, nil, err
	}

	key := fmt.Sprintf("modify_or_create_%d_%d_%s", uid, vaultID, params.PathHash)
	type result struct {
		isNew bool
		dto   *dto.SettingDTO
	}

	val, err, _ := s.sf.Do(key, func() (any, error) {
		setting, _ := s.settingRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)

		if setting != nil {
			// Check if content is consistent, excluding settings marked as deleted
			// 检查内容是否一致,排除掉已被标记删除的设置
			if mtimeCheck && setting.Action != domain.SettingActionDelete && setting.Mtime == params.Mtime && setting.ContentHash == params.ContentHash {
				return &result{isNew: false, dto: s.domainToDTO(setting)}, nil
			}
			// If content is consistent but modification time is different, only update modification time
			// 检查内容是否一致但修改时间不同，则只更新修改时间
			if mtimeCheck && setting.Mtime < params.Mtime && setting.ContentHash == params.ContentHash {
				err := s.settingRepo.UpdateActionMtime(ctx, domain.SettingActionModify, params.Mtime, setting.ID, uid)
				if err != nil {
					return nil, code.ErrorDBQuery.WithDetails(err.Error())
				}
				setting.Mtime = params.Mtime
				// Log mtime-only update // 记录仅 mtime 变更日志
				if s.syncLogService != nil {
					s.syncLogService.Log(uid, vaultID, domain.SyncLogTypeSetting, domain.SyncLogActionModify, "mtime", setting.Path, setting.PathHash, s.clientType, s.clientName, s.clientVer, int64(len(setting.Content)))
				}
				return &result{isNew: false, dto: s.domainToDTO(setting)}, nil
			}

			// Set action
			// 设置 action
			var action domain.SettingAction
			if setting.Action == domain.SettingActionDelete {
				action = domain.SettingActionCreate
			} else {
				action = domain.SettingActionModify
			}

			// Update configuration
			// 更新配置
			setting.VaultID = vaultID
			setting.Path = params.Path
			setting.PathHash = params.PathHash
			setting.Content = params.Content
			setting.ContentHash = params.ContentHash
			setting.Size = int64(len(params.Content))
			setting.Mtime = params.Mtime
			setting.Ctime = params.Ctime
			setting.Action = action

			updated, err := s.settingRepo.Update(ctx, setting, uid)
			if err != nil {
				return nil, code.ErrorDBQuery.WithDetails(err.Error())
			}

			// Log content modify // 记录内容变更日志
			if s.syncLogService != nil {
				s.syncLogService.Log(uid, vaultID, domain.SyncLogTypeSetting, domain.SyncLogActionModify, "content,mtime", updated.Path, updated.PathHash, s.clientType, s.clientName, s.clientVer, updated.Size)
			}

			return &result{isNew: false, dto: s.domainToDTO(updated)}, nil
		}

		// Create new configuration
		// 创建新配置
		newSetting := &domain.Setting{
			VaultID:     vaultID,
			Path:        params.Path,
			PathHash:    params.PathHash,
			Content:     params.Content,
			ContentHash: params.ContentHash,
			Size:        int64(len(params.Content)),
			Mtime:       params.Mtime,
			Ctime:       params.Ctime,
			Action:      domain.SettingActionCreate,
		}

		created, err := s.settingRepo.Create(ctx, newSetting, uid)
		if err != nil {
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}

		// Log create // 记录新建日志
		if s.syncLogService != nil {
			s.syncLogService.Log(uid, vaultID, domain.SyncLogTypeSetting, domain.SyncLogActionCreate, "", created.Path, created.PathHash, s.clientType, s.clientName, s.clientVer, created.Size)
		}

		return &result{isNew: true, dto: s.domainToDTO(created)}, nil
	})

	if err != nil {
		return false, nil, err
	}

	res := val.(*result)
	return res.isNew, res.dto, nil
}

// Modify modifies configuration (alias for ModifyOrCreate)
// Modify 修改配置（ModifyOrCreate 的别名）
func (s *settingService) Modify(ctx context.Context, uid int64, params *dto.SettingModifyOrCreateRequest) (bool, *dto.SettingDTO, error) {
	return s.ModifyOrCreate(ctx, uid, params, true)
}

// Delete deletes configuration
// Delete 删除配置
func (s *settingService) Delete(ctx context.Context, uid int64, params *dto.SettingDeleteRequest) (*dto.SettingDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	setting, err := s.settingRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorSettingNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Update to deleted status
	// 更新为删除状态
	setting.Action = domain.SettingActionDelete
	setting.Content = ""
	setting.ContentHash = ""
	setting.Size = 0

	updated, err := s.settingRepo.Update(ctx, setting, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Log soft delete // 记录软删除日志
	if s.syncLogService != nil {
		s.syncLogService.Log(uid, vaultID, domain.SyncLogTypeSetting, domain.SyncLogActionSoftDelete, "", setting.Path, setting.PathHash, s.clientType, s.clientName, s.clientVer, 0)
	}

	return s.domainToDTO(updated), nil
}

// Get retrieves a single configuration
// Get 获取单条配置
func (s *settingService) Get(ctx context.Context, uid int64, params *dto.SettingGetRequest) (*dto.SettingDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	setting, err := s.settingRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorSettingNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	return s.domainToDTO(setting), nil
}

// ListByLastTime retrieves configurations updated after lastTime
// ListByLastTime 获取在 lastTime 之后更新的配置
func (s *settingService) ListByLastTime(ctx context.Context, uid int64, params *dto.SettingSyncRequest) ([]*dto.SettingDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	settings, err := s.settingRepo.ListByUpdatedTimestamp(ctx, params.LastTime, vaultID, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var results []*dto.SettingDTO
	cacheList := make(map[string]bool)
	for _, setting := range settings {
		if cacheList[setting.PathHash] {
			continue
		}
		results = append(results, s.domainToDTO(setting))
		cacheList[setting.PathHash] = true
	}

	return results, nil
}

// Sync synchronizes configuration (alias for ListByLastTime)
// Sync 同步配置（ListByLastTime 的别名）
func (s *settingService) Sync(ctx context.Context, uid int64, params *dto.SettingSyncRequest) ([]*dto.SettingDTO, error) {
	return s.ListByLastTime(ctx, uid, params)
}

// List retrieves configurations with pagination
// List 分页获取配置列表
func (s *settingService) List(ctx context.Context, uid int64, params *dto.SettingListRequest, pager *pkgapp.Pager) ([]*dto.SettingDTO, int64, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.settingRepo.ListCount(ctx, vaultID, uid, params.Keyword)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}

	settings, err := s.settingRepo.List(ctx, vaultID, pager.Page, pager.PageSize, uid, params.Keyword)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var results []*dto.SettingDTO
	for _, setting := range settings {
		results = append(results, s.domainToDTO(setting))
	}

	return results, total, nil
}

func (s *settingService) Rename(ctx context.Context, uid int64, params *dto.SettingRenameRequest) (*dto.SettingDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	// 1. Find the old setting
	n, err := s.settingRepo.GetByPathHash(ctx, params.OldPathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorSettingNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 2. Check if new path already exists
	existSetting, _ := s.settingRepo.GetByPathHash(ctx, params.NewPathHash, vaultID, uid)
	if existSetting != nil && existSetting.Action != domain.SettingActionDelete {
		return nil, code.ErrorSettingExist
	}

	// 3. Mark old setting as deleted with rename flag
	n.Action = domain.SettingActionDelete
	n.Rename = 1
	n.UpdatedTimestamp = timex.Now().UnixMilli()
	_, err = s.settingRepo.Update(ctx, n, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 4. Create new or reuse setting record
	var newSettingCreated *domain.Setting
	if existSetting != nil {
		existSetting.Action = domain.SettingActionModify
		existSetting.Path = params.NewPath
		existSetting.PathHash = params.NewPathHash
		existSetting.Content = n.Content
		existSetting.ContentHash = n.ContentHash
		existSetting.Size = n.Size
		existSetting.Ctime = n.Ctime
		existSetting.Mtime = n.Mtime // Preserve original mtime
		existSetting.Rename = 0
		existSetting.UpdatedTimestamp = timex.Now().UnixMilli()
		newSettingCreated, err = s.settingRepo.Update(ctx, existSetting, uid)
	} else {
		newSetting := &domain.Setting{
			VaultID:          vaultID,
			Action:           domain.SettingActionCreate,
			Path:             params.NewPath,
			PathHash:         params.NewPathHash,
			Content:          n.Content,
			ContentHash:      n.ContentHash,
			Size:             n.Size,
			Ctime:            n.Ctime,
			Mtime:            n.Mtime, // Preserve original mtime
			Rename:           0,
			UpdatedTimestamp: timex.Now().UnixMilli(),
		}
		newSettingCreated, err = s.settingRepo.Create(ctx, newSetting, uid)
	}

	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Log rename // 记录重命名日志
	if s.syncLogService != nil {
		s.syncLogService.Log(uid, vaultID, domain.SyncLogTypeSetting, domain.SyncLogActionRename, "path", newSettingCreated.Path, newSettingCreated.PathHash, s.clientType, s.clientName, s.clientVer, newSettingCreated.Size)
	}

	return s.domainToDTO(newSettingCreated), nil
}

// Cleanup cleans up expired soft-deleted configurations
// Cleanup 清理过期的软删除配置
func (s *settingService) Cleanup(ctx context.Context, uid int64) error {
	if s.config == nil {
		return nil
	}
	retentionTimeStr := s.config.App.SoftDeleteRetentionTime
	if retentionTimeStr == "" || retentionTimeStr == "0" {
		return nil
	}

	retentionDuration, err := util.ParseDuration(retentionTimeStr)
	if err != nil {
		return err
	}

	if retentionDuration <= 0 {
		return nil
	}

	cutoffTime := time.Now().Add(-retentionDuration).UnixMilli()
	return s.settingRepo.DeletePhysicalByTime(ctx, cutoffTime, uid)
}

// CleanupByTime cleans up expired soft-deleted configurations for all users by cutoff time
// CleanupByTime 按截止时间清理所有用户的过期软删除配置
func (s *settingService) CleanupByTime(ctx context.Context, cutoffTime int64) error {
	return s.settingRepo.DeletePhysicalByTimeAll(ctx, cutoffTime)
}

// ClearByVault clears all settings for a specific vault of a user
// ClearByVault 清除用户指定笔记本的所有配置
func (s *settingService) ClearByVault(ctx context.Context, uid int64, vaultName string) error {
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, vaultName)
	if err != nil {
		return err
	}
	return s.settingRepo.DeleteByVault(ctx, vaultID, uid)
}

// CleanDuplicateSettings cleans up duplicate configuration records
// CleanDuplicateSettings 清理重复的配置记录
func (s *settingService) CleanDuplicateSettings(ctx context.Context, uid int64, vaultID int64) error {
	// Get all configurations (including deleted ones for global deduplication)
	// 获取所有配置（包含已删除，以便全局去重）
	settings, err := s.settingRepo.ListByUpdatedTimestamp(ctx, 0, vaultID, uid)
	if err != nil {
		return err
	}

	// Group by PathHash // 按 PathHash 分组
	grouped := make(map[string][]*domain.Setting)
	for _, s_ := range settings {
		grouped[s_.PathHash] = append(grouped[s_.PathHash], s_)
	}

	for pathHash, list := range grouped {
		if len(list) <= 1 {
			continue
		}

		// Retention rules:
		// 1. Prioritize retaining records with Action != delete
		// 2. If multiple active records exist, keep the one with the largest (latest) UpdatedTimestamp
		// 3. If timestamps are identical, keep the record with the largest ID
		// 保留规则：
		// 1. 优先保留 Action != delete 的记录
		// 2. 如果有多个活跃记录，保留 UpdatedTimestamp 最大（最新）的一条
		// 3. 如果时间戳一致，保留 ID 最大的记录

		var bestSetting *domain.Setting
		for _, s_ := range list {
			if bestSetting == nil {
				bestSetting = s_
				continue
			}

			// 比较逻辑
			isBetter := false
			if s_.Action != domain.SettingActionDelete && bestSetting.Action == domain.SettingActionDelete {
				isBetter = true
			} else if s_.Action == bestSetting.Action {
				if s_.UpdatedTimestamp > bestSetting.UpdatedTimestamp {
					isBetter = true
				} else if s_.UpdatedTimestamp == bestSetting.UpdatedTimestamp && s_.ID > bestSetting.ID {
					isBetter = true
				}
			}

			if isBetter {
				bestSetting = s_
			}
		}

		// Delete all records except the bestSetting
		// 删除非 bestSetting 的所有记录
		for _, s_ := range list {
			if s_.ID != bestSetting.ID {
				// Clear singleflight cache to prevent residue
				// 清除 singleflight 缓存，防止残留
				s.sf.Forget(fmt.Sprintf("modify_or_create_%d_%d_%s", uid, vaultID, pathHash))
				
				_ = s.settingRepo.Delete(ctx, s_.ID, uid)
			}
		}
	}

	return nil
}

// Verify settingService implements SettingService interface
// 确保 settingService 实现了 SettingService 接口
var _ SettingService = (*settingService)(nil)
