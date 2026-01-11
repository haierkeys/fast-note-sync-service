// Package dao 实现数据访问层
package dao

import (
	"context"
	"strconv"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// settingRepository 实现 domain.SettingRepository 接口
type settingRepository struct {
	dao             *Dao
	customPrefixKey string
}

// NewSettingRepository 创建 SettingRepository 实例
func NewSettingRepository(dao *Dao) domain.SettingRepository {
	return &settingRepository{dao: dao, customPrefixKey: "user_setting_"}
}

func (r *settingRepository) GetKey(uid int64) string {
	return r.customPrefixKey + strconv.FormatInt(uid, 10)
}

// setting 获取配置查询对象
func (r *settingRepository) setting(uid int64) *query.Query {
	key := r.GetKey(uid)
	return r.dao.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "Setting")
	}, key+"#setting", key)
}

// toDomain 将数据库模型转换为领域模型
func (r *settingRepository) toDomain(m *model.Setting) *domain.Setting {
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

// fillSettingContent 填充配置内容
func (r *settingRepository) fillSettingContent(uid int64, s *domain.Setting) {
	if s == nil {
		return
	}
	folder := r.dao.GetSettingFolderPath(uid, s.ID)

	if content, exists, _ := r.dao.LoadContentFromFile(folder, "content.txt"); exists {
		s.Content = content
	} else if s.Content != "" {
		if err := r.dao.SaveContentToFile(folder, "content.txt", s.Content); err != nil {
			r.dao.Logger().Warn("lazy migration: SaveContentToFile failed for setting content",
				zap.Int64(logger.FieldUID, uid),
				zap.Int64("settingId", s.ID),
				zap.String(logger.FieldMethod, "settingRepository.fillSettingContent"),
				zap.Error(err),
			)
		}
	}
}

// GetByPathHash 根据路径哈希获取配置
func (r *settingRepository) GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*domain.Setting, error) {
	u := r.setting(uid).Setting
	m, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.PathHash.Eq(pathHash),
	).First()
	if err != nil {
		return nil, err
	}
	result := r.toDomain(m)
	r.fillSettingContent(uid, result)
	return result, nil
}

// Create 创建配置
func (r *settingRepository) Create(ctx context.Context, setting *domain.Setting, uid int64) (*domain.Setting, error) {
	var result *domain.Setting
	var createErr error

	err := r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.setting(uid).Setting
		m := &model.Setting{
			VaultID:          setting.VaultID,
			Action:           string(setting.Action),
			Path:             setting.Path,
			PathHash:         setting.PathHash,
			ContentHash:      setting.ContentHash,
			Size:             setting.Size,
			Ctime:            setting.Ctime,
			Mtime:            setting.Mtime,
			UpdatedTimestamp: timex.Now().UnixMilli(),
			CreatedAt:        timex.Now(),
			UpdatedAt:        timex.Now(),
		}

		content := setting.Content
		m.Content = "" // 不在数据库存储内容

		createErr = u.WithContext(ctx).Create(m)
		if createErr != nil {
			return createErr
		}

		// 保存到文件存储
		folderPath := r.dao.GetSettingFolderPath(uid, m.ID)
		if err := r.dao.SaveContentToFile(folderPath, "content.txt", content); err != nil {
			return err
		}

		result = r.toDomain(m)
		result.Content = content
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// Update 更新配置
func (r *settingRepository) Update(ctx context.Context, setting *domain.Setting, uid int64) (*domain.Setting, error) {
	var result *domain.Setting
	var updateErr error

	err := r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.setting(uid).Setting
		m := &model.Setting{
			ID:               setting.ID,
			VaultID:          setting.VaultID,
			Action:           string(setting.Action),
			Path:             setting.Path,
			PathHash:         setting.PathHash,
			ContentHash:      setting.ContentHash,
			Size:             setting.Size,
			Ctime:            setting.Ctime,
			Mtime:            setting.Mtime,
			UpdatedTimestamp: timex.Now().UnixMilli(),
			UpdatedAt:        timex.Now(),
		}

		content := setting.Content
		m.Content = "" // 不在数据库更新内容

		updateErr = u.WithContext(ctx).Where(u.ID.Eq(setting.ID)).Save(m)
		if updateErr != nil {
			return updateErr
		}

		// 保存到文件存储
		folderPath := r.dao.GetSettingFolderPath(uid, setting.ID)
		if err := r.dao.SaveContentToFile(folderPath, "content.txt", content); err != nil {
			return err
		}

		result = r.toDomain(m)
		result.Content = content
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateMtime 更新配置修改时间
func (r *settingRepository) UpdateMtime(ctx context.Context, mtime int64, id, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.setting(uid).Setting
		_, err := u.WithContext(ctx).Where(u.ID.Eq(id)).UpdateSimple(
			u.Mtime.Value(mtime),
			u.UpdatedTimestamp.Value(timex.Now().UnixMilli()),
			u.UpdatedAt.Value(timex.Now()),
		)
		return err
	})
}

// Delete 物理删除配置
func (r *settingRepository) Delete(ctx context.Context, id, uid int64) error {
	// 暂不实现物理删除单条记录
	return nil
}

// DeletePhysicalByTime 根据时间物理删除已标记删除的配置
func (r *settingRepository) DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.setting(uid).Setting

		// 查找待物理删除的记录，清理文件
		mList, err := u.WithContext(ctx).Where(
			u.Action.Eq("delete"),
			u.UpdatedTimestamp.Lt(timestamp),
		).Find()

		if err == nil {
			for _, m := range mList {
				folder := r.dao.GetSettingFolderPath(uid, m.ID)
				_ = r.dao.RemoveContentFolder(folder)
			}
		}

		_, err = u.WithContext(ctx).Where(
			u.Action.Eq("delete"),
			u.UpdatedTimestamp.Lt(timestamp),
		).Delete()

		return err
	})
}

// DeletePhysicalByTimeAll 根据时间物理删除所有用户的已标记删除的配置
func (r *settingRepository) DeletePhysicalByTimeAll(ctx context.Context, timestamp int64) error {
	uids, err := r.dao.GetAllUserUIDs()
	if err != nil {
		return err
	}

	for _, uid := range uids {
		if err := r.DeletePhysicalByTime(ctx, timestamp, uid); err != nil {
			continue
		}
	}
	return nil
}

// ListByUpdatedTimestamp 根据更新时间戳获取配置列表
func (r *settingRepository) ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.Setting, error) {
	u := r.setting(uid).Setting
	mList, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.UpdatedTimestamp.Gt(timestamp),
	).Order(u.UpdatedTimestamp.Desc()).Find()

	if err != nil {
		return nil, err
	}

	var results []*domain.Setting
	for _, m := range mList {
		s := r.toDomain(m)
		r.fillSettingContent(uid, s)
		results = append(results, s)
	}
	return results, nil
}

// 确保 settingRepository 实现了 domain.SettingRepository 接口
var _ domain.SettingRepository = (*settingRepository)(nil)
