package domain

import (
	"context"
	"time"
)


const (
	BackupStatusIdle     = 0
	BackupStatusRunning  = 1
	BackupStatusSuccess  = 2
	BackupStatusFailed   = 3
	BackupStatusStopped  = 4
	BackupStatusNoUpdate = 5
)

// BackupConfig 备份配置领域模型
type BackupConfig struct {
	ID               int64
	UID              int64
	VaultID          int64     // 关联库 ID (0 表示所有库)
	Type             string    // full, incremental, sync
	StorageIds       string    // JSON 数组，如 "[1, 2]"
	IsEnabled        bool      // 是否启用
	CronStrategy     string    // daily, weekly, monthly, custom
	CronExpression   string    // Cron 表达式
	IncludeVaultName bool      // 同步路径是否包含仓库名前缀
	RetentionDays    int       // 保留天数
	LastRunTime      time.Time // 上次运行时间
	NextRunTime      time.Time // 下次运行时间
	LastStatus       int       // 上次状态 (0: Idle, 1: Running, 2: Success, 3: Failed, 4: Stopped, 5: SuccessNoUpdate)
	LastMessage      string    // 上次运行结果消息
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// BackupHistory 备份历史领域模型
type BackupHistory struct {
	ID        int64
	UID       int64
	ConfigID  int64
	StorageID int64
	Type      string // full, incremental, sync
	StartTime time.Time
	EndTime   time.Time
	Status    int // 0: Idle, 1: Running, 2: Success, 3: Failed, 4: Stopped, 5: SuccessNoUpdate
	FileSize  int64
	FileCount int64
	Message   string
	FilePath  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BackupRepository 备份仓储接口
type BackupRepository interface {
	// ListConfigs 获取用户的备份配置列表
	ListConfigs(ctx context.Context, uid int64) ([]*BackupConfig, error)
	// GetByID 根据ID获取备份配置
	GetByID(ctx context.Context, id, uid int64) (*BackupConfig, error)
	// DeleteConfig 删除备份配置
	DeleteConfig(ctx context.Context, id, uid int64) error
	// SaveConfig 保存备份配置
	SaveConfig(ctx context.Context, config *BackupConfig, uid int64) (*BackupConfig, error)
	// ListEnabledConfigs 获取所有已启用的备份配置
	ListEnabledConfigs(ctx context.Context) ([]*BackupConfig, error)
	// UpdateNextRunTime 更新下次执行时间
	UpdateNextRunTime(ctx context.Context, id, uid int64, nextRun time.Time) error

	// CreateHistory 创建备份历史记录
	CreateHistory(ctx context.Context, history *BackupHistory, uid int64) (*BackupHistory, error)
	// ListHistory 分页获取备份历史记录
	ListHistory(ctx context.Context, uid int64, configID int64, page, pageSize int) ([]*BackupHistory, int64, error)
	// ListOldHistory List old history records created before cutoffTime
	// 获取早于 cutoffTime 的历史记录
	ListOldHistory(ctx context.Context, uid int64, configID int64, cutoffTime time.Time) ([]*BackupHistory, error)
	// DeleteOldHistory Delete old history records created before cutoffTime
	// 删除早于 cutoffTime 的历史记录
	DeleteOldHistory(ctx context.Context, uid int64, configID int64, cutoffTime time.Time) error

	// DisableByVaultID 禁用仓库下的备份任务
	DisableByVaultID(ctx context.Context, vaultID, uid int64) error
}

