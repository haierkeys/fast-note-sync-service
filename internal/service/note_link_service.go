// Package service implements business logic layer
// Package service 实现业务逻辑层
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
// NoteLinkService 定义笔记链接服务接口
type NoteLinkService interface {
	// GetBacklinks gets all notes that link to a target note
	// GetBacklinks 获取链接到目标笔记的所有笔记
	GetBacklinks(ctx context.Context, uid int64, params *dto.NoteLinkQueryRequest) ([]*dto.NoteLinkItem, error)

	// GetOutlinks gets all links from a source note
	// GetOutlinks 获取源笔记中的所有链接
	GetOutlinks(ctx context.Context, uid int64, params *dto.NoteLinkQueryRequest) ([]*dto.NoteLinkItem, error)
}

// noteLinkService implements NoteLinkService interface
// noteLinkService 实现 NoteLinkService 接口
type noteLinkService struct {
	noteLinkRepo domain.NoteLinkRepository
	noteRepo     domain.NoteRepository
	vaultService VaultService
}

// NewNoteLinkService creates a NoteLinkService instance
// NewNoteLinkService 创建 NoteLinkService 实例
func NewNoteLinkService(noteLinkRepo domain.NoteLinkRepository, noteRepo domain.NoteRepository, vaultService VaultService) NoteLinkService {
	return &noteLinkService{
		noteLinkRepo: noteLinkRepo,
		noteRepo:     noteRepo,
		vaultService: vaultService,
	}
}

// GetBacklinks gets all notes that link to a target note.
// GetBacklinks 获取链接到目标笔记的所有笔记。
// Uses path variations to match links stored as partial paths (e.g., [[note]], [[folder/note]]).
// 使用路径变体来匹配存储为部分路径的链接（例如 [[note]]，[[folder/note]]）。
func (s *noteLinkService) GetBacklinks(ctx context.Context, uid int64, params *dto.NoteLinkQueryRequest) ([]*dto.NoteLinkItem, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	// Generate all path variations for matching
	// 生成所有用于匹配的路径变体
	// e.g., "projects/folder/note.md" -> ["note", "folder/note", "projects/folder/note"]
	// 例如 "projects/folder/note.md" -> ["note", "folder/note", "projects/folder/note"]
	pathVariations := util.GeneratePathVariations(params.Path)
	if len(pathVariations) == 0 {
		return nil, nil
	}

	// Generate hashes for all variations
	// 为所有变体生成哈希
	var pathHashes []string
	for _, variation := range pathVariations {
		pathHashes = append(pathHashes, util.EncodeHash32(variation))
	}

	// Get backlinks matching any of the path variations
	// 获取匹配任何路径变体的反向链接
	links, err := s.noteLinkRepo.GetBacklinksByHashes(ctx, pathHashes, vaultID, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var results []*dto.NoteLinkItem
	for _, link := range links {
		// Get source note to get its path and content for context
		// 获取源笔记以获取其路径和内容作为上下文
		sourceNote, err := s.noteRepo.GetByID(ctx, link.SourceNoteID, uid)
		if err != nil {
			continue // Skip if note not found / 如果未找到笔记则跳过
		}

		item := &dto.NoteLinkItem{
			Path:     sourceNote.Path,
			LinkText: link.LinkText,
			IsEmbed:  link.IsEmbed,
		}

		// Extract context around the link (try all variations)
		// 提取链接周围的上下文（尝试所有变体）
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
// GetOutlinks 获取源笔记中的所有链接
func (s *noteLinkService) GetOutlinks(ctx context.Context, uid int64, params *dto.NoteLinkQueryRequest) ([]*dto.NoteLinkItem, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// Get note by path to get its ID
	// 通过路径获取笔记以获取其 ID
	note, err := s.noteRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorNoteNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Get outlinks from repository
	// 从存储库获取出站链接
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
		// 提取链接周围的上下文
		item.Context = s.extractLinkContext(note.Content, link.TargetPath)

		results = append(results, item)
	}

	return results, nil
}

// extractLinkContext extracts approximately 50 characters of context around a link
// extractLinkContext 提取链接周围约 50 个字符的上下文
func (s *noteLinkService) extractLinkContext(content, targetPath string) string {
	// Look for [[targetPath]] or [[targetPath|alias]]
	// 查找 [[targetPath]] 或 [[targetPath|alias]]
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
	// 提取上下文：链接前后各 25 个字符
	contextRadius := 25
	start := pos - contextRadius
	if start < 0 {
		start = 0
	}

	// Find the end of the link (closing ]])
	// 查找链接的结尾（闭合的 ]]）
	linkEnd := strings.Index(content[pos:], "]]")
	if linkEnd < 0 {
		linkEnd = matchLen
	} else {
		linkEnd += 2 // Include ]] / 包含 ]]
	}

	end := pos + linkEnd + contextRadius
	if end > len(content) {
		end = len(content)
	}

	context := content[start:end]

	// Clean up: replace newlines with spaces and trim
	// 清理：将换行符替换为空格并修剪
	context = strings.ReplaceAll(context, "\n", " ")
	context = strings.TrimSpace(context)

	// Add ellipsis if truncated
	// 如果被截断则添加省略号
	if start > 0 {
		context = "..." + context
	}
	if end < len(content) {
		context = context + "..."
	}

	return context
}

// Ensure noteLinkService implements NoteLinkService interface
// 确保 noteLinkService 实现了 NoteLinkService 接口
var _ NoteLinkService = (*noteLinkService)(nil)
