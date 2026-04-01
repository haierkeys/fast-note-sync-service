package service

import (
	"context"
	"testing"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/internal/mocks"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNoteService_Get(t *testing.T) {
	mockNoteRepo := new(mocks.NoteRepositoryMock)
	mockVaultSvc := new(mocks.VaultServiceMock)
	mockNoteLinkRepo := new(mocks.NoteRepositoryMock) // Simplified for this test
	
	s := &noteService{
		noteRepo:     mockNoteRepo,
		vaultService: mockVaultSvc,
	}

	ctx := context.Background()
	uid := int64(1)
	vaultName := "testvault"
	vaultID := int64(100)
	pathHash := "hash123"

	t.Run("Success", func(t *testing.T) {
		params := &dto.NoteGetRequest{
			Vault:     vaultName,
			PathHash:  pathHash,
			IsRecycle: false,
		}

		mockVaultSvc.On("MustGetID", ctx, uid, vaultName).Return(vaultID, nil)
		mockNoteRepo.On("GetByPathHashIncludeRecycle", ctx, pathHash, vaultID, uid, false).Return(&domain.Note{
			ID:       1,
			Path:     "note1.md",
			PathHash: pathHash,
			Content:  "hello",
		}, nil)

		res, err := s.Get(ctx, uid, params)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "note1.md", res.Path)
		mockVaultSvc.AssertExpectations(t)
		mockNoteRepo.AssertExpectations(t)
	})

	t.Run("VaultNotFound", func(t *testing.T) {
		params := &dto.NoteGetRequest{
			Vault: "invalid",
		}

		mockVaultSvc.On("MustGetID", ctx, uid, "invalid").Return(int64(0), code.ErrorVaultNotFound)

		res, err := s.Get(ctx, uid, params)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, code.ErrorVaultNotFound, err)
	})

	t.Run("NoteNotFound", func(t *testing.T) {
		params := &dto.NoteGetRequest{
			Vault:    vaultName,
			PathHash: "nonexistent",
		}

		mockVaultSvc.On("MustGetID", ctx, uid, vaultName).Return(vaultID, nil)
		mockNoteRepo.On("GetByPathHashIncludeRecycle", ctx, "nonexistent", vaultID, uid, false).Return(nil, code.ErrorNoteNotFound)

		res, err := s.Get(ctx, uid, params)

		assert.Error(t, err)
		assert.Nil(t, res)
		// NoteRepo might return gorm.ErrRecordNotFound which is then wrapped by Service as code.ErrorNoteNotFound
	})
}
