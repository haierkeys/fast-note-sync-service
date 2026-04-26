// Package domain 定义领域模型和接口
package domain

import (
	"context"
	"time"
)


// NoteAction 定义笔记操作类型
type NoteAction string

const (
	NoteActionCreate NoteAction = "create"
	NoteActionModify NoteAction = "modify"
	NoteActionDelete NoteAction = "delete"
)

// Note 笔记领域模型
type Note struct {
	ID                      int64
	VaultID                 int64
	Action                  NoteAction
	Rename                  int64
	FID                     int64
	Path                    string
	PathHash                string
	Content                 string
	ContentHash             string
	ContentLastSnapshot     string
	ContentLastSnapshotHash string
	Version                 int64
	ClientName              string
	ClientType              string
	ClientVersion           string
	Size                    int64
	Ctime                   int64
	Mtime                   int64
	UpdatedTimestamp        int64
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

// CountSizeResult 统计结果
type CountSizeResult struct {
	Count int64
	Size  int64
}

// IsDeleted 判断笔记是否已删除
func (n *Note) IsDeleted() bool {
	return n.Action == NoteActionDelete
}

// IsCreated 判断笔记是否为新建
func (n *Note) IsCreated() bool {
	return n.Action == NoteActionCreate
}

// IsModified 判断笔记是否已修改
func (n *Note) IsModified() bool {
	return n.Action == NoteActionModify
}

// NoteRepository 笔记仓储接口
type NoteRepository interface {
	// GetByID 根据ID获取笔记
	GetByID(ctx context.Context, id, uid int64) (*Note, error)

	// GetByPathHash 根据路径哈希获取笔记（排除已删除）
	GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*Note, error)

	// GetByPathHashIncludeRecycle 根据路径哈希获取笔记（可选包含回收站）
	GetByPathHashIncludeRecycle(ctx context.Context, pathHash string, vaultID, uid int64, isRecycle bool) (*Note, error)

	// GetAllByPathHash 根据路径哈希获取笔记（包含所有状态）
	GetAllByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*Note, error)

	// ListByPathHash 根据路径哈希获取笔记列表（处理重复记录）
	ListByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) ([]*Note, error)

	// GetByPath 根据路径获取笔记
	GetByPath(ctx context.Context, path string, vaultID, uid int64) (*Note, error)

	// Create 创建笔记
	Create(ctx context.Context, note *Note, uid int64) (*Note, error)

	// Update 更新笔记
	Update(ctx context.Context, note *Note, uid int64) (*Note, error)

	// UpdateDelete 更新笔记为删除状态
	UpdateDelete(ctx context.Context, note *Note, uid int64) error

	// UpdateMtime 更新笔记修改时间
	UpdateMtime(ctx context.Context, mtime int64, id, uid int64) error

	// UpdateActionMtime 更新笔记类型并修改时间
	UpdateActionMtime(ctx context.Context, action NoteAction, mtime int64, id, uid int64) error

	// UpdateFID 仅更新笔记的文件夹关联 ID，不更新 updated_timestamp
	// 用于 SyncResourceFID 内部整理，避免污染增量同步时间戳
	// Only updates the folder ID (FID), does not update updated_timestamp
	// Used by SyncResourceFID to avoid polluting incremental sync timestamps
	UpdateFID(ctx context.Context, id, fid, uid int64) error

	// UpdateSnapshot 更新笔记快照
	UpdateSnapshot(ctx context.Context, snapshot, snapshotHash string, version, id, uid int64) error

	// Delete 物理删除笔记
	Delete(ctx context.Context, id, uid int64) error

	// DeletePhysicalByTime 根据时间物理删除已标记删除的笔记
	DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error

	// DeletePhysicalByTimeAll 根据时间物理删除所有用户的已标记删除的笔记
	DeletePhysicalByTimeAll(ctx context.Context, timestamp int64) error

	// List 分页获取笔记列表
	// searchMode: path(默认), content, regex
	// sortBy: mtime(默认), ctime, path
	// sortOrder: desc(默认), asc
	// paths: 逗号分隔的精确路径列表，非空时忽略 keyword 做 IN 查询
	List(ctx context.Context, vaultID int64, page, pageSize int, uid int64, keyword string, isRecycle bool, searchMode string, searchContent bool, sortBy string, sortOrder string, paths []string) ([]*Note, error)

	// ListCount 获取笔记数量
	// searchMode: path(默认), content, regex
	ListCount(ctx context.Context, vaultID, uid int64, keyword string, isRecycle bool, searchMode string, searchContent bool, paths []string) (int64, error)

	// ListByUpdatedTimestamp 根据更新时间戳获取笔记列表
	ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*Note, error)

	// ListByUpdatedTimestampPage 根据更新时间戳分页获取笔记列表
	ListByUpdatedTimestampPage(ctx context.Context, timestamp, vaultID, uid int64, offset, limit int) ([]*Note, error)

	// ListContentUnchanged 获取内容未变更的笔记列表
	ListContentUnchanged(ctx context.Context, uid int64) ([]*Note, error)

	// CountSizeSum 获取笔记数量和大小总和
	CountSizeSum(ctx context.Context, vaultID, uid int64) (*CountSizeResult, error)

	// ListByFID 根据文件夹ID获取笔记列表
	ListByFID(ctx context.Context, fid, vaultID, uid int64, page, pageSize int, sortBy, sortOrder string) ([]*Note, error)

	// ListByFIDCount 根据文件夹ID获取笔记数量
	ListByFIDCount(ctx context.Context, fid, vaultID, uid int64) (int64, error)

	// ListByFIDs 根据多个文件夹ID获取笔记列表（处理重复文件夹记录）
	ListByFIDs(ctx context.Context, fids []int64, vaultID, uid int64, page, pageSize int, sortBy, sortOrder string) ([]*Note, error)

	// ListByFIDsCount 根据多个文件夹ID获取笔记数量
	ListByFIDsCount(ctx context.Context, fids []int64, vaultID, uid int64) (int64, error)

	// ListByIDs 根据ID列表获取笔记列表
	ListByIDs(ctx context.Context, ids []int64, uid int64) ([]*Note, error)

	// ListByPathPrefix 根据路径前缀获取笔记列表
	ListByPathPrefix(ctx context.Context, pathPrefix string, vaultID, uid int64) ([]*Note, error)

	// RecycleClear 清理回收站
	RecycleClear(ctx context.Context, path, pathHash string, vaultID, uid int64) error

	// DeleteByVaultID 物理删除仓库下的所有笔记
	DeleteByVaultID(ctx context.Context, vaultID, uid int64) error
}

