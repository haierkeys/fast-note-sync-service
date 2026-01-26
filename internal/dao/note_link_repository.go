// Package dao implements the data access layer
package dao

import (
	"context"
	"strconv"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

// noteLinkRepository implements domain.NoteLinkRepository interface
type noteLinkRepository struct {
	dao             *Dao
	customPrefixKey string
}

// NewNoteLinkRepository creates a NoteLinkRepository instance
func NewNoteLinkRepository(dao *Dao) domain.NoteLinkRepository {
	return &noteLinkRepository{dao: dao, customPrefixKey: "user_note_link_"}
}

func (r *noteLinkRepository) GetKey(uid int64) string {
	return r.customPrefixKey + strconv.FormatInt(uid, 10)
}

// getDB gets the per-user database connection and ensures the table is migrated
func (r *noteLinkRepository) getDB(uid int64) *gorm.DB {
	key := r.GetKey(uid)
	migrateKey := key + "#migrated"

	db := r.dao.UserDB(uid)

	// Ensure migration only happens once per user
	if _, loaded := r.dao.onceKeys.LoadOrStore(migrateKey, true); !loaded {
		model.AutoMigrate(db, "NoteLink")
	}

	return db
}

// toDomain converts database model to domain model
func (r *noteLinkRepository) toDomain(m *model.NoteLink) *domain.NoteLink {
	if m == nil {
		return nil
	}
	return &domain.NoteLink{
		ID:             m.ID,
		SourceNoteID:   m.SourceNoteID,
		TargetPath:     m.TargetPath,
		TargetPathHash: m.TargetPathHash,
		LinkText:       m.LinkText,
		VaultID:        m.VaultID,
		CreatedAt:      time.Time(m.CreatedAt),
	}
}

// CreateBatch creates multiple note links in batch
func (r *noteLinkRepository) CreateBatch(ctx context.Context, links []*domain.NoteLink, uid int64) error {
	if len(links) == 0 {
		return nil
	}

	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		var models []*model.NoteLink
		now := timex.Now()
		for _, link := range links {
			models = append(models, &model.NoteLink{
				SourceNoteID:   link.SourceNoteID,
				TargetPath:     link.TargetPath,
				TargetPathHash: link.TargetPathHash,
				LinkText:       link.LinkText,
				VaultID:        link.VaultID,
				UID:            uid,
				CreatedAt:      now,
			})
		}
		return r.getDB(uid).WithContext(ctx).CreateInBatches(models, 100).Error
	})
}

// DeleteBySourceNoteID deletes all links from a source note
func (r *noteLinkRepository) DeleteBySourceNoteID(ctx context.Context, sourceNoteID, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		return r.getDB(uid).WithContext(ctx).
			Where("source_note_id = ?", sourceNoteID).
			Delete(&model.NoteLink{}).Error
	})
}

// GetBacklinks gets all notes that link to a target path
func (r *noteLinkRepository) GetBacklinks(ctx context.Context, targetPathHash string, vaultID, uid int64) ([]*domain.NoteLink, error) {
	var modelList []*model.NoteLink
	err := r.getDB(uid).WithContext(ctx).
		Where("target_path_hash = ? AND vault_id = ?", targetPathHash, vaultID).
		Find(&modelList).Error
	if err != nil {
		return nil, err
	}

	var results []*domain.NoteLink
	for _, m := range modelList {
		results = append(results, r.toDomain(m))
	}
	return results, nil
}

// GetOutlinks gets all links from a source note
func (r *noteLinkRepository) GetOutlinks(ctx context.Context, sourceNoteID, uid int64) ([]*domain.NoteLink, error) {
	var modelList []*model.NoteLink
	err := r.getDB(uid).WithContext(ctx).
		Where("source_note_id = ?", sourceNoteID).
		Find(&modelList).Error
	if err != nil {
		return nil, err
	}

	var results []*domain.NoteLink
	for _, m := range modelList {
		results = append(results, r.toDomain(m))
	}
	return results, nil
}

// Ensure noteLinkRepository implements domain.NoteLinkRepository interface
var _ domain.NoteLinkRepository = (*noteLinkRepository)(nil)
