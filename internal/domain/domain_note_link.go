// Package domain defines domain models and interfaces
package domain

import (
	"context"
	"time"
)

// NoteLink represents a wiki-style link between notes
type NoteLink struct {
	ID             int64
	SourceNoteID   int64
	TargetPath     string
	TargetPathHash string
	LinkText       string // alias from [[link|alias]]
	IsEmbed        bool   // true if embed (![[...]]) vs regular link ([[...]])
	VaultID        int64
	CreatedAt      time.Time
}

// NoteLinkRepository note link repository interface
type NoteLinkRepository interface {
	// CreateBatch creates multiple note links in batch
	CreateBatch(ctx context.Context, links []*NoteLink, uid int64) error

	// DeleteBySourceNoteID deletes all links from a source note
	DeleteBySourceNoteID(ctx context.Context, sourceNoteID, uid int64) error

	// GetBacklinks gets all notes that link to a target path
	GetBacklinks(ctx context.Context, targetPathHash string, vaultID, uid int64) ([]*NoteLink, error)

	// GetBacklinksByHashes gets all notes that link to any of the target path hashes
	// Used for matching path variations (e.g., [[note]], [[folder/note]], [[full/path/note]])
	GetBacklinksByHashes(ctx context.Context, targetPathHashes []string, vaultID, uid int64) ([]*NoteLink, error)

	// GetOutlinks gets all links from a source note
	GetOutlinks(ctx context.Context, sourceNoteID, uid int64) ([]*NoteLink, error)

	// DeleteByVaultID deletes all links for a vault
	DeleteByVaultID(ctx context.Context, vaultID, uid int64) error
}
