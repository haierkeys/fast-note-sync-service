// Package domain defines domain models and interfaces
package domain

import "time"

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
