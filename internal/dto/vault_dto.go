// Package dto Defines data transfer objects (request parameters and response structs)
// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

// VaultPostRequest Request parameters for creating or updating a vault
// 创建或更新保险库的请求参数
type VaultPostRequest struct {
	Vault string `json:"vault" form:"vault" binding:"required" example:"MyVault"` // Vault name // 保险库名称
	ID    int64  `json:"id" form:"id" example:"1"`                                // Vault ID (optional for update) // 保险库 ID（可选，用于更新）
}

// VaultGetRequest Request parameters for retrieving a vault
// 获取保险库的请求参数
type VaultGetRequest struct {
	ID int64 `form:"id" binding:"required,gte=1" example:"1"` // Vault ID // 保险库 ID
}

// ---------------- DTO / Response ----------------

// VaultDTO Vault data transfer object
// VaultDTO Vault 数据传输对象
type VaultDTO struct {
	ID        int64  `json:"id"`        // Vault ID // 保险库 ID
	Name      string `json:"vault"`     // Vault name // 保险库名称
	NoteCount int64  `json:"noteCount"` // Number of notes // 笔记数量
	NoteSize  int64  `json:"noteSize"`  // Size of notes // 笔记大小
	FileCount int64  `json:"fileCount"` // Number of files // 文件数量
	FileSize  int64  `json:"fileSize"`  // Size of files // 文件大小
	Size      int64  `json:"size"`      // Total size // 总大小
	CreatedAt string `json:"createdAt"` // Creation time // 创建时间
	UpdatedAt string `json:"updatedAt"` // Updated time // 更新时间
}
