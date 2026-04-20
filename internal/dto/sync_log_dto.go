// Package dto Defines data transfer objects (request parameters and response structs)
// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

import "time"

// SyncLogListRequest Request parameters for listing sync logs
// SyncLogListRequest 查询同步日志列表的请求参数
type SyncLogListRequest struct {
	Vault  string `json:"vault" form:"vault" example:"MyVault"`          // Vault name (optional filter) // 保险库名称（可选过滤）
	Type   string `json:"type" form:"type" example:"note"`               // Resource type: note / file / setting / folder // 资源类型
	Action string `json:"action" form:"action" example:"modify"`         // Action type // 操作类型
}

// SyncLogDTO Sync log data transfer object
// SyncLogDTO 同步日志数据传输对象
type SyncLogDTO struct {
	ID            int64     `json:"id"`            // Record ID // 记录 ID
	VaultID       int64     `json:"vaultId"`       // Vault ID // 笔记本 ID
	Type          string    `json:"type"`          // Resource type // 资源类型
	Action        string    `json:"action"`        // Action type // 操作类型
	ChangedFields string    `json:"changedFields"` // Changed fields // 变更字段
	Path          string    `json:"path"`          // Resource path // 资源路径
	PathHash      string    `json:"pathHash"`      // Resource path hash // 路径哈希
	Size          int64     `json:"size"`          // Size in bytes // 大小（字节）
	ClientName    string    `json:"clientName"`    // Client name // 客户端名称
	Status        int       `json:"status"`        // Status: 1 success, 2 failed // 状态
	Message       string    `json:"message"`       // Additional message // 附加消息
	CreatedAt     time.Time `json:"createdAt"`     // Log creation time // 创建时间
}
