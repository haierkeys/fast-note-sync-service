// Package model 定义数据模型
package model

import (
	"gorm.io/gorm"
)

// FTS 表版本号，修改此值会触发重建索引
const NoteFTSVersion = 3

// NoteFTS 存储笔记全文搜索的快照数据
type NoteFTS struct {
	NoteID  int64  `gorm:"column:note_id;primaryKey" json:"noteId"`
	Path    string `gorm:"column:path" json:"path"`
	Content string `gorm:"column:content" json:"content"`
}

// TableName 返回表名
func (*NoteFTS) TableName() string {
	return "note_fts"
}

// NoteFTSToken 倒排索引表，支持多数据库全文搜索
type NoteFTSToken struct {
	ID     int64  `gorm:"column:id;primaryKey;autoIncrement"`
	NoteID int64  `gorm:"column:note_id;index:idx_fts_note_id;index:idx_fts_token_note_id,priority:2"`
	Token  string `gorm:"column:token;type:varchar(255);index:idx_fts_token;index:idx_fts_token_note_id,priority:1"`
}

func (*NoteFTSToken) TableName() string {
	return "note_fts_token"
}

// NoteFTSMeta FTS 元数据表，用于存储版本信息
type NoteFTSMeta struct {
	Key   string `gorm:"column:key;primaryKey"`
	Value string `gorm:"column:value"`
}

func (*NoteFTSMeta) TableName() string {
	return "note_fts_meta"
}

// CreateNoteFTSTable 创建搜索相关的标准数据库表
func CreateNoteFTSTable(db *gorm.DB) error {
	// 创建元数据表
	if err := db.AutoMigrate(&NoteFTSMeta{}); err != nil {
		return err
	}

	// 检查版本
	var meta NoteFTSMeta
	db.Where("key = ?", "version").First(&meta)
	currentVersion := meta.Value

	// 如果版本不匹配，删除旧表重建
	if currentVersion != "" && currentVersion != string(rune(NoteFTSVersion+'0')) {
		_ = DropNoteFTSTable(db)
	}

	// 执行自动迁移
	if err := db.AutoMigrate(&NoteFTS{}, &NoteFTSToken{}); err != nil {
		return err
	}

	// 更新版本号
	db.Save(&NoteFTSMeta{Key: "version", Value: string(rune(NoteFTSVersion + '0'))})

	return nil
}

// DropNoteFTSTable 删除全文搜索相关的表
func DropNoteFTSTable(db *gorm.DB) error {
	_ = db.Migrator().DropTable(&NoteFTS{})
	_ = db.Migrator().DropTable(&NoteFTSToken{})
	return nil
}
