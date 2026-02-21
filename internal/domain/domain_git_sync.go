package domain

import (
	"time"
)

// GitSyncConfig Git 仓库同步任务
type GitSyncConfig struct {
	ID           int64      `json:"id"`
	UID          int64      `json:"uid"`
	VaultID      int64      `json:"vaultId"`
	RepoURL      string     `json:"repoUrl"`
	Username     string     `json:"username"`
	Password     string     `json:"password"`
	Branch       string     `json:"branch"`
	IsEnabled    bool       `json:"isEnabled"`
	Delay        int64      `json:"delay"` // 延迟时间（秒）
	LastSyncTime *time.Time `json:"lastSyncTime"`
	LastStatus   int64      `json:"lastStatus"` // 0: 闲置, 1: 运行中, 2: 成功, 3: 失败
	LastMessage  string     `json:"lastMessage"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}
