// Package domain 定义领域模型和接口
package domain

import "time"

// NoteHistory 笔记历史领域模型
type NoteHistory struct {
	ID          int64
	NoteID      int64
	VaultID     int64
	Path        string
	DiffPatch   string
	Content     string
	ContentHash string
	ClientName  string
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
