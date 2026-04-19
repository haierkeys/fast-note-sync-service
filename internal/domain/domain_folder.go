package domain

import (
	"context"
	"time"
)


// FolderAction 定义文件夹操作类型
type FolderAction string

const (
	FolderActionCreate FolderAction = "create"
	FolderActionModify FolderAction = "modify"
	FolderActionDelete FolderAction = "delete"
)

// Folder 文件夹领域模型
type Folder struct {
	ID               int64
	VaultID          int64
	Action           FolderAction
	Path             string
	PathHash         string
	Level            int64
	FID              int64
	Ctime            int64
	Mtime            int64
	UpdatedTimestamp int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IsDeleted 判断文件夹是否已删除
func (f *Folder) IsDeleted() bool {
	return f.Action == FolderActionDelete
}

// FolderRepository 文件夹仓储接口
type FolderRepository interface {
	// GetByID 根据ID获取文件夹
	GetByID(ctx context.Context, id, uid int64) (*Folder, error)

	// GetByPathHash 根据路径哈希获取文件夹
	GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*Folder, error)

	// GetAllByPathHash 根据路径哈希获取所有匹配的文件夹（处理重复记录）
	GetAllByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) ([]*Folder, error)

	// GetByFID 根据父级ID获取文件夹列表
	GetByFID(ctx context.Context, fid int64, vaultID, uid int64) ([]*Folder, error)

	// Create 创建文件夹
	Create(ctx context.Context, folder *Folder, uid int64) (*Folder, error)

	// Update 更新文件夹
	Update(ctx context.Context, folder *Folder, uid int64) (*Folder, error)

	// Delete 物理删除文件夹
	Delete(ctx context.Context, id, uid int64) error

	// ListByUpdatedTimestamp 根据更新时间戳获取文件夹列表
	ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*Folder, error)

	// List 获取指定仓库下的所有文件夹
	List(ctx context.Context, vaultID int64, uid int64) ([]*Folder, error)
	// ListAll 获取该用户所有的文件夹
	ListAll(ctx context.Context, uid int64) ([]*Folder, error)
}

