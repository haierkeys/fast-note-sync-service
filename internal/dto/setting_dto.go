// Package dto Defines data transfer objects (request parameters and response structs)
// Package dto 定义数据传输对象（请求参数和响应结构体）
package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// SettingUpdateCheckRequest Client request parameters for checking setting updates
// 客户端检查更新请求参数
type SettingUpdateCheckRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`
	Path        string `json:"path" form:"path" binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash"`
	Ctime       int64  `json:"ctime" form:"ctime" binding:"required"`
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required"`
}

// SettingModifyOrCreateRequest Request parameters for creating or modifying settings
// 修改或创建配置参数
type SettingModifyOrCreateRequest struct {
	Vault       string `json:"vault" form:"vault" binding:"required"`
	Path        string `json:"path" form:"path" binding:"required"`
	PathHash    string `json:"pathHash" form:"pathHash"`
	Content     string `json:"content" form:"content"`
	ContentHash string `json:"contentHash" form:"contentHash"`
	Ctime       int64  `json:"ctime" form:"ctime"`
	Mtime       int64  `json:"mtime" form:"mtime"`
}

// SettingDeleteRequest Parameters for deleting settings
// 删除配置参数
type SettingDeleteRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash"`
}

// SettingClearRequest Parameters for clearing settings
// 清除配置参数
type SettingClearRequest struct {
	Vault string `json:"vault" form:"vault" binding:"required"`
}

// SettingGetRequest Parameters for retrieving a single setting
// 获取单条配置参数
type SettingGetRequest struct {
	Vault    string `json:"vault" form:"vault" binding:"required"`
	Path     string `json:"path" form:"path"`
	PathHash string `json:"pathHash" form:"pathHash"`
}

type SettingSyncDelSetting struct {
	Path     string `json:"path" form:"path" binding:"required"`
	PathHash string `json:"pathHash" form:"pathHash" binding:"required"`
}

// SettingSyncRequest Synchronization request parameters
// 同步请求参数
type SettingSyncRequest struct {
	Vault           string                    `json:"vault" form:"vault" binding:"required"`
	LastTime        int64                     `json:"lastTime" form:"lastTime"`
	Cover           bool                      `json:"cover" form:"cover"`
	Settings        []SettingSyncCheckRequest `json:"settings" form:"settings"`
	DelSettings     []SettingSyncDelSetting   `json:"delSettings" form:"delSettings"`
	MissingSettings []SettingSyncDelSetting   `json:"missingFiles" form:"missingFiles"`
}

// SettingSyncCheckRequest Parameters for checking synchronization of a single setting
// 单条同步检查参数
type SettingSyncCheckRequest struct {
	Path        string `json:"path" form:"path"`
	PathHash    string `json:"pathHash" form:"pathHash" binding:"required"`
	ContentHash string `json:"contentHash" form:"contentHash"`
	Mtime       int64  `json:"mtime" form:"mtime" binding:"required"`
}

// SettingDTO Setting data transfer object
// SettingDTO 配置数据传输对象
type SettingDTO struct {
	ID               int64      `json:"id" form:"id"`
	Action           string     `json:"-" form:"action"`
	Path             string     `json:"path" form:"path"`
	PathHash         string     `json:"pathHash" form:"pathHash"`
	Content          string     `json:"content" form:"content"`
	ContentHash      string     `json:"contentHash" form:"contentHash"`
	Ctime            int64      `json:"ctime" form:"ctime"`
	Mtime            int64      `json:"mtime" form:"mtime"`
	UpdatedTimestamp int64      `json:"lastTime" form:"updatedTimestamp"`
	UpdatedAt        timex.Time `json:"updatedAt"`
	CreatedAt        timex.Time `json:"createdAt"`
}
