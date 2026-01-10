// Package service 实现业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/sergi/go-diff/diffmatchpatch"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

// NoteHistoryService 定义笔记历史业务服务接口
type NoteHistoryService interface {
	// Get 获取指定 ID 的笔记历史详情
	Get(ctx context.Context, uid int64, id int64) (*dto.NoteHistoryDTO, error)

	// GetByNoteIDAndHash 根据笔记ID和内容哈希获取历史记录
	GetByNoteIDAndHash(ctx context.Context, uid int64, noteID int64, contentHash string) (*dto.NoteHistoryDTO, error)

	// List 获取指定笔记的历史版本列表
	List(ctx context.Context, uid int64, params *dto.NoteHistoryListRequest, pager *app.Pager) ([]*dto.NoteHistoryNoContentDTO, int64, error)

	// ProcessDelay 延时处理笔记历史（计算 diff 并保存补丁版本）
	ProcessDelay(ctx context.Context, noteID int64, uid int64) error

	// Migrate 处理笔记历史迁移
	Migrate(ctx context.Context, oldNoteID, newNoteID int64, uid int64) error

	// CleanupByTime 按截止时间清理历史记录，保留每个笔记最近 N 个版本
	CleanupByTime(ctx context.Context, cutoffTime int64, keepVersions int) error
}

// noteHistoryService 实现 NoteHistoryService 接口
type noteHistoryService struct {
	historyRepo  domain.NoteHistoryRepository
	noteRepo     domain.NoteRepository
	userRepo     domain.UserRepository
	vaultService VaultService
	sf           *singleflight.Group
	logger       *zap.Logger
}

// NewNoteHistoryService 创建 NoteHistoryService 实例
func NewNoteHistoryService(historyRepo domain.NoteHistoryRepository, noteRepo domain.NoteRepository, userRepo domain.UserRepository, vaultSvc VaultService, logger *zap.Logger) NoteHistoryService {
	return &noteHistoryService{
		historyRepo:  historyRepo,
		noteRepo:     noteRepo,
		userRepo:     userRepo,
		vaultService: vaultSvc,
		sf:           &singleflight.Group{},
		logger:       logger,
	}
}

// domainToDTO 将领域模型转换为 DTO（包含 diff 计算）
func (s *noteHistoryService) domainToDTO(history *domain.NoteHistory) *dto.NoteHistoryDTO {
	if history == nil {
		return nil
	}

	dmp := diffmatchpatch.New()
	parsedPatches, _ := dmp.PatchFromText(history.DiffPatch)
	restoredNewVersion, _ := dmp.PatchApply(parsedPatches, history.Content)
	diffResults := dmp.DiffMain(history.Content, restoredNewVersion, false)

	return &dto.NoteHistoryDTO{
		ID:          history.ID,
		NoteID:      history.NoteID,
		VaultID:     history.VaultID,
		Path:        history.Path,
		Diffs:       diffResults,
		Content:     history.Content,
		ContentHash: history.ContentHash,
		ClientName:  history.ClientName,
		Version:     history.Version,
		CreatedAt:   timex.Time(history.CreatedAt),
	}
}

// domainToNoContentDTO 将领域模型转换为不含内容的 DTO
func (s *noteHistoryService) domainToNoContentDTO(history *domain.NoteHistory) *dto.NoteHistoryNoContentDTO {
	if history == nil {
		return nil
	}
	return &dto.NoteHistoryNoContentDTO{
		ID:         history.ID,
		NoteID:     history.NoteID,
		VaultID:    history.VaultID,
		Path:       history.Path,
		ClientName: history.ClientName,
		Version:    history.Version,
		CreatedAt:  timex.Time(history.CreatedAt),
	}
}

// Get 获取指定 ID 的笔记历史详情
func (s *noteHistoryService) Get(ctx context.Context, uid int64, id int64) (*dto.NoteHistoryDTO, error) {
	history, err := s.historyRepo.GetByID(ctx, id, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorHistoryNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	return s.domainToDTO(history), nil
}

// GetByNoteIDAndHash 根据笔记ID和内容哈希获取历史记录
func (s *noteHistoryService) GetByNoteIDAndHash(ctx context.Context, uid int64, noteID int64, contentHash string) (*dto.NoteHistoryDTO, error) {
	history, err := s.historyRepo.GetByNoteIDAndHash(ctx, noteID, contentHash, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 记录不存在时返回 nil，调用方会处理
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	return s.domainToDTO(history), nil
}

// List 获取指定笔记的历史版本列表
func (s *noteHistoryService) List(ctx context.Context, uid int64, params *dto.NoteHistoryListRequest, pager *app.Pager) ([]*dto.NoteHistoryNoContentDTO, int64, error) {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, 0, err
	}

	pathHash := params.PathHash
	if pathHash == "" {
		pathHash = util.EncodeHash32(params.Path)
	}

	// 获取笔记 ID
	note, err := s.noteRepo.GetByPathHashIncludeRecycle(ctx, pathHash, vaultID, uid, params.IsRecycle)
	if err != nil {
		return nil, 0, err
	}
	if note == nil {
		return nil, 0, fmt.Errorf("note not found")
	}

	// 获取历史记录列表
	histories, count, err := s.historyRepo.ListByNoteID(ctx, note.ID, pager.Page, pager.PageSize, uid)
	if err != nil {
		return nil, 0, err
	}

	var results []*dto.NoteHistoryNoContentDTO
	for _, h := range histories {
		results = append(results, s.domainToNoContentDTO(h))
	}
	return results, count, nil
}

// ProcessDelay 延时处理笔记历史（计算 diff 并保存补丁版本）
func (s *noteHistoryService) ProcessDelay(ctx context.Context, noteID int64, uid int64) error {
	note, err := s.noteRepo.GetByID(ctx, noteID, uid)
	if err != nil {
		return fmt.Errorf("noteRepo.GetByID: %w", err)
	}

	if note.Content == note.ContentLastSnapshot {
		return nil
	}

	// 计算 diff
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(note.ContentLastSnapshot, note.Content, false)
	patchText := dmp.PatchToText(dmp.PatchMake(note.ContentLastSnapshot, diffs))

	latestVersion, err := s.historyRepo.GetLatestVersion(ctx, note.ID, uid)
	if err != nil {
		return fmt.Errorf("historyRepo.GetLatestVersion: %w", err)
	}

	history := &domain.NoteHistory{
		NoteID:      note.ID,
		VaultID:     note.VaultID,
		Path:        note.Path,
		DiffPatch:   patchText,
		Content:     note.ContentLastSnapshot,
		ContentHash: note.ContentLastSnapshotHash,
		ClientName:  note.ClientName,
		Version:     latestVersion + 1,
		CreatedAt:   note.UpdatedAt,
	}

	_, err = s.historyRepo.Create(ctx, history, uid)
	if err != nil {
		return fmt.Errorf("historyRepo.Create: %w", err)
	}

	// 更新 ContentLastSnapshot
	return s.noteRepo.UpdateSnapshot(ctx, note.Content, note.ContentHash, latestVersion+1, note.ID, uid)
}

// Migrate 处理笔记历史迁移
func (s *noteHistoryService) Migrate(ctx context.Context, oldNoteID, newNoteID int64, uid int64) error {
	return s.historyRepo.Migrate(ctx, oldNoteID, newNoteID, uid)
}

// CleanupByTime 按截止时间清理历史记录，保留每个笔记最近 N 个版本
func (s *noteHistoryService) CleanupByTime(ctx context.Context, cutoffTime int64, keepVersions int) error {
	// 获取所有用户 UID
	uids, err := s.userRepo.GetAllUIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all UIDs: %w", err)
	}

	var totalCleaned int64
	for _, uid := range uids {
		// 获取该用户有旧历史记录的笔记 ID
		noteIDs, err := s.historyRepo.GetNoteIDsWithOldHistory(ctx, cutoffTime, uid)
		if err != nil {
			s.logger.Error("failed to get note IDs with old history",
				zap.Int64("uid", uid),
				zap.Error(err))
			continue
		}

		for _, noteID := range noteIDs {
			// 删除旧版本，保留最近 N 个版本
			if err := s.historyRepo.DeleteOldVersions(ctx, noteID, cutoffTime, keepVersions, uid); err != nil {
				s.logger.Error("failed to cleanup history",
					zap.Int64("uid", uid),
					zap.Int64("noteID", noteID),
					zap.Error(err))
				continue
			}
			totalCleaned++
		}
	}

	s.logger.Info("note history cleanup completed",
		zap.Int64("cutoffTime", cutoffTime),
		zap.Int("keepVersions", keepVersions),
		zap.Int64("notesProcessed", totalCleaned))

	return nil
}

// 确保 noteHistoryService 实现了 NoteHistoryService 接口
var _ NoteHistoryService = (*noteHistoryService)(nil)
