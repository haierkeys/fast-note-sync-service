// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

// VaultPostRequest 创建或更新保险库的请求参数
type VaultPostRequest struct {
	Vault string `json:"vault" form:"vault" binding:"required"`
	ID    int64  `json:"id" form:"id"`
}

// VaultGetRequest 获取保险库的请求参数
type VaultGetRequest struct {
	ID int64 `form:"id" binding:"required,gte=1"`
}

// VaultDTO Vault 数据传输对象
type VaultDTO struct {
	ID        int64  `json:"id"`
	Name      string `json:"vault"`
	NoteCount int64  `json:"noteCount"`
	NoteSize  int64  `json:"noteSize"`
	FileCount int64  `json:"fileCount"`
	FileSize  int64  `json:"fileSize"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
