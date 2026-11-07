package domain

import "context"

// NoteTag is one (note, tag) association row.
// NoteTag 是一条"笔记-标签"关联。
type NoteTag struct {
	ID      int64
	NoteID  int64
	Tag     string
	VaultID int64
	UID     int64
	Source  string // frontmatter | inline | both
}

// TagCount is a tag with its usage count inside one vault.
// TagCount 是一个标签及其在某个库中的引用次数。
type TagCount struct {
	Tag   string
	Count int
}

// NoteTagRepository persists the note->tag index.
// NoteTagRepository 维护 笔记->标签 索引。
type NoteTagRepository interface {
	// ReplaceForNote atomically replaces every tag row of one note.
	// Passing an empty slice clears the note's tags.
	// ReplaceForNote 原子替换一篇笔记的全部标签行；传空切片表示清空。
	ReplaceForNote(ctx context.Context, noteID int64, tags []string, vaultID, uid int64, source string) error

	// DeleteByNoteID removes all tag rows of one note.
	// DeleteByNoteID 删除一篇笔记的全部标签行。
	DeleteByNoteID(ctx context.Context, noteID, uid int64) error

	// DeleteByVaultID removes all tag rows of a vault.
	// DeleteByVaultID 删除一个库的全部标签行。
	DeleteByVaultID(ctx context.Context, vaultID, uid int64) error

	// ListByVault returns tag usage counts for a vault, ordered by count desc
	// then tag asc. prefix filters tags by prefix (case-insensitive), empty means
	// all; scope restricts the source ('frontmatter' | 'inline'), empty or 'all'
	// means every source; limit <= 0 means unlimited.
	// ListByVault 返回一个库的标签引用计数，按次数降序、标签名升序。
	// prefix 按前缀过滤（忽略大小写），空表示不过滤；
	// scope 限定来源（frontmatter | inline），空或 all 表示全部；limit <= 0 表示不限。
	ListByVault(ctx context.Context, vaultID, uid int64, prefix, scope string, limit int) ([]*TagCount, error)

	// GetTagsByNoteID returns the tags of one note.
	// GetTagsByNoteID 返回一篇笔记的标签。
	GetTagsByNoteID(ctx context.Context, noteID, uid int64) ([]string, error)

	// RebuildForVault drops and re-derives tag rows for every note in a vault
	// by scanning note content. Used by the upgrade migration and by explicit rebuilds.
	// RebuildForVault 扫描库内全部笔记正文，重建标签行。用于升级迁移和显式重建。
	RebuildForVault(ctx context.Context, vaultID, uid int64) (int, error)
}
