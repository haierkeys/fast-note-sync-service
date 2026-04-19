package domain

import (
	"context"
	"time"
)


// Storage 存储配置领域模型
type Storage struct {
	ID              int64
	UID             int64
	Type            string
	Endpoint        string
	Region          string
	AccountID       string
	BucketName      string
	AccessKeyID     string
	AccessKeySecret string
	CustomPath      string
	AccessURLPrefix string
	User            string
	Password        string
	IsEnabled       bool
	IsDeleted       bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// StorageRepository 存储仓储接口
type StorageRepository interface {
	// GetByID 根据ID获取存储配置
	GetByID(ctx context.Context, id, uid int64) (*Storage, error)

	// Create 创建存储配置
	Create(ctx context.Context, storage *Storage, uid int64) (*Storage, error)

	// Update 更新存储配置
	Update(ctx context.Context, storage *Storage, uid int64) (*Storage, error)

	// List 获取用户的存储配置列表
	List(ctx context.Context, uid int64) ([]*Storage, error)

	// Delete 删除存储配置（软删除）
	Delete(ctx context.Context, id, uid int64) error
}

