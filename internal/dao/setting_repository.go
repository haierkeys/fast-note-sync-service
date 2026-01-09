// Package dao 实现数据访问层
package dao

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
)

// settingRepository 实现 domain.SettingRepository 接口
type settingRepository struct {
	dao *Dao
}

// NewSettingRepository 创建 SettingRepository 实例
func NewSettingRepository(dao *Dao) domain.SettingRepository {
	return &settingRepository{dao: dao}
}

// modelToDomain 将数据库模型转换为领域模型
func (r *settingRepository) modelToDomain(m *Setting) *domain.Setting {
	if m == nil {
		return nil
	}
	return &domain.Setting{
		ID:               m.ID,
		VaultID:          m.VaultID,
		Action:           domain.SettingAction(m.Action),
		Path:             m.Path,
		PathHash:         m.PathHash,
		Content:          m.Content,
		ContentHash:      m.ContentHash,
		Size:             m.Size,
		Ctime:            m.Ctime,
		Mtime:            m.Mtime,
		UpdatedTimestamp: m.UpdatedTimestamp,
		CreatedAt:        time.Time(m.CreatedAt),
		UpdatedAt:        time.Time(m.UpdatedAt),
	}
}

// domainToModel 将领域模型转换为数据库模型
func (r *settingRepository) domainToModel(s *domain.Setting) *model.Setting {
	if s == nil {
		return nil
	}
	return &model.Setting{
		ID:               s.ID,
		VaultID:          s.VaultID,
		Action:           string(s.Action),
		Path:             s.Path,
		PathHash:         s.PathHash,
		Content:          s.Content,
		ContentHash:      s.ContentHash,
		Size:             s.Size,
		Ctime:            s.Ctime,
		Mtime:            s.Mtime,
		UpdatedTimestamp: s.UpdatedTimestamp,
		CreatedAt:        timex.Time(s.CreatedAt),
		UpdatedAt:        timex.Time(s.UpdatedAt),
	}
}

// GetByPathHash 根据路径哈希获取配置
func (r *settingRepository) GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*domain.Setting, error) {
	setting, err := r.dao.SettingGetByPathHash(pathHash, vaultID, uid)
	if err != nil {
		return nil, err
	}
	return r.modelToDomain(setting), nil
}

// Create 创建配置
func (r *settingRepository) Create(ctx context.Context, setting *domain.Setting, uid int64) (*domain.Setting, error) {
	params := &SettingSet{
		VaultID:     setting.VaultID,
		Action:      string(setting.Action),
		Path:        setting.Path,
		PathHash:    setting.PathHash,
		Content:     setting.Content,
		ContentHash: setting.ContentHash,
		Size:        setting.Size,
		Ctime:       setting.Ctime,
		Mtime:       setting.Mtime,
	}
	created, err := r.dao.SettingCreate(params, uid)
	if err != nil {
		return nil, err
	}
	return r.modelToDomain(created), nil
}

// Update 更新配置
func (r *settingRepository) Update(ctx context.Context, setting *domain.Setting, uid int64) (*domain.Setting, error) {
	params := &SettingSet{
		VaultID:     setting.VaultID,
		Action:      string(setting.Action),
		Path:        setting.Path,
		PathHash:    setting.PathHash,
		Content:     setting.Content,
		ContentHash: setting.ContentHash,
		Size:        setting.Size,
		Ctime:       setting.Ctime,
		Mtime:       setting.Mtime,
	}
	updated, err := r.dao.SettingUpdate(params, setting.ID, uid)
	if err != nil {
		return nil, err
	}
	return r.modelToDomain(updated), nil
}

// UpdateMtime 更新配置修改时间
func (r *settingRepository) UpdateMtime(ctx context.Context, mtime int64, id, uid int64) error {
	return r.dao.SettingUpdateMtime(mtime, id, uid)
}

// Delete 物理删除配置
func (r *settingRepository) Delete(ctx context.Context, id, uid int64) error {
	// 暂不实现物理删除单条记录
	return nil
}

// DeletePhysicalByTime 根据时间物理删除已标记删除的配置
func (r *settingRepository) DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error {
	return r.dao.SettingDeletePhysicalByTime(timestamp, uid)
}

// ListByUpdatedTimestamp 根据更新时间戳获取配置列表
func (r *settingRepository) ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.Setting, error) {
	settings, err := r.dao.SettingListByUpdatedTimestamp(timestamp, vaultID, uid)
	if err != nil {
		return nil, err
	}

	var results []*domain.Setting
	for _, s := range settings {
		results = append(results, r.modelToDomain(s))
	}
	return results, nil
}

// 确保 settingRepository 实现了 domain.SettingRepository 接口
var _ domain.SettingRepository = (*settingRepository)(nil)
