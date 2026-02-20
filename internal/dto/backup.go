package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// BackupConfigRequest 备份配置请求
type BackupConfigRequest struct {
	ID             int64  `json:"id" form:"id" example:"1"`
	Vault          string `json:"vault" form:"vault" example:"test"`
	Type           string `json:"type" form:"type" binding:"required,oneof=full incremental sync" example:"sync"`
	StorageIds     string `json:"storageIds" form:"storageIds" binding:"required" example:"[1, 2]"`
	IsEnabled      bool   `json:"isEnabled" form:"isEnabled" example:"true"`
	CronStrategy   string `json:"cronStrategy" form:"cronStrategy" binding:"required,oneof=daily weekly monthly custom" example:"daily"`
	CronExpression string `json:"cronExpression" form:"cronExpression" example:"0 0 * * *"`
	RetentionDays  int    `json:"retentionDays" form:"retentionDays" binding:"min=1" example:"7"`
}

// BackupExecuteRequest 备份执行请求
type BackupExecuteRequest struct {
	ID int64 `json:"id" form:"id" example:"1"`
}

// BackupHistoryListRequest 备份历史列表请求
type BackupHistoryListRequest struct {
	ConfigID int64 `json:"configId" form:"configId" binding:"required" example:"1"`
	Page     int   `json:"page" form:"page" example:"1"`
	PageSize int   `json:"pageSize" form:"pageSize" example:"10"`
}

// BackupConfigDTO 备份配置 DTO
type BackupConfigDTO struct {
	ID             int64      `json:"id"`             // 配置ID
	UID            int64      `json:"uid"`            // 用户ID
	Vault          string     `json:"vault"`          // 关联库名称
	Type           string     `json:"type"`           // 备份类型 (full, incremental, sync)
	StorageIds     string     `json:"storageIds"`     // 存储ID列表
	IsEnabled      bool       `json:"isEnabled"`      // 是否启用
	CronStrategy   string     `json:"cronStrategy"`   // 定时策略
	CronExpression string     `json:"cronExpression"` // Cron表达式
	RetentionDays  int        `json:"retentionDays"`  // 保留天数
	LastRunTime    timex.Time `json:"lastRunTime"`    // 上次运行时间
	NextRunTime    timex.Time `json:"nextRunTime"`    // 下次运行时间
	LastStatus     int        `json:"lastStatus"`     // 上次状态 (0:Idle, 1:Running, 2:Success, 3:Failed, 4:Stopped)
	LastMessage    string     `json:"lastMessage"`    // 上次运行结果消息
	CreatedAt      timex.Time `json:"createdAt"`      // 创建时间
	UpdatedAt      timex.Time `json:"updatedAt"`      // 更新时间
}

// BackupHistoryDTO 备份历史 DTO
type BackupHistoryDTO struct {
	ID        int64      `json:"id"`        // 历史记录ID
	UID       int64      `json:"uid"`       // 用户ID
	ConfigID  int64      `json:"configId"`  // 配置ID
	StorageID int64      `json:"storageId"` // 存储ID
	Type      string     `json:"type"`      // 备份类型
	StartTime timex.Time `json:"startTime"` // 开始时间
	EndTime   timex.Time `json:"endTime"`   // 结束时间
	Status    int        `json:"status"`    // 状态 (0:Idle, 1:Running, 2:Success, 3:Failed, 4:Stopped)
	FileSize  int64      `json:"fileSize"`  // 文件大小
	FileCount int64      `json:"fileCount"` // 文件数量
	Message   string     `json:"message"`   // 结果消息
	FilePath  string     `json:"filePath"`  // 文件路径
	CreatedAt timex.Time `json:"createdAt"` // 创建时间
	UpdatedAt timex.Time `json:"updatedAt"` // 更新时间
}
