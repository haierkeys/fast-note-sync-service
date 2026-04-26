// Package domain 定义领域模型和接口
package domain

import (
	"context"
	"time"
)


// SettingAction 定义配置操作类型
type SettingAction string

const (
	SettingActionCreate SettingAction = "create"
	SettingActionModify SettingAction = "modify"
	SettingActionDelete SettingAction = "delete"
)

// Setting 配置领域模型
type Setting struct {
	ID               int64
	VaultID          int64
	Action           SettingAction
	Path             string
	PathHash         string
	Content          string
	ContentHash      string
	Size             int64
	Rename           int64
	Ctime            int64
	Mtime            int64
	UpdatedTimestamp int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IsDeleted 判断配置是否已删除
func (s *Setting) IsDeleted() bool {
	return s.Action == SettingActionDelete
}

// IsCreated 判断配置是否为新建
func (s *Setting) IsCreated() bool {
	return s.Action == SettingActionCreate
}

// IsModified 判断配置是否已修改
func (s *Setting) IsModified() bool {
	return s.Action == SettingActionModify
}

// SettingRepository 配置仓储接口
type SettingRepository interface {
	// GetByPathHash 根据路径哈希获取配置
	GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*Setting, error)

	// ListByPathHash 根据路径哈希获取配置列表（处理重复记录）
	ListByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) ([]*Setting, error)

	// Create 创建配置
	Create(ctx context.Context, setting *Setting, uid int64) (*Setting, error)

	// Update 更新配置
	Update(ctx context.Context, setting *Setting, uid int64) (*Setting, error)

	// UpdateMtime 更新配置修改时间
	UpdateMtime(ctx context.Context, mtime int64, id, uid int64) error

	// UpdateActionMtime 更新配置类型并修改时间
	UpdateActionMtime(ctx context.Context, action SettingAction, mtime int64, id, uid int64) error

	// Delete 物理删除配置
	Delete(ctx context.Context, id, uid int64) error

	// DeletePhysicalByTime 根据时间物理删除已标记删除的配置
	DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error

	// DeletePhysicalByTimeAll 根据时间物理删除所有用户的已标记删除的配置
	DeletePhysicalByTimeAll(ctx context.Context, timestamp int64) error

	// List 分页获取配置列表
	List(ctx context.Context, vaultID int64, page, pageSize int, uid int64, keyword string) ([]*Setting, error)

	// ListCount 获取配置数量
	ListCount(ctx context.Context, vaultID, uid int64, keyword string) (int64, error)

	// ListByUpdatedTimestamp 根据更新时间戳获取配置列表
	ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*Setting, error)

	// DeleteByVaultID 物理删除该用户指定笔记本的所有配置
	DeleteByVaultID(ctx context.Context, vaultID, uid int64) error
}

