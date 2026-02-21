package domain

import "time"

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
	ID             int64
	UID            int64
	VaultID        int64     // 关联库 ID (0 表示所有库)
	Type           string    // full, incremental, sync
	StorageIds     string    // JSON 数组，如 "[1, 2]"
	IsEnabled      bool      // 是否启用
	CronStrategy   string    // daily, weekly, monthly, custom
	CronExpression string    // Cron 表达式
	RetentionDays  int       // 保留天数
	LastRunTime    time.Time // 上次运行时间
	NextRunTime    time.Time // 下次运行时间
	LastStatus     int       // 上次状态 (0: Idle, 1: Running, 2: Success, 3: Failed, 4: Stopped, 5: SuccessNoUpdate)
	LastMessage    string    // 上次运行结果消息
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
