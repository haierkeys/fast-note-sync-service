package domain

import (
	"context"
	"time"
)


const (
	GitSyncStatusIdle     = 0
	GitSyncStatusRunning  = 1
	GitSyncStatusSuccess  = 2
	GitSyncStatusFailed   = 3
	GitSyncStatusShutdown = 4
)

// GitSyncConfig Git 仓库同步任务
type GitSyncConfig struct {
	ID            int64      `json:"id"`
	UID           int64      `json:"uid"`
	VaultID       int64      `json:"vaultId"`
	RepoURL       string     `json:"repoUrl"`
	Username      string     `json:"username"`
	Password      string     `json:"password"`
	Branch        string     `json:"branch"`
	IsEnabled     bool       `json:"isEnabled"`
	Delay         int64      `json:"delay"` // 延迟时间（秒）
	RetentionDays int64      `json:"retentionDays"`
	LastSyncTime  *time.Time `json:"lastSyncTime"`
	LastStatus    int64      `json:"lastStatus"` // 0: 闲置, 1: 运行中, 2: 成功, 3: 失败, 4: 系统关闭
	LastMessage   string     `json:"lastMessage"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// GitSyncHistory Git 同步历史
type GitSyncHistory struct {
	ID        int64     `json:"id"`
	UID       int64     `json:"uid"`
	ConfigID  int64     `json:"configId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Status    int64     `json:"status"` // 0: 闲置, 1: 运行中, 2: 成功, 3: 失败, 4: 系统关闭
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// GitSyncRepository Git 同步任务仓储接口
type GitSyncRepository interface {
	// GetByID 根据ID获取 Git 同步任务
	GetByID(ctx context.Context, id, uid int64) (*GitSyncConfig, error)
	// GetByVaultID 根据 VaultID 获取 Git 同步任务
	GetByVaultID(ctx context.Context, vaultID, uid int64) (*GitSyncConfig, error)
	// Save 保存 (创建或更新) Git 同步任务
	Save(ctx context.Context, config *GitSyncConfig, uid int64) (*GitSyncConfig, error)
	// Delete 删除 Git 同步任务
	Delete(ctx context.Context, id, uid int64) error
	// List 获取用户的 Git 同步任务列表
	List(ctx context.Context, uid int64) ([]*GitSyncConfig, error)
	// ListByVaultID 根据笔记仓库ID获取关联的 Git 同步任务列表
	ListByVaultID(ctx context.Context, vaultID, uid int64) ([]*GitSyncConfig, error)
	// ListEnabled 获取所有已启用的 Git 同步任务 (跨用户)
	ListEnabled(ctx context.Context) ([]*GitSyncConfig, error)

	// CreateHistory 创建 Git 同步历史记录
	CreateHistory(ctx context.Context, history *GitSyncHistory, uid int64) (*GitSyncHistory, error)
	// ListHistory 分页获取 Git 同步历史记录
	ListHistory(ctx context.Context, uid int64, configID int64, page, pageSize int) ([]*GitSyncHistory, int64, error)
	// DeleteHistory 删除 Git 同步历史记录
	DeleteHistory(ctx context.Context, uid int64, configID int64) error
	// DeleteOldHistory 删除指定时间之前的同步历史记录
	DeleteOldHistory(ctx context.Context, uid int64, configID int64, cutoffTime time.Time) error

	// DisableByVaultID 禁用仓库下的 Git 同步任务
	DisableByVaultID(ctx context.Context, vaultID, uid int64) error
}

