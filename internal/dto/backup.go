package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// BackupConfigRequest backup configuration request
// BackupConfigRequest 备份配置请求
type BackupConfigRequest struct {
	ID               int64  `json:"id" form:"id" example:"1"`                                                                              // ID // ID
	Vault            string `json:"vault" form:"vault" example:"test"`                                                                     // Vault name // 仓库名称
	Type             string `json:"type" form:"type" binding:"required,oneof=full incremental sync" example:"sync"`                        // Backup type // 备份类型
	StorageIds       string `json:"storageIds" form:"storageIds" binding:"required" example:"[1, 2]"`                                      // Storage IDs // 存储 ID 列表
	IsEnabled        bool   `json:"isEnabled" form:"isEnabled" example:"true"`                                                             // Is enabled // 是否启用
	CronStrategy     string `json:"cronStrategy" form:"cronStrategy" binding:"required,oneof=daily weekly monthly custom" example:"daily"` // Cron strategy // 定时策略
	CronExpression   string `json:"cronExpression" form:"cronExpression" example:"0 0 * * *"`                                              // Cron expression // Cron 表达式
	RetentionDays    int    `json:"retentionDays" form:"retentionDays" binding:"min=-1" example:"7"`                                       // Retention days // 保留天数
	IncludeVaultName bool   `json:"includeVaultName" form:"includeVaultName" example:"false"`                                              // Include vault name // 同步路径是否包含仓库名
}

// BackupExecuteRequest backup execution request
// BackupExecuteRequest 备份执行请求
type BackupExecuteRequest struct {
	ID int64 `json:"id" form:"id" example:"1"` // ID // ID
}

// BackupHistoryListRequest backup history list request
// BackupHistoryListRequest 备份历史列表请求
type BackupHistoryListRequest struct {
	ConfigID int64 `json:"configId" form:"configId" binding:"required" example:"1"` // Config ID // 配置 ID
	Page     int   `json:"page" form:"page" example:"1"`                            // Page number // 页码
	PageSize int   `json:"pageSize" form:"pageSize" example:"10"`                   // Page size // 每页大小
}

// BackupConfigDTO backup configuration DTO
// BackupConfigDTO 备份配置 DTO
type BackupConfigDTO struct {
	ID               int64      `json:"id"`               // Config ID // 配置ID
	UID              int64      `json:"uid"`              // User UID // 用户ID
	Vault            string     `json:"vault"`            // Associated vault name // 关联库名称
	Type             string     `json:"type"`             // Backup type (full, incremental, sync) // 备份类型 (full, incremental, sync)
	StorageIds       string     `json:"storageIds"`       // Storage ID list // 存储ID列表
	IsEnabled        bool       `json:"isEnabled"`        // Is enabled // 是否启用
	CronStrategy     string     `json:"cronStrategy"`     // Cron strategy // 定时策略
	CronExpression   string     `json:"cronExpression"`   // Cron expression // Cron表达式
	RetentionDays    int        `json:"retentionDays"`    // Retention days // 保留天数
	IncludeVaultName bool       `json:"includeVaultName"` // Whether sync path includes vault name // 同步路径是否包含仓库名
	LastRunTime      timex.Time `json:"lastRunTime"`      // Last run time // 上次运行时间
	NextRunTime      timex.Time `json:"nextRunTime"`      // Next run time // 下次运行时间
	LastStatus       int        `json:"lastStatus"`       // Last status (0:Idle, 1:Running, 2:Success, 3:Failed, 4:Stopped) // 上次状态 (0:Idle, 1:Running, 2:Success, 3:Failed, 4:Stopped)
	LastMessage      string     `json:"lastMessage"`      // Last run result message // 上次运行结果消息
	CreatedAt        timex.Time `json:"createdAt"`        // Created at // 创建时间
	UpdatedAt        timex.Time `json:"updatedAt"`        // Updated at // 更新时间
}

// BackupHistoryDTO backup history DTO
// BackupHistoryDTO 备份历史 DTO
type BackupHistoryDTO struct {
	ID        int64      `json:"id"`        // History record ID // 历史记录ID
	UID       int64      `json:"uid"`       // User UID // 用户ID
	ConfigID  int64      `json:"configId"`  // Config ID // 配置ID
	StorageID int64      `json:"storageId"` // Storage ID // 存储ID
	Type      string     `json:"type"`      // Backup type // 备份类型
	StartTime timex.Time `json:"startTime"` // Start time // 开始时间
	EndTime   timex.Time `json:"endTime"`   // End time // 结束时间
	Status    int        `json:"status"`    // Status (0:Idle, 1:Running, 2:Success, 3:Failed, 4:Stopped) // 状态 (0:Idle, 1:Running, 2:Success, 3:Failed, 4:Stopped)
	FileSize  int64      `json:"fileSize"`  // File size // 文件大小
	FileCount int64      `json:"fileCount"` // File count // 文件数量
	Message   string     `json:"message"`   // Result message // 结果消息
	FilePath  string     `json:"filePath"`  // File path // 文件路径
	Password  string     `json:"password"`  // Password // 密码
	CreatedAt timex.Time `json:"createdAt"` // Created at // 创建时间
	UpdatedAt timex.Time `json:"updatedAt"` // Updated at // 更新时间
}
