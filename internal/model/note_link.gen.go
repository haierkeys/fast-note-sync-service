package model

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

const TableNameNoteLink = "note_link"

// NoteLink mapped from table <note_link>
type NoteLink struct {
	ID             int64      `gorm:"column:id;primaryKey" json:"id" form:"id"`
	SourceNoteID   int64      `gorm:"column:source_note_id;not null;index:idx_source_note" json:"sourceNoteId" form:"sourceNoteId"`
	TargetPath     string     `gorm:"column:target_path;not null" json:"targetPath" form:"targetPath"`
	TargetPathHash string     `gorm:"column:target_path_hash;not null;index:idx_target_path_hash,priority:1" json:"targetPathHash" form:"targetPathHash"`
	LinkText       string     `gorm:"column:link_text" json:"linkText" form:"linkText"`
	IsEmbed        bool       `gorm:"column:is_embed;default:false" json:"isEmbed" form:"isEmbed"`
	VaultID        int64      `gorm:"column:vault_id;not null;index:idx_target_path_hash,priority:2" json:"vaultId" form:"vaultId"`
	UID            int64      `gorm:"column:uid;not null;index:idx_target_path_hash,priority:3" json:"uid" form:"uid"`
	CreatedAt      timex.Time `gorm:"column:created_at;type:datetime;default:NULL;autoCreateTime:false" json:"createdAt" form:"createdAt"`
}

// TableName NoteLink's table name
func (*NoteLink) TableName() string {
	return TableNameNoteLink
}
