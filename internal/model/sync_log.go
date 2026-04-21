// Code generated manually. DO NOT EDIT by code generator.
// 手动编写，请勿由代码生成器覆盖。

package model

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

const TableNameSyncLog = "sync_log"

// SyncLog mapped from table <sync_log>
// SyncLog 对应数据库表 sync_log
type SyncLog struct {
	ID            int64      `gorm:"column:id;primaryKey" json:"id" form:"id"`
	UID           int64      `gorm:"column:uid;not null;default:0" json:"uid" form:"uid"`
	VaultID       int64      `gorm:"column:vault_id;not null;default:0" json:"vaultId" form:"vaultId"`
	Type          string     `gorm:"column:type;not null;default:''" json:"type" form:"type"`
	Action        string     `gorm:"column:action;not null;default:''" json:"action" form:"action"`
	ChangedFields string     `gorm:"column:changed_fields;not null;default:''" json:"changedFields" form:"changedFields"`
	Path          string     `gorm:"column:path;default:''" json:"path" form:"path"`
	PathHash      string     `gorm:"column:path_hash;default:''" json:"pathHash" form:"pathHash"`
	Size          int64      `gorm:"column:size;default:0" json:"size" form:"size"`
	ClientName    string     `gorm:"column:client_name;default:''" json:"clientName" form:"clientName"`
	ClientType    string     `gorm:"column:client_type;default:''" json:"clientType" form:"clientType"`
	ClientVersion string     `gorm:"column:client_version;default:''" json:"clientVersion" form:"clientVersion"`
	Status        int        `gorm:"column:status;default:1" json:"status" form:"status"`
	Message       string     `gorm:"column:message;default:''" json:"message" form:"message"`
	CreatedAt     timex.Time `gorm:"column:created_at;default:NULL;autoCreateTime:false" json:"createdAt" form:"createdAt"`
}

// TableName SyncLog's table name
// TableName 返回 SyncLog 的表名
func (*SyncLog) TableName() string {
	return TableNameSyncLog
}
