package dao

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate models
	err = db.AutoMigrate(&domain.Note{}, &domain.Vault{})
	require.NoError(t, err)

	return db
}

func TestNoteRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNoteRepository(db)
	ctx := context.Background()
	uid := int64(1)
	vaultID := int64(100)

	t.Run("CreateAndGet", func(t *testing.T) {
		note := &domain.Note{
			VaultID:     vaultID,
			Path:        "test.md",
			PathHash:    "hash1",
			Content:     "hello world",
			ContentHash: "chash1",
			Action:      domain.NoteActionCreate,
			Mtime:       time.Now().UnixMilli(),
		}

		created, err := repo.Create(ctx, note, uid)
		assert.NoError(t, err)
		assert.NotZero(t, created.ID)

		fetched, err := repo.GetByID(ctx, created.ID, uid)
		assert.NoError(t, err)
		assert.Equal(t, "test.md", fetched.Path)
		assert.Equal(t, "hello world", fetched.Content)
	})

	t.Run("Update", func(t *testing.T) {
		note, _ := repo.GetByPathHash(ctx, "hash1", vaultID, uid)
		note.Content = "updated content"
		note.ContentHash = "chash2"
		
		updated, err := repo.Update(ctx, note, uid)
		assert.NoError(t, err)
		assert.Equal(t, "updated content", updated.Content)

		fetched, _ := repo.GetByID(ctx, note.ID, uid)
		assert.Equal(t, "updated content", fetched.Content)
	})

	t.Run("Delete", func(t *testing.T) {
		note, _ := repo.GetByPathHash(ctx, "hash1", vaultID, uid)
		err := repo.Delete(ctx, note.ID, uid)
		assert.NoError(t, err)

		_, err = repo.GetByID(ctx, note.ID, uid)
		assert.Error(t, err) // Should return gorm.ErrRecordNotFound
	})
}
