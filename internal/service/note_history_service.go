// Package service implements the business logic layer
// Package service 实现业务逻辑层
package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

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

// NoteHistoryService defines the note history business service interface
// NoteHistoryService 定义笔记历史业务服务接口
type NoteHistoryService interface {
	// Get retrieves note history details for a specified ID
	// Get 获取指定 ID 的笔记历史详情
	Get(ctx context.Context, uid int64, id int64) (*dto.NoteHistoryDTO, error)

	// GetByNoteIDAndHash retrieves history record by note ID and content hash
	// GetByNoteIDAndHash 根据笔记 ID 和内容哈希获取历史记录
	GetByNoteIDAndHash(ctx context.Context, uid int64, noteID int64, contentHash string) (*dto.NoteHistoryDTO, error)

	// List retrieves history version list for a specified note
	// List 获取指定笔记的历史版本列表
	List(ctx context.Context, uid int64, params *dto.NoteHistoryListRequest, pager *app.Pager) ([]*dto.NoteHistoryNoContentDTO, int64, error)

	// RestoreFromHistory restores note content from a history version
	// RestoreFromHistory 从历史版本恢复笔记内容
	RestoreFromHistory(ctx context.Context, uid int64, historyID int64) (*dto.NoteDTO, error)

	// ProcessDelay processes note history with delay (calculates diff and saves patch version)
	// ProcessDelay 延时处理笔记历史（计算 diff 并保存补丁版本）
	ProcessDelay(ctx context.Context, noteID int64, uid int64) error

	// Migrate handles note history migration
	// Migrate 处理笔记历史迁移
	Migrate(ctx context.Context, oldNoteID, newNoteID int64, uid int64) error

	// CleanupByTime cleans up history records by cutoff time, keeping recent N versions per note
	// CleanupByTime 按截止时间清理历史记录，保留每个笔记最近 N 个版本
	CleanupByTime(ctx context.Context, cutoffTime int64, keepVersions int) error
}

// noteHistoryService implementation of NoteHistoryService interface
// noteHistoryService 实现 NoteHistoryService 接口
type noteHistoryService struct {
	historyRepo   domain.NoteHistoryRepository // History repository // 历史记录仓库
	noteRepo      domain.NoteRepository        // Note repository // 笔记仓库
	userRepo      domain.UserRepository        // User repository // 用户仓库
	vaultService  VaultService                 // Vault service // 仓库服务
	folderService FolderService                // Folder service // 文件夹服务
	noteService   NoteService                  // Note service // 笔记服务
	sf            *singleflight.Group          // Singleflight group // 并发请求合并组
	logger        *zap.Logger                  // Logger // 日志对象
	config        *AppServiceConfig            // Service configuration // 服务配置
}

// NewNoteHistoryService creates NoteHistoryService instance
// NewNoteHistoryService 创建 NoteHistoryService 实例
func NewNoteHistoryService(historyRepo domain.NoteHistoryRepository, noteRepo domain.NoteRepository, userRepo domain.UserRepository, vaultSvc VaultService, folderSvc FolderService, noteSvc NoteService, logger *zap.Logger, config *AppServiceConfig) NoteHistoryService {
	if config == nil {
		config = &AppServiceConfig{HistoryKeepVersions: 100}
	}
	return &noteHistoryService{
		historyRepo:   historyRepo,
		noteRepo:      noteRepo,
		userRepo:      userRepo,
		vaultService:  vaultSvc,
		folderService: folderSvc,
		noteService:   noteSvc,
		sf:            &singleflight.Group{},
		logger:        logger,
		config:        config,
	}
}

// domainToDTO converts domain model to DTO (includes diff calculation)
// domainToDTO 将领域模型转换为 DTO（包含 diff 计算）
func (s *noteHistoryService) domainToDTO(history *domain.NoteHistory) *dto.NoteHistoryDTO {
	if history == nil {
		return nil
	}

	dmp := diffmatchpatch.New()

	// Add recover protection as a second line of defense
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Panic recovered in domainToDTO",
				zap.Any("panic", r),
				zap.Int64("historyID", history.ID))
		}
	}()

	content := s.ensureValidUTF8(history.Content)
	diffPatch := s.ensureValidUTF8(history.DiffPatch)

	parsedPatches, _ := dmp.PatchFromText(diffPatch)
	restoredNewVersion, _ := dmp.PatchApply(parsedPatches, content)
	diffResults := dmp.DiffMain(content, restoredNewVersion, false)

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

// domainToNoContentDTO converts domain model to DTO without content
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

// Get retrieves note history details for a specified ID
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

// GetByNoteIDAndHash retrieves history record by note ID and content hash
// GetByNoteIDAndHash 根据笔记 ID 和内容哈希获取历史记录
func (s *noteHistoryService) GetByNoteIDAndHash(ctx context.Context, uid int64, noteID int64, contentHash string) (*dto.NoteHistoryDTO, error) {
	history, err := s.historyRepo.GetByNoteIDAndHash(ctx, noteID, contentHash, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil when record not found, caller will handle // 记录不存在时返回 nil，调用方会处理
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	return s.domainToDTO(history), nil
}

// List retrieves history version list for a specified note
// List 获取指定笔记的历史版本列表
func (s *noteHistoryService) List(ctx context.Context, uid int64, params *dto.NoteHistoryListRequest, pager *app.Pager) ([]*dto.NoteHistoryNoContentDTO, int64, error) {
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, 0, err
	}

	pathHash := params.PathHash
	if pathHash == "" {
		pathHash = util.EncodeHash32(params.Path)
	}

	// Get note ID
	// 获取笔记 ID
	note, err := s.noteRepo.GetByPathHashIncludeRecycle(ctx, pathHash, vaultID, uid, params.IsRecycle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, code.ErrorNoteNotFound
		}
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}
	if note == nil {
		return nil, 0, code.ErrorNoteNotFound
	}

	// Get history record list
	// 获取历史记录列表
	histories, count, err := s.historyRepo.ListByNoteID(ctx, note.ID, pager.Page, pager.PageSize, uid)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var results []*dto.NoteHistoryNoContentDTO
	for _, h := range histories {
		results = append(results, s.domainToNoContentDTO(h))
	}
	return results, count, nil
}

// ProcessDelay processes note history with delay (calculates diff and saves patch version)
// ProcessDelay 延时处理笔记历史（计算 diff 并保存补丁版本）
func (s *noteHistoryService) ProcessDelay(ctx context.Context, noteID int64, uid int64) error {
	note, err := s.noteRepo.GetByID(ctx, noteID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.ErrorNoteNotFound
		}
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	if note.Content == note.ContentLastSnapshot {
		return nil
	}

	// Calculate diff
	// 计算 diff
	dmp := diffmatchpatch.New()

	// Add recover protection
	var patchText string
	func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("Panic recovered in ProcessDelay during diff calculation",
					zap.Any("panic", r),
					zap.Int64("noteID", noteID))
			}
		}()

		content1 := s.ensureValidUTF8(note.ContentLastSnapshot)
		content2 := s.ensureValidUTF8(note.Content)

		diffs := dmp.DiffMain(content1, content2, false)
		patchText = dmp.PatchToText(dmp.PatchMake(content1, diffs))
	}()

	if patchText == "" && note.Content != note.ContentLastSnapshot {
		// If patch calculation failed (due to panic and recover), we don't proceed with history creation
		// to avoid saving corrupted data.
		return nil
	}

	latestVersion, err := s.historyRepo.GetLatestVersion(ctx, note.ID, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
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
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Update ContentLastSnapshot
	// 更新 ContentLastSnapshot
	if err := s.noteRepo.UpdateSnapshot(ctx, note.Content, note.ContentHash, latestVersion+1, note.ID, uid); err != nil {
		return err
	}

	// Check version count limit, delete oldest version when exceeding limit
	// 检查版本数量限制，超过限制时删除最旧的版本
	return s.cleanupExcessVersions(ctx, noteID, uid)
}

// Migrate handles note history migration
// Migrate 处理笔记历史迁移
func (s *noteHistoryService) Migrate(ctx context.Context, oldNoteID, newNoteID int64, uid int64) error {
	return s.historyRepo.Migrate(ctx, oldNoteID, newNoteID, uid)
}

// RestoreFromHistory restores note content from a history version
// Restores to the content after the modification of this history version
// RestoreFromHistory 从历史版本恢复笔记内容
// 恢复到该历史版本修改后的内容
func (s *noteHistoryService) RestoreFromHistory(ctx context.Context, uid int64, historyID int64) (*dto.NoteDTO, error) {
	// 1. Get history record
	// 1. 获取历史记录
	history, err := s.historyRepo.GetByID(ctx, historyID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorHistoryNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 2. Get current note
	// 2. 获取当前笔记
	note, err := s.noteRepo.GetByID(ctx, history.NoteID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorNoteNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 3. Calculate content after this history version modification
	// history.Content is the snapshot before this version modification (i.e., content of the previous version)
	// history.DiffPatch is the difference patch from before modification to after modification
	// Apply patch to get the complete content after this version modification
	// 3. 计算该历史版本修改后的内容
	// history.Content 是该版本修改前的快照（即上一版本的内容）
	// history.DiffPatch 是从修改前到修改后的差异补丁
	// 应用补丁得到该版本修改后的完整内容
	dmp := diffmatchpatch.New()

	// Add recover protection
	var restoredContent string
	func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("Panic recovered in RestoreFromHistory during patch application",
					zap.Any("panic", r),
					zap.Int64("historyID", historyID))
			}
		}()

		historyContent := s.ensureValidUTF8(history.Content)
		diffPatch := s.ensureValidUTF8(history.DiffPatch)

		parsedPatches, _ := dmp.PatchFromText(diffPatch)
		restoredContent, _ = dmp.PatchApply(parsedPatches, historyContent)
	}()

	if restoredContent == "" {
		return nil, code.ErrorHistoryNotFound.WithDetails("failed to restore content from history due to internal error")
	}

	// 4. Calculate hash of restored content
	// 4. 计算恢复内容的哈希
	restoredContentHash := util.EncodeHash32(restoredContent)

	// Debug log
	// 调试日志
	s.logger.Info("RestoreFromHistory",
		zap.Int64("historyID", historyID),
		zap.Int64("version", history.Version),
		zap.Int("beforeContentLen", len(history.Content)),
		zap.Int("afterContentLen", len(restoredContent)),
	)

	// 5. Update note with restored content and set modification time
	// 5. 使用恢复的内容更新笔记, 并设置修改时间
	note.Content = restoredContent
	note.ContentHash = restoredContentHash
	note.Mtime = timex.Now().UnixMilli()
	note.Action = domain.NoteActionModify

	// 6. Update note
	// 6. 更新笔记
	updated, err := s.noteRepo.Update(ctx, note, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	vaultID := history.VaultID
	go s.folderService.SyncResourceFID(context.Background(), uid, vaultID, []int64{updated.ID}, nil)
	go s.noteService.CountSizeSum(context.Background(), vaultID, uid)
	go s.noteService.UpdateNoteLinks(context.Background(), updated.ID, updated.Content, vaultID, uid)

	NoteHistoryDelayPush(updated.ID, uid)

	// 7. Save history record immediately (restore operation does not use delay queue, direct history version creation)
	// 7. 立即保存历史记录（恢复操作不走延迟队列，直接创建历史版本）
	if err := s.ProcessDelay(ctx, updated.ID, uid); err != nil {
		s.logger.Warn("RestoreFromHistory: failed to create history",
			zap.Int64("noteID", updated.ID),
			zap.Error(err))
	}

	// 8. Return updated note DTO
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

// CleanupByTime cleans up history records by cutoff time, keeping recent N versions per note
// CleanupByTime 按截止时间清理历史记录，保留每个笔记最近 N 个版本
func (s *noteHistoryService) CleanupByTime(ctx context.Context, cutoffTime int64, keepVersions int) error {
	// Get all user UIDs
	// 获取所有用户 UID
	uids, err := s.userRepo.GetAllUIDs(ctx)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	var totalCleaned int64
	for i, uid := range uids {
		// 增加错峰延迟，避免瞬间触发大量写事务
		if i > 0 {
			time.Sleep(500 * time.Millisecond)
		}
		// Get note IDs with old history records for this user
		// 获取该用户有旧历史记录的笔记 ID
		noteIDs, err := s.historyRepo.GetNoteIDsWithOldHistory(ctx, cutoffTime, uid)
		if err != nil {
			s.logger.Error("failed to get note IDs with old history",
				zap.Int64("uid", uid),
				zap.Error(err))
			continue
		}

		for _, noteID := range noteIDs {
			// Delete old versions, keep recent N versions
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

// cleanupExcessVersions cleans up history records exceeding version count limit
// Delete oldest version when note history versions exceed HistoryKeepVersions
// cleanupExcessVersions 清理超过版本数量限制的历史记录
// 当笔记的历史版本数超过 HistoryKeepVersions 时，删除最旧的版本
func (s *noteHistoryService) cleanupExcessVersions(ctx context.Context, noteID int64, uid int64) error {
	// Get version retention count from configuration
	// 获取配置中的版本保留数
	keepVersions := 100 // Default value // 默认值
	if s.config != nil && s.config.HistoryKeepVersions > 0 {
		keepVersions = s.config.HistoryKeepVersions
	}

	// Get all history versions for this note
	// 获取该笔记的所有历史版本
	histories, _, err := s.historyRepo.ListByNoteID(ctx, noteID, 1, keepVersions+1, uid)
	if err != nil {
		s.logger.Warn("failed to list note histories for cleanup",
			zap.Int64("noteID", noteID),
			zap.Int64("uid", uid),
			zap.Error(err))
		return nil // Does not affect main flow // 不影响主流程
	}

	// No cleanup needed if version count does not exceed limit
	// 如果版本数未超过限制，无需清理
	if len(histories) <= keepVersions {
		return nil
	}

	// Delete oldest version exceeding limit
	// histories are sorted by Version DESC, so the last one is the oldest
	// 删除超出限制的最旧版本
	// histories 已按 Version DESC 排序，所以最后一个是最旧的
	oldestHistory := histories[len(histories)-1]
	if err := s.historyRepo.Delete(ctx, oldestHistory.ID, uid); err != nil {
		s.logger.Warn("failed to delete excess history version",
			zap.Int64("noteID", noteID),
			zap.Int64("historyID", oldestHistory.ID),
			zap.Int64("uid", uid),
			zap.Error(err))
		return nil // Does not affect main flow // 不影响主流程
	}

	return nil
}

// ensureValidUTF8 ensures the string is valid UTF-8
// ensureValidUTF8 确保字符串是有效的 UTF-8 编码
func (s *noteHistoryService) ensureValidUTF8(str string) string {
	if utf8.ValidString(str) {
		return str
	}
	return strings.ToValidUTF8(str, "")
}

// Verify noteHistoryService implements NoteHistoryService interface
// 确保 noteHistoryService 实现了 NoteHistoryService 接口
var _ NoteHistoryService = (*noteHistoryService)(nil)
