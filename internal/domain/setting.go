// Package domain 定义领域模型和接口
package domain

import "time"

// SettingAction 定义配置操作类型
type SettingAction string

const (
	SettingActionCreate SettingAction = "create"
	SettingActionModify SettingAction = "modify"
	SettingActionDelete SettingAction = "delete"
)

// Setting 配置领域模型
type Setting struct {
	ID               int64
	VaultID          int64
	Action           SettingAction
	Path             string
	PathHash         string
	Content          string
	ContentHash      string
	Size             int64
	Ctime            int64
	Mtime            int64
	UpdatedTimestamp int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IsDeleted 判断配置是否已删除
func (s *Setting) IsDeleted() bool {
	return s.Action == SettingActionDelete
}

// IsCreated 判断配置是否为新建
func (s *Setting) IsCreated() bool {
	return s.Action == SettingActionCreate
}

// IsModified 判断配置是否已修改
func (s *Setting) IsModified() bool {
	return s.Action == SettingActionModify
}
