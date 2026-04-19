package domain

import (
	"context"
	"time"
)


// FileAction 定义文件操作类型
type FileAction string

const (
	FileActionCreate FileAction = "create"
	FileActionModify FileAction = "modify"
	FileActionDelete FileAction = "delete"
)

// File 文件领域模型
type File struct {
	ID               int64
	VaultID          int64
	Action           FileAction
	FID              int64
	Path             string
	PathHash         string
	ContentHash      string
	SavePath         string
	Rename           int64
	Size             int64
	Ctime            int64
	Mtime            int64
	UpdatedTimestamp int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IsDeleted 判断文件是否已删除
func (f *File) IsDeleted() bool {
	return f.Action == FileActionDelete
}

// IsCreated 判断文件是否为新建
func (f *File) IsCreated() bool {
	return f.Action == FileActionCreate
}

// IsModified 判断文件是否已修改
func (f *File) IsModified() bool {
	return f.Action == FileActionModify
}

// FileRepository 文件仓储接口
type FileRepository interface {
	// GetByID 根据 ID 获取文件
	GetByID(ctx context.Context, id, uid int64) (*File, error)

	// GetByPathHash 根据路径哈希获取文件
	GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*File, error)

	// ListByPathHash 根据路径哈希获取文件列表（处理重复记录）
	ListByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) ([]*File, error)

	// GetByPath 根据路径获取文件
	GetByPath(ctx context.Context, path string, vaultID, uid int64) (*File, error)

	// GetByPathLike 根据路径后缀获取文件
	GetByPathLike(ctx context.Context, path string, vaultID, uid int64) (*File, error)

	// Create 创建文件
	Create(ctx context.Context, file *File, uid int64) (*File, error)

	// Update 更新文件
	Update(ctx context.Context, file *File, uid int64) (*File, error)

	// UpdateMtime 更新文件修改时间
	UpdateMtime(ctx context.Context, mtime int64, id, uid int64) error

	// UpdateActionMtime 更新文件类型并修改时间
	UpdateActionMtime(ctx context.Context, action FileAction, mtime int64, id, uid int64) error

	// UpdateFID 仅更新文件的文件夹关联 ID，不更新 updated_timestamp
	// 用于 SyncResourceFID 内部整理，避免污染增量同步时间戳
	// Only updates the folder ID (FID), does not update updated_timestamp
	// Used by SyncResourceFID to avoid polluting incremental sync timestamps
	UpdateFID(ctx context.Context, id, fid, uid int64) error

	// Delete 物理删除文件
	Delete(ctx context.Context, id, uid int64) error

	// DeletePhysicalByTime 根据时间物理删除已标记删除的文件
	DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error

	// DeletePhysicalByTimeAll 根据时间物理删除所有用户的已标记删除的文件
	DeletePhysicalByTimeAll(ctx context.Context, timestamp int64) error

	// List 分页获取文件列表
	List(ctx context.Context, vaultID int64, page, pageSize int, uid int64, keyword string, isRecycle bool, sortBy string, sortOrder string) ([]*File, error)

	// ListCount 获取文件数量
	ListCount(ctx context.Context, vaultID, uid int64, keyword string, isRecycle bool) (int64, error)

	// ListByUpdatedTimestamp 根据更新时间戳获取文件列表
	ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*File, error)

	// ListByUpdatedTimestampPage 根据更新时间戳分页获取文件列表
	ListByUpdatedTimestampPage(ctx context.Context, timestamp, vaultID, uid int64, offset, limit int) ([]*File, error)

	// ListByMtime 根据修改时间戳获取文件列表
	ListByMtime(ctx context.Context, timestamp, vaultID, uid int64) ([]*File, error)

	// CountSizeSum 获取文件数量和大小总和
	CountSizeSum(ctx context.Context, vaultID, uid int64) (*CountSizeResult, error)

	// ListByFID 根据文件夹ID获取文件列表
	ListByFID(ctx context.Context, fid, vaultID, uid int64, page, pageSize int, sortBy, sortOrder string) ([]*File, error)

	// ListByFIDCount 根据文件夹ID获取文件数量
	ListByFIDCount(ctx context.Context, fid, vaultID, uid int64) (int64, error)

	// ListByFIDs 根据多个文件夹ID获取文件列表（处理重复文件夹记录）
	ListByFIDs(ctx context.Context, fids []int64, vaultID, uid int64, page, pageSize int, sortBy, sortOrder string) ([]*File, error)

	// ListByFIDsCount 根据多个文件夹ID获取文件数量
	ListByFIDsCount(ctx context.Context, fids []int64, vaultID, uid int64) (int64, error)

	// ListByIDs 根据ID列表获取文件列表
	ListByIDs(ctx context.Context, ids []int64, uid int64) ([]*File, error)

	// ListByPathPrefix 根据路径前缀获取文件列表
	ListByPathPrefix(ctx context.Context, pathPrefix string, vaultID, uid int64) ([]*File, error)

	// RecycleClear 清理回收站
	RecycleClear(ctx context.Context, path, pathHash string, vaultID, uid int64) error
}

