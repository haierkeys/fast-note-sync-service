// Package model 定义数据模型
package model

import (
	"gorm.io/gorm"
)

// FTS 表版本号，修改此值会触发重建索引
const NoteFTSVersion = 2

// NoteFTS FTS5 全文搜索虚拟表
// 注意：这不是普通的 GORM 模型，需要手动创建 FTS5 虚拟表
type NoteFTS struct {
	NoteID  int64  `gorm:"column:note_id" json:"noteId"`
	Path    string `gorm:"column:path" json:"path"`
	Content string `gorm:"column:content" json:"content"`
}

// TableName 返回表名
func (*NoteFTS) TableName() string {
	return "note_fts"
}

// NoteFTSMeta FTS 元数据表，用于存储版本信息
type NoteFTSMeta struct {
	Key   string `gorm:"column:key;primaryKey"`
	Value string `gorm:"column:value"`
}

func (*NoteFTSMeta) TableName() string {
	return "note_fts_meta"
}

// CreateNoteFTSTable 创建 FTS5 虚拟表
// FTS5 是 SQLite 的全文搜索扩展，专门用于高效的文本搜索
func CreateNoteFTSTable(db *gorm.DB) error {
	// 创建元数据表
	db.Exec(`CREATE TABLE IF NOT EXISTS note_fts_meta (key TEXT PRIMARY KEY, value TEXT)`)

	// 检查版本
	var meta NoteFTSMeta
	db.Where("key = ?", "version").First(&meta)
	currentVersion := meta.Value

	// 如果版本不匹配，删除旧表重建
	if currentVersion != "" && currentVersion != string(rune(NoteFTSVersion+'0')) {
		_ = DropNoteFTSTable(db)
	}

	// 检查 FTS 表是否已存在
	var count int64
	db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='note_fts'").Scan(&count)
	if count > 0 {
		return nil
	}

	// 创建 FTS5 虚拟表
	// tokenize='unicode61 remove_diacritics 2' 支持 Unicode 字符分词
	sql := `
		CREATE VIRTUAL TABLE IF NOT EXISTS note_fts USING fts5(
			note_id UNINDEXED,
			path,
			content,
			tokenize='unicode61 remove_diacritics 2'
		)
	`
	if err := db.Exec(sql).Error; err != nil {
		return err
	}

	// 更新版本号
	db.Exec(`INSERT OR REPLACE INTO note_fts_meta (key, value) VALUES ('version', ?)`, string(rune(NoteFTSVersion+'0')))

	return nil
}

// DropNoteFTSTable 删除 FTS5 表（用于重建索引）
func DropNoteFTSTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS note_fts").Error
}
