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
