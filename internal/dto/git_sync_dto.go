package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// GitSyncConfigRequest git repository sync task creation/update request
// GitSyncConfigRequest git 仓库同步任务创建/更新请求
type GitSyncConfigRequest struct {
	ID            int64  `json:"id" form:"id"`
	Vault         string `json:"vault" form:"vault"` // Associated vault name // 关联笔记本名称
	RepoURL       string `json:"repoUrl" form:"repoUrl" binding:"required"`
	Username      string `json:"username" form:"username"`
	Password      string `json:"password" form:"password"`
	Branch        string `json:"branch" form:"branch"`
	IsEnabled     bool   `json:"isEnabled" form:"isEnabled"`
	Delay         int64  `json:"delay" form:"delay"` // Delay time (seconds) // 延迟时间（秒）
	RetentionDays int64  `json:"retentionDays" form:"retentionDays"`
}

// GitSyncValidateRequest git repository sync task parameter validation request
// GitSyncValidateRequest git 仓库同步任务参数验证请求
type GitSyncValidateRequest struct {
	RepoURL  string `json:"repoUrl" form:"repoUrl" binding:"required"`
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
	Branch   string `json:"branch" form:"branch"`
}

// GitSyncExecuteRequest manually execute git repository sync task request
// GitSyncExecuteRequest 手动执行 git 仓库同步任务请求
type GitSyncExecuteRequest struct {
	ID int64 `json:"id" form:"id" binding:"required"`
}

// GitSyncCleanRequest cleanup git repository sync task workspace request
// GitSyncCleanRequest 清理 git 仓库同步任务工作区请求
type GitSyncCleanRequest struct {
	ConfigID int64 `json:"configId" form:"configId"`
}

// GitSyncDeleteRequest delete git repository sync task request
// GitSyncDeleteRequest 删除 git 仓库同步任务请求
type GitSyncDeleteRequest struct {
	ID int64 `json:"id" form:"id" binding:"required"`
}

// GitSyncConfigDTO git repository sync task DTO
// GitSyncConfigDTO git 仓库同步任务 DTO
type GitSyncConfigDTO struct {
	ID            int64      `json:"id"`            // Task ID // 任务ID
	UID           int64      `json:"uid"`           // User ID // 用户ID
	Vault         string     `json:"vault"`         // Associated vault name // 关联库名称
	RepoURL       string     `json:"repoUrl"`       // Repository URL // 仓库地址
	Username      string     `json:"username"`      // Username // 用户名
	Password      string     `json:"password"`      // Password // 密码
	Branch        string     `json:"branch"`        // Branch // 分支
	IsEnabled     bool       `json:"isEnabled"`     // Is enabled // 是否启用
	Delay         int64      `json:"delay"`         // Delay time (seconds) // 延迟时间（秒）
	RetentionDays int64      `json:"retentionDays"` // History retention days // 历史记录保留天数
	LastSyncTime  timex.Time `json:"lastSyncTime"`  // Last sync time // 上次同步时间
	LastStatus    int64      `json:"lastStatus"`    // Last status (0:Idle, 1:Running, 2:Success, 3:Failed, 4:Shutdown) // 上次状态 (0:Idle, 1:Running, 2:Success, 3:Failed, 4:Shutdown)
	LastMessage   string     `json:"lastMessage"`   // Last run result message // 上次运行结果消息
	CreatedAt     timex.Time `json:"createdAt"`     // Created at // 创建时间
	UpdatedAt     timex.Time `json:"updatedAt"`     // Updated at // 更新时间
}

// GitSyncHistoryRequest get sync history request
// GitSyncHistoryRequest 获取同步历史请求
type GitSyncHistoryRequest struct {
	ConfigID int64 `json:"configId" form:"configId"`
	Page     int   `json:"page" form:"page"`
	PageSize int   `json:"pageSize" form:"pageSize"`
}

// GitSyncHistoryDTO git sync history DTO
// GitSyncHistoryDTO git 同步历史 DTO
type GitSyncHistoryDTO struct {
	ID        int64      `json:"id"`
	ConfigID  int64      `json:"configId"`
	StartTime timex.Time `json:"startTime"`
	EndTime   timex.Time `json:"endTime"`
	Status    int64      `json:"status"` // 0:Idle, 1:Running, 2:Success, 3:Failed, 4:Shutdown
	Message   string     `json:"message"`
	CreatedAt timex.Time `json:"createdAt"`
}
