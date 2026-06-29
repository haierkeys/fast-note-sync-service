package dto

// SyncPageMessage 服务端分页下发控制消息
// SyncPageMessage server-to-client page control message
type SyncPageMessage struct {
	Context    string `json:"context"`    // Sync context // 同步上下文
	PageIndex  int    `json:"pageIndex"`  // Current page index // 当前页码 (0-indexed)
	PageSize   int    `json:"pageSize"`   // Messages per page // 每页大小
	TotalCount int    `json:"totalCount"` // Count of messages in current page // 当前页包含的消息数量
	IsLast     bool   `json:"isLast"`     // Is the final page // 是否为最后一页
}

// SyncPageAckRequest 客户端确认接收完当前页所有下载任务的请求
// SyncPageAckRequest request from client to ack completion of current page
type SyncPageAckRequest struct {
	Context   string `json:"context" form:"context" binding:"required"`     // Sync context // 同步上下文
	Vault     string `json:"vault" form:"vault" binding:"required"`         // Vault name // 笔记库名称
	PageIndex int    `json:"pageIndex" form:"pageIndex" binding:"min=0"` // Acked page index // 已确认接收完成的页码 (0-indexed)
}
