// Package domain defines the core business domain models and repository interfaces
// Package domain 定义核心业务领域模型和仓储接口
package domain

import (
	"context"

	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
)

// SyncLogType represents the type of resource being synchronized
// SyncLogType 表示同步的资源类型
type SyncLogType string

// SyncLogAction represents the type of synchronization action
// SyncLogAction 表示同步操作类型
type SyncLogAction string

const (
	// SyncLogTypeNote represents a note resource
	// SyncLogTypeNote 表示笔记资源
	SyncLogTypeNote SyncLogType = "note"

	// SyncLogTypeFile represents a file (attachment) resource
	// SyncLogTypeFile 表示文件（附件）资源
	SyncLogTypeFile SyncLogType = "file"

	// SyncLogTypeSetting represents a configuration resource
	// SyncLogTypeSetting 表示配置资源
	SyncLogTypeSetting SyncLogType = "setting"

	// SyncLogTypeFolder represents a folder resource
	// SyncLogTypeFolder 表示文件夹资源
	SyncLogTypeFolder SyncLogType = "folder"

	// SyncLogActionCreate represents a create action
	// SyncLogActionCreate 表示新建操作
	SyncLogActionCreate SyncLogAction = "create"

	// SyncLogActionModify represents a modify action (content or mtime changed)
	// SyncLogActionModify 表示修改操作（内容或时间戳变更）
	SyncLogActionModify SyncLogAction = "modify"

	// SyncLogActionSoftDelete represents moving a resource to the recycle bin
	// SyncLogActionSoftDelete 表示将资源移至回收站（软删除）
	SyncLogActionSoftDelete SyncLogAction = "soft_delete"

	// SyncLogActionDelete represents permanently deleting a resource from the recycle bin
	// SyncLogActionDelete 表示从回收站彻底删除（物理删除）
	SyncLogActionDelete SyncLogAction = "delete"

	// SyncLogActionRename represents renaming a resource
	// SyncLogActionRename 表示重命名操作
	SyncLogActionRename SyncLogAction = "rename"

	// SyncLogActionRestore represents restoring a resource from the recycle bin
	// SyncLogActionRestore 表示从回收站恢复
	SyncLogActionRestore SyncLogAction = "restore"
)

// SyncLog represents a synchronization log entry
// SyncLog 同步日志领域模型
type SyncLog struct {
	ID            int64         // Record ID // 记录 ID
	UID           int64         // User ID // 用户 ID
	VaultID       int64         // Vault ID // 笔记本 ID
	Type          SyncLogType   // Resource type: note / file / setting // 资源类型
	Action        SyncLogAction // Action type: create / modify / soft_delete / delete / rename / restore // 操作类型
	ChangedFields string        // Comma-separated changed fields, e.g. "content,mtime" / "mtime" / "path" // 逗号分隔的变更字段
	Path          string        // Resource path // 资源路径
	PathHash      string        // Resource path hash // 资源路径哈希
	Size          int64         // Resource size in bytes // 资源大小（字节）
	ClientName    string        // Client name that initiated the sync // 发起同步的客户端名称
	ClientType    string        // Client type // 客户端类型
	ClientVersion string        // Client version // 客户端版本
	Status        int           // 1: success, 2: failed // 状态：1 成功，2 失败
	Message       string        // Additional message or error detail // 附加消息或错误详情
	CreatedAt     timex.Time    // Log creation time // 日志创建时间
}

// SyncLogRepository defines the data access interface for sync logs
// SyncLogRepository 定义同步日志的数据访问接口
type SyncLogRepository interface {
	// Create stores a new sync log entry
	// Create 存储一条新的同步日志
	Create(ctx context.Context, log *SyncLog, uid int64) error

	// List retrieves sync logs for a user with filtering and pagination
	// List 按条件分页查询用户的同步日志
	List(ctx context.Context, uid int64, logType, action string, page, pageSize int) ([]*SyncLog, int64, error)

	// CleanupByTime removes sync logs older than the given timestamp for a specific user
	// CleanupByTime 清理指定用户在指定时间戳之前的同步日志
	CleanupByTime(ctx context.Context, timestamp int64, uid int64) error

	// CleanupByTimeAll removes sync logs older than the given timestamp for all users
	// CleanupByTimeAll 清理所有用户在指定时间戳之前的同步日志
	CleanupByTimeAll(ctx context.Context, timestamp int64) error
}
