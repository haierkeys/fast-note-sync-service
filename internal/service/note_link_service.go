// Package service implements business logic layer
package service

import (
	"context"
	"errors"
	"strings"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"gorm.io/gorm"
)

// NoteLinkService defines the note link service interface
type NoteLinkService interface {
	// GetBacklinks gets all notes that link to a target note
	GetBacklinks(ctx context.Context, uid int64, params *dto.NoteLinkQueryRequest) ([]*dto.NoteLinkItem, error)

	// GetOutlinks gets all links from a source note
	GetOutlinks(ctx context.Context, uid int64, params *dto.NoteLinkQueryRequest) ([]*dto.NoteLinkItem, error)
}

// noteLinkService implements NoteLinkService interface
type noteLinkService struct {
	noteLinkRepo domain.NoteLinkRepository
	noteRepo     domain.NoteRepository
	vaultService VaultService
}

// NewNoteLinkService creates a NoteLinkService instance
func NewNoteLinkService(noteLinkRepo domain.NoteLinkRepository, noteRepo domain.NoteRepository, vaultService VaultService) NoteLinkService {
	return &noteLinkService{
		noteLinkRepo: noteLinkRepo,
		noteRepo:     noteRepo,
		vaultService: vaultService,
	}
}

// GetBacklinks gets all notes that link to a target note.
// Uses path variations to match links stored as partial paths (e.g., [[note]], [[folder/note]]).
func (s *noteLinkService) GetBacklinks(ctx context.Context, uid int64, params *dto.NoteLinkQueryRequest) ([]*dto.NoteLinkItem, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	// Generate all path variations for matching
	// e.g., "projects/folder/note.md" -> ["note", "folder/note", "projects/folder/note"]
	pathVariations := util.GeneratePathVariations(params.Path)
	if len(pathVariations) == 0 {
		return nil, nil
	}

	// Generate hashes for all variations
	var pathHashes []string
	for _, variation := range pathVariations {
		pathHashes = append(pathHashes, util.EncodeHash32(variation))
	}

	// Get backlinks matching any of the path variations
	links, err := s.noteLinkRepo.GetBacklinksByHashes(ctx, pathHashes, vaultID, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var results []*dto.NoteLinkItem
	for _, link := range links {
		// Get source note to get its path and content for context
		sourceNote, err := s.noteRepo.GetByID(ctx, link.SourceNoteID, uid)
		if err != nil {
			continue // Skip if note not found
		}

		item := &dto.NoteLinkItem{
			Path:     sourceNote.Path,
			LinkText: link.LinkText,
			IsEmbed:  link.IsEmbed,
		}

		// Extract context around the link (try all variations)
		for _, variation := range pathVariations {
			item.Context = s.extractLinkContext(sourceNote.Content, variation)
			if item.Context != "" {
				break
			}
		}

		results = append(results, item)
	}

	return results, nil
}

// GetOutlinks gets all links from a source note
func (s *noteLinkService) GetOutlinks(ctx context.Context, uid int64, params *dto.NoteLinkQueryRequest) ([]*dto.NoteLinkItem, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// Get note by path to get its ID
	note, err := s.noteRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorNoteNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Get outlinks from repository
	links, err := s.noteLinkRepo.GetOutlinks(ctx, note.ID, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var results []*dto.NoteLinkItem
	for _, link := range links {
		item := &dto.NoteLinkItem{
			Path:     link.TargetPath,
			LinkText: link.LinkText,
			IsEmbed:  link.IsEmbed,
		}

		// Extract context around the link
		item.Context = s.extractLinkContext(note.Content, link.TargetPath)

		results = append(results, item)
	}

	return results, nil
}

// extractLinkContext extracts approximately 50 characters of context around a link
func (s *noteLinkService) extractLinkContext(content, targetPath string) string {
	// Look for [[targetPath]] or [[targetPath|alias]]
	searchPatterns := []string{
		"[[" + targetPath + "]]",
		"[[" + targetPath + "|",
	}

	var pos int = -1
	var matchLen int

	for _, pattern := range searchPatterns {
		idx := strings.Index(content, pattern)
		if idx >= 0 && (pos < 0 || idx < pos) {
			pos = idx
			matchLen = len(pattern)
		}
	}

	if pos < 0 {
		return ""
	}

	// Extract context: 25 chars before and after the link
	contextRadius := 25
	start := pos - contextRadius
	if start < 0 {
		start = 0
	}

	// Find the end of the link (closing ]])
	linkEnd := strings.Index(content[pos:], "]]")
	if linkEnd < 0 {
		linkEnd = matchLen
	} else {
		linkEnd += 2 // Include ]]
	}

	end := pos + linkEnd + contextRadius
	if end > len(content) {
		end = len(content)
	}

	context := content[start:end]

	// Clean up: replace newlines with spaces and trim
	context = strings.ReplaceAll(context, "\n", " ")
	context = strings.TrimSpace(context)

	// Add ellipsis if truncated
	if start > 0 {
		context = "..." + context
	}
	if end < len(content) {
		context = context + "..."
	}

	return context
}

// Ensure noteLinkService implements NoteLinkService interface
var _ NoteLinkService = (*noteLinkService)(nil)
