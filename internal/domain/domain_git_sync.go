package domain

import (
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
