package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

// storageRepository 实现 domain.StorageRepository 接口
type storageRepository struct {
	dao *Dao
}

// NewStorageRepository 创建 StorageRepository 实例
func NewStorageRepository(dao *Dao) domain.StorageRepository {
	return &storageRepository{dao: dao}
}

func (r *storageRepository) GetKey(uid int64) string {
	return "user_storage_" + fmt.Sprintf("%d", uid)
}

// storage 获取存储配置查询对象
func (r *storageRepository) storage(uid int64) *query.Query {
	return r.dao.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "Storage")
	}, r.GetKey(uid)+"#storage", r.GetKey(uid))
}

// toDomain 将数据库模型转换为领域模型
func (r *storageRepository) toDomain(m *model.Storage) *domain.Storage {
	if m == nil {
		return nil
	}
	return &domain.Storage{
		ID:              m.ID,
		UID:             m.UID,
		Type:            m.Type,
		Endpoint:        m.Endpoint,
		Region:          m.Region,
		AccountID:       m.AccountID,
		BucketName:      m.BucketName,
		AccessKeyID:     m.AccessKeyID,
		AccessKeySecret: m.AccessKeySecret,
		CustomPath:      m.CustomPath,
		AccessURLPrefix: m.AccessURLPrefix,
		User:            m.User,
		Password:        m.Password,
		IsEnabled:       m.IsEnabled == 1,
		IsDeleted:       m.IsDeleted == 1,
		CreatedAt:       time.Time(m.CreatedAt),
		UpdatedAt:       time.Time(m.UpdatedAt),
	}
}

// toModel 将领域模型转换为数据库模型
func (r *storageRepository) toModel(s *domain.Storage) *model.Storage {
	if s == nil {
		return nil
	}
	isDeleted := int64(0)
	if s.IsDeleted {
		isDeleted = 1
	}
	modelStorage := &model.Storage{
		ID:              s.ID,
		UID:             s.UID,
		Type:            s.Type,
		Endpoint:        s.Endpoint,
		Region:          s.Region,
		AccountID:       s.AccountID,
		BucketName:      s.BucketName,
		AccessKeyID:     s.AccessKeyID,
		AccessKeySecret: s.AccessKeySecret,
		CustomPath:      s.CustomPath,
		AccessURLPrefix: s.AccessURLPrefix,
		User:            s.User,
		Password:        s.Password,
		IsEnabled:       int64(0),
		IsDeleted:       isDeleted,
		CreatedAt:       timex.Time(s.CreatedAt),
		UpdatedAt:       timex.Time(s.UpdatedAt),
	}

	if s.IsEnabled {
		modelStorage.IsEnabled = 1
	}
	return modelStorage
}

// GetByID 根据ID获取存储配置
func (r *storageRepository) GetByID(ctx context.Context, id, uid int64) (*domain.Storage, error) {
	u := r.storage(uid).Storage
	m, err := u.WithContext(ctx).Where(u.ID.Eq(id), u.UID.Eq(uid), u.IsDeleted.Eq(0)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m), nil
}

// Create 创建存储配置
func (r *storageRepository) Create(ctx context.Context, storage *domain.Storage, uid int64) (*domain.Storage, error) {
	var result *domain.Storage
	var createErr error

	err := r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.storage(uid).Storage
		m := r.toModel(storage)
		m.UID = uid
		m.IsDeleted = 0
		m.CreatedAt = timex.Now()
		m.UpdatedAt = timex.Now()

		createErr = u.WithContext(ctx).Create(m)
		if createErr != nil {
			return createErr
		}
		result = r.toDomain(m)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, createErr
}

// Update 更新存储配置
func (r *storageRepository) Update(ctx context.Context, storage *domain.Storage, uid int64) (*domain.Storage, error) {
	var result *domain.Storage
	var updateErr error

	err := r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.storage(uid).Storage

		// 获取原有记录确认归属
		old, err := u.WithContext(ctx).Where(u.ID.Eq(storage.ID), u.UID.Eq(uid), u.IsDeleted.Eq(0)).First()
		if err != nil {
			return err
		}

		m := r.toModel(storage)
		m.UID = uid
		m.CreatedAt = old.CreatedAt
		m.UpdatedAt = timex.Now()

		updateErr = u.WithContext(ctx).Where(u.ID.Eq(storage.ID)).Save(m)
		if updateErr != nil {
			return updateErr
		}
		result = r.toDomain(m)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, updateErr
}

// List 获取用户的存储配置列表
func (r *storageRepository) List(ctx context.Context, uid int64) ([]*domain.Storage, error) {
	u := r.storage(uid).Storage
	modelList, err := u.WithContext(ctx).Where(u.UID.Eq(uid), u.IsDeleted.Eq(0)).Order(u.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}

	var list []*domain.Storage
	for _, m := range modelList {
		list = append(list, r.toDomain(m))
	}
	return list, nil
}

// Delete 删除存储配置（软删除）
func (r *storageRepository) Delete(ctx context.Context, id, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.storage(uid).Storage
		_, err := u.WithContext(ctx).Where(u.ID.Eq(id), u.UID.Eq(uid)).UpdateSimple(u.IsDeleted.Value(1), u.DeletedAt.Value(timex.Now()))
		return err
	})
}

// 确保 storageRepository 实现了 domain.StorageRepository 接口
var _ domain.StorageRepository = (*storageRepository)(nil)
