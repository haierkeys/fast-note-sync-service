package domain

import (
	"context"
)

// NoteFTSRepository FTS full-text search repository interface
// NoteFTSRepository FTS 全文搜索仓库接口
type NoteFTSRepository interface {
	// Upsert inserts or updates FTS index
	// Upsert 插入或更新 FTS 索引
	Upsert(ctx context.Context, noteID int64, path, content string, uid int64) error
	// Delete deletes FTS index
	// Delete 删除 FTS 索引
	Delete(ctx context.Context, noteID int64, uid int64) error
	// Search full-text search, returns list of matching note_id
	// Search 全文搜索，返回匹配的 note_id 列表
	Search(ctx context.Context, keyword string, vaultID, uid int64, limit, offset int) ([]int64, error)
	// SearchCount full-text search count
	// SearchCount 全文搜索计数
	SearchCount(ctx context.Context, keyword string, vaultID, uid int64) (int64, error)
	// RebuildIndex rebuilds index (reads all note content from file system)
	// RebuildIndex 重建索引（从文件系统读取所有笔记内容）
	RebuildIndex(ctx context.Context, uid int64) error
	// DeleteByVaultID deletes all FTS records for a vault
	// DeleteByVaultID 删除指定仓库的所有 FTS 记录
	DeleteByVaultID(ctx context.Context, vaultID, uid int64) error
}
