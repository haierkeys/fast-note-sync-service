// Package dto Defines data transfer objects (request parameters and response structs)
// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// SettingUpdateCheckRequest Client request parameters for checking setting updates
// 客户端检查更新请求参数
type SettingUpdateCheckRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`       // Vault name // 保险库名称
	Path        string `json:"path" form:"path" binding:"required"`         // Setting path // 配置路径
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"` // Path hash // 路径哈希
	ContentHash string `json:"contentHash" form:"contentHash"`              // Content hash // 内容哈希
	Ctime       int64  `json:"ctime" form:"ctime" binding:"required"`       // Creation timestamp // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required"`       // Modification timestamp // 修改时间戳
}

// SettingModifyOrCreateRequest Request parameters for creating or modifying settings
// 修改或创建配置参数
type SettingModifyOrCreateRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Path        string `json:"path" form:"path" binding:"required"`   // Setting path // 配置路径
	PathHash    string `json:"pathHash" form:"pathHash"`              // Path hash // 路径哈希
	Content     string `json:"content" form:"content"`                // Setting content // 配置内容
	ContentHash string `json:"contentHash" form:"contentHash"`        // Content hash // 内容哈希
	Ctime       int64  `json:"ctime" form:"ctime"`                    // Creation timestamp // 创建时间戳
	Mtime       int64  `json:"mtime" form:"mtime"`                    // Modification timestamp // 修改时间戳
}

// SettingDeleteRequest Parameters for deleting settings
// 删除配置参数
type SettingDeleteRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Path     string `json:"path" form:"path" binding:"required"`   // Setting path // 配置路径
	PathHash string `json:"pathHash" form:"pathHash"`              // Path hash // 路径哈希
}

// SettingClearRequest Parameters for clearing settings
// 清除配置参数
type SettingClearRequest struct {
	Vault string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
}

// SettingGetRequest Parameters for retrieving a single setting
// 获取单条配置参数
type SettingGetRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	Path     string `json:"path" form:"path"`                      // Setting path // 配置路径
	PathHash string `json:"pathHash" form:"pathHash"`              // Path hash // 路径哈希
}

// SettingSyncCheckRequest Parameters for checking synchronization of a single setting
// 单条同步检查参数
type SettingSyncCheckRequest struct {
	Path        string `json:"path" form:"path"`                            // Setting path // 配置路径
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"` // Path hash // 路径哈希
	ContentHash string `json:"contentHash" form:"contentHash"`              // Content hash // 内容哈希
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required"`       // Modification timestamp // 修改时间戳
}

// SettingSyncDelSetting Parameters for deleting sets during sync
// 同步删除配置参数
type SettingSyncDelSetting struct {
	Path     string `json:"path" form:"path" binding:"required"`         // Setting path // 配置路径
	PathHash string `json:"pathHash" form:"pathHash" binding:"required"` // Path hash // 路径哈希
}

// SettingSyncRequest Synchronization request parameters
// 同步请求参数
type SettingSyncRequest struct {
	Vault           string                    `json:"vault" form:"vault" binding:"required"` // Vault name // 保险库名称
	LastTime        int64                     `json:"lastTime" form:"lastTime"`              // Last sync time // 最后同步时间
	Cover           bool                      `json:"cover" form:"cover"`                    // Whether to cover existing // 是否覆盖现有配置
	Settings        []SettingSyncCheckRequest `json:"settings" form:"settings"`              // Settings to check // 待检查配置列表
	DelSettings     []SettingSyncDelSetting   `json:"delSettings" form:"delSettings"`        // Settings to delete // 待删除配置列表
	MissingSettings []SettingSyncDelSetting   `json:"missingFiles" form:"missingFiles"`      // Missing settings // 缺失配置列表
}

// ---------------- DTO / Response ----------------

// SettingDTO Setting data transfer object
// SettingDTO 配置数据传输对象
type SettingDTO struct {
	ID               int64      `json:"id" form:"id"`                     // Setting ID // 配置 ID
	Action           string     `json:"-" form:"action"`                  // Action // 动作
	Path             string     `json:"path" form:"path"`                 // Setting path // 配置路径
	PathHash         string     `json:"pathHash" form:"pathHash"`         // Path hash // 路径哈希值
	Content          string     `json:"content" form:"content"`           // Setting content // 配置内容
	ContentHash      string     `json:"contentHash" form:"contentHash"`   // Content hash // 内容哈希
	Ctime            int64      `json:"ctime" form:"ctime"`               // Creation timestamp // 创建时间戳
	Mtime            int64      `json:"mtime" form:"mtime"`               // Modification timestamp // 修改时间戳
	UpdatedTimestamp int64      `json:"lastTime" form:"updatedTimestamp"` // Record update timestamp // 记录更新时间戳
	UpdatedAt        timex.Time `json:"updatedAt"`                        // Updated at time // 更新时间
	CreatedAt        timex.Time `json:"createdAt"`                        // Created at time // 创建时间
}
