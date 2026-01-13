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

	// RestoreFromHistory 从历史版本恢复笔记内容
	RestoreFromHistory(ctx context.Context, uid int64, historyID int64) (*dto.NoteDTO, error)

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
	config       *AppServiceConfig
}

// NewNoteHistoryService 创建 NoteHistoryService 实例
func NewNoteHistoryService(historyRepo domain.NoteHistoryRepository, noteRepo domain.NoteRepository, userRepo domain.UserRepository, vaultSvc VaultService, logger *zap.Logger, config *AppServiceConfig) NoteHistoryService {
	if config == nil {
		config = &AppServiceConfig{HistoryKeepVersions: 100}
	}
	return &noteHistoryService{
		historyRepo:  historyRepo,
		noteRepo:     noteRepo,
		userRepo:     userRepo,
		vaultService: vaultSvc,
		sf:           &singleflight.Group{},
		logger:       logger,
		config:       config,
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
	if err := s.noteRepo.UpdateSnapshot(ctx, note.Content, note.ContentHash, latestVersion+1, note.ID, uid); err != nil {
		return err
	}

	// 检查版本数量限制，超过限制时删除最旧的版本
	return s.cleanupExcessVersions(ctx, noteID, uid)
}

// Migrate 处理笔记历史迁移
func (s *noteHistoryService) Migrate(ctx context.Context, oldNoteID, newNoteID int64, uid int64) error {
	return s.historyRepo.Migrate(ctx, oldNoteID, newNoteID, uid)
}

// RestoreFromHistory 从历史版本恢复笔记内容
// 恢复到该历史版本修改后的内容
func (s *noteHistoryService) RestoreFromHistory(ctx context.Context, uid int64, historyID int64) (*dto.NoteDTO, error) {
	// 1. 获取历史记录
	history, err := s.historyRepo.GetByID(ctx, historyID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorHistoryNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 2. 获取当前笔记
	note, err := s.noteRepo.GetByID(ctx, history.NoteID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorNoteNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 3. 计算该历史版本修改后的内容
	// history.Content 是该版本修改前的快照（即上一版本的内容）
	// history.DiffPatch 是从修改前到修改后的差异补丁
	// 应用补丁得到该版本修改后的完整内容
	dmp := diffmatchpatch.New()
	parsedPatches, _ := dmp.PatchFromText(history.DiffPatch)
	restoredContent, _ := dmp.PatchApply(parsedPatches, history.Content)

	// 4. 计算恢复内容的哈希
	restoredContentHash := util.EncodeHash32(restoredContent)

	// 调试日志
	s.logger.Info("RestoreFromHistory",
		zap.Int64("historyID", historyID),
		zap.Int64("version", history.Version),
		zap.Int("beforeContentLen", len(history.Content)),
		zap.Int("afterContentLen", len(restoredContent)),
	)

	// 5. 使用恢复的内容更新笔记
	note.Content = restoredContent
	note.ContentHash = restoredContentHash
	note.Action = domain.NoteActionModify

	// 6. 更新笔记
	updated, err := s.noteRepo.Update(ctx, note, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 7. 立即保存历史记录（恢复操作不走延迟队列，直接创建历史版本）
	if err := s.ProcessDelay(ctx, updated.ID, uid); err != nil {
		s.logger.Warn("RestoreFromHistory: failed to create history",
			zap.Int64("noteID", updated.ID),
			zap.Error(err))
	}

	// 8. 返回更新后的笔记 DTO
	return &dto.NoteDTO{
		ID:               updated.ID,
		Action:           string(updated.Action),
		Path:             updated.Path,
		PathHash:         updated.PathHash,
		Content:          updated.Content,
		ContentHash:      updated.ContentHash,
		Version:          updated.Version,
		Ctime:            updated.Ctime,
		Mtime:            updated.Mtime,
		UpdatedTimestamp: updated.UpdatedTimestamp,
		UpdatedAt:        timex.Time(updated.UpdatedAt),
		CreatedAt:        timex.Time(updated.CreatedAt),
	}, nil
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

// cleanupExcessVersions 清理超过版本数量限制的历史记录
// 当笔记的历史版本数超过 HistoryKeepVersions 时，删除最旧的版本
func (s *noteHistoryService) cleanupExcessVersions(ctx context.Context, noteID int64, uid int64) error {
	// 获取配置中的版本保留数
	keepVersions := 100 // 默认值
	if s.config != nil && s.config.HistoryKeepVersions > 0 {
		keepVersions = s.config.HistoryKeepVersions
	}

	// 获取该笔记的所有历史版本
	histories, _, err := s.historyRepo.ListByNoteID(ctx, noteID, 1, keepVersions+1, uid)
	if err != nil {
		s.logger.Warn("failed to list note histories for cleanup",
			zap.Int64("noteID", noteID),
			zap.Int64("uid", uid),
			zap.Error(err))
		return nil // 不影响主流程
	}

	// 如果版本数未超过限制，无需清理
	if len(histories) <= keepVersions {
		return nil
	}

	// 删除超出限制的最旧版本
	// histories 已按 Version DESC 排序，所以最后一个是最旧的
	oldestHistory := histories[len(histories)-1]
	if err := s.historyRepo.Delete(ctx, oldestHistory.ID, uid); err != nil {
		s.logger.Warn("failed to delete excess history version",
			zap.Int64("noteID", noteID),
			zap.Int64("historyID", oldestHistory.ID),
			zap.Int64("uid", uid),
			zap.Error(err))
		return nil // 不影响主流程
	}

	return nil
}

// 确保 noteHistoryService 实现了 NoteHistoryService 接口
var _ NoteHistoryService = (*noteHistoryService)(nil)
