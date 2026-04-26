// Package domain 定义领域模型和接口
package domain

import (
	"context"
	"time"
)


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
	ClientType  string
	ClientVersion string
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NoteHistoryRepository 笔记历史仓储接口
type NoteHistoryRepository interface {
	// GetByID 根据ID获取历史记录
	GetByID(ctx context.Context, id, uid int64) (*NoteHistory, error)

	// GetByNoteIDAndHash 根据笔记ID和内容哈希获取历史记录
	GetByNoteIDAndHash(ctx context.Context, noteID int64, contentHash string, uid int64) (*NoteHistory, error)

	// Create 创建历史记录
	Create(ctx context.Context, history *NoteHistory, uid int64) (*NoteHistory, error)

	// ListByNoteID 根据笔记ID获取历史记录列表
	ListByNoteID(ctx context.Context, noteID int64, page, pageSize int, uid int64) ([]*NoteHistory, int64, error)

	// GetLatestVersion 获取笔记的最新版本号
	GetLatestVersion(ctx context.Context, noteID, uid int64) (int64, error)

	// Migrate 迁移历史记录（更新 NoteID）
	Migrate(ctx context.Context, oldNoteID, newNoteID, uid int64) error

	// GetNoteIDsWithOldHistory 获取有旧历史记录的笔记ID列表
	// cutoffTime: 截止时间戳（毫秒），返回有早于此时间历史记录的笔记ID
	GetNoteIDsWithOldHistory(ctx context.Context, cutoffTime int64, uid int64) ([]int64, error)

	// DeleteOldVersions 删除旧版本历史记录，保留最近 N 个版本
	// noteID: 笔记ID
	// cutoffTime: 截止时间戳（毫秒），删除早于此时间的记录
	// keepVersions: 保留的最近版本数量
	DeleteOldVersions(ctx context.Context, noteID int64, cutoffTime int64, keepVersions int, uid int64) error

	// Delete 删除指定ID的历史记录
	Delete(ctx context.Context, id, uid int64) error

	// DeleteByVaultID 删除仓库下的所有历史记录（包含物理目录）
	DeleteByVaultID(ctx context.Context, vaultID, uid int64) error
}

