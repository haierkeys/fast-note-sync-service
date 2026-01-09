// Package service 实现业务逻辑层
package service

import (
	"context"
	"sync"
	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"golang.org/x/sync/singleflight"
)

// SettingService 定义配置业务服务接口
type SettingService interface {
	// UpdateCheck 检查配置是否需要更新
	UpdateCheck(ctx context.Context, uid int64, params *dto.SettingUpdateCheckRequest) (string, *SettingDTO, error)

	// ModifyCheck 检查配置修改（UpdateCheck 的别名）
	ModifyCheck(ctx context.Context, uid int64, params *dto.SettingUpdateCheckRequest) (string, *SettingDTO, error)

	// ModifyOrCreate 创建或修改配置
	ModifyOrCreate(ctx context.Context, uid int64, params *dto.SettingModifyOrCreateRequest, mtimeCheck bool) (bool, *SettingDTO, error)

	// Modify 修改配置（ModifyOrCreate 的别名）
	Modify(ctx context.Context, uid int64, params *dto.SettingModifyOrCreateRequest) (bool, *SettingDTO, error)

	// Delete 删除配置
	Delete(ctx context.Context, uid int64, params *dto.SettingDeleteRequest) (*SettingDTO, error)

	// ListByLastTime 获取在 lastTime 之后更新的配置
	ListByLastTime(ctx context.Context, uid int64, params *dto.SettingSyncRequest) ([]*SettingDTO, error)

	// Sync 同步配置（ListByLastTime 的别名）
	Sync(ctx context.Context, uid int64, params *dto.SettingSyncRequest) ([]*SettingDTO, error)

	// Cleanup 清理过期的软删除配置
	Cleanup(ctx context.Context, uid int64) error

	// CleanupAll 清理所有用户的过期软删除配置
	CleanupAll(ctx context.Context) error
}

// SettingDTO 配置数据传输对象
type SettingDTO struct {
	ID               int64      `json:"id" form:"id"`
	Action           string     `json:"-" form:"action"`
	Path             string     `json:"path" form:"path"`
	PathHash         string     `json:"pathHash" form:"pathHash"`
	Content          string     `json:"content" form:"content"`
	ContentHash      string     `json:"contentHash" form:"contentHash"`
	Ctime            int64      `json:"ctime" form:"ctime"`
	Mtime            int64      `json:"mtime" form:"mtime"`
	UpdatedTimestamp int64      `json:"lastTime" form:"updatedTimestamp"`
	UpdatedAt        timex.Time `json:"-"`
	CreatedAt        timex.Time `json:"-"`
}

// settingService 实现 SettingService 接口
type settingService struct {
	settingRepo  domain.SettingRepository
	vaultService VaultService
	sf           *singleflight.Group

	// 清理相关
	lastCleanupTime time.Time
	cleanupMutex    sync.Mutex
}

// NewSettingService 创建 SettingService 实例
func NewSettingService(settingRepo domain.SettingRepository, vaultSvc VaultService) SettingService {
	return &settingService{
		settingRepo:  settingRepo,
		vaultService: vaultSvc,
		sf:           &singleflight.Group{},
	}
}

// domainToDTO 将领域模型转换为 DTO
func (s *settingService) domainToDTO(setting *domain.Setting) *SettingDTO {
	if setting == nil {
		return nil
	}
	return &SettingDTO{
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

// UpdateCheck 检查配置是否需要更新
func (s *settingService) UpdateCheck(ctx context.Context, uid int64, params *dto.SettingUpdateCheckRequest) (string, *SettingDTO, error) {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return "", nil, err
	}

	setting, _ := s.settingRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if setting != nil {
		settingDTO := s.domainToDTO(setting)
		// 检查内容是否一致
		if setting.ContentHash == params.ContentHash {
			// 当用户 mtime 小于服务端 mtime 时，通知用户更新 mtime
			if params.Mtime < setting.Mtime {
				return "UpdateMtime", settingDTO, nil
			} else if params.Mtime > setting.Mtime {
				_ = s.settingRepo.UpdateMtime(ctx, params.Mtime, setting.ID, uid)
			}
			return "", settingDTO, nil
		}
		return "UpdateContent", settingDTO, nil
	}
	return "Create", nil, nil
}

// ModifyCheck 检查配置修改（UpdateCheck 的别名）
func (s *settingService) ModifyCheck(ctx context.Context, uid int64, params *dto.SettingUpdateCheckRequest) (string, *SettingDTO, error) {
	return s.UpdateCheck(ctx, uid, params)
}

// ModifyOrCreate 创建或修改配置
func (s *settingService) ModifyOrCreate(ctx context.Context, uid int64, params *dto.SettingModifyOrCreateRequest, mtimeCheck bool) (bool, *SettingDTO, error) {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return false, nil, err
	}

	setting, _ := s.settingRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if setting != nil {
		// 检查内容是否一致
		if mtimeCheck && setting.Mtime == params.Mtime && setting.ContentHash == params.ContentHash {
			return false, nil, nil
		}
		// 检查内容是否一致但修改时间不同，则只更新修改时间
		if mtimeCheck && setting.Mtime < params.Mtime && setting.ContentHash == params.ContentHash {
			err := s.settingRepo.UpdateMtime(ctx, params.Mtime, setting.ID, uid)
			if err != nil {
				return false, nil, err
			}
			setting.Mtime = params.Mtime
			return false, s.domainToDTO(setting), nil
		}

		// 设置 action
		var action domain.SettingAction
		if setting.Action == domain.SettingActionDelete {
			action = domain.SettingActionCreate
		} else {
			action = domain.SettingActionModify
		}

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
			return false, nil, err
		}

		return false, s.domainToDTO(updated), nil
	}

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
		return true, nil, err
	}

	return true, s.domainToDTO(created), nil
}

// Modify 修改配置（ModifyOrCreate 的别名）
func (s *settingService) Modify(ctx context.Context, uid int64, params *dto.SettingModifyOrCreateRequest) (bool, *SettingDTO, error) {
	return s.ModifyOrCreate(ctx, uid, params, true)
}

// Delete 删除配置
func (s *settingService) Delete(ctx context.Context, uid int64, params *dto.SettingDeleteRequest) (*SettingDTO, error) {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	setting, err := s.settingRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		return nil, err
	}

	// 更新为删除状态
	setting.Action = domain.SettingActionDelete
	setting.Content = ""
	setting.ContentHash = ""
	setting.Size = 0

	updated, err := s.settingRepo.Update(ctx, setting, uid)
	if err != nil {
		return nil, err
	}

	return s.domainToDTO(updated), nil
}

// ListByLastTime 获取在 lastTime 之后更新的配置
func (s *settingService) ListByLastTime(ctx context.Context, uid int64, params *dto.SettingSyncRequest) ([]*SettingDTO, error) {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	settings, err := s.settingRepo.ListByUpdatedTimestamp(ctx, params.LastTime, vaultID, uid)
	if err != nil {
		return nil, err
	}

	var results []*SettingDTO
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

// Sync 同步配置（ListByLastTime 的别名）
func (s *settingService) Sync(ctx context.Context, uid int64, params *dto.SettingSyncRequest) ([]*SettingDTO, error) {
	return s.ListByLastTime(ctx, uid, params)
}

// Cleanup 清理过期的软删除配置
func (s *settingService) Cleanup(ctx context.Context, uid int64) error {
	retentionTimeStr := global.Config.App.SoftDeleteRetentionTime
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

// CleanupAll 清理所有用户的过期软删除配置
func (s *settingService) CleanupAll(ctx context.Context) error {
	retentionTimeStr := global.Config.App.SoftDeleteRetentionTime
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

	s.cleanupMutex.Lock()
	defer s.cleanupMutex.Unlock()

	// 动态计算检查间隔
	var checkInterval time.Duration
	if retentionDuration < time.Hour {
		checkInterval = time.Minute
	} else {
		checkInterval = retentionDuration / 10
		if checkInterval > time.Hour {
			checkInterval = time.Hour
		}
		if checkInterval < time.Minute {
			checkInterval = time.Minute
		}
	}

	// 如果距离上次清理时间不足检查间隔，则跳过
	if time.Since(s.lastCleanupTime) < checkInterval {
		return nil
	}

	s.lastCleanupTime = time.Now()
	return nil
}

// 确保 settingService 实现了 SettingService 接口
var _ SettingService = (*settingService)(nil)
