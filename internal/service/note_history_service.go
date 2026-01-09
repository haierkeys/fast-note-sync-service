// Package service 实现业务逻辑层
package service

import (
	"context"
	"fmt"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"github.com/sergi/go-diff/diffmatchpatch"
	"golang.org/x/sync/singleflight"
)

// NoteHistoryService 定义笔记历史业务服务接口
type NoteHistoryService interface {
	// Get 获取指定 ID 的笔记历史详情
	Get(ctx context.Context, uid int64, id int64) (*NoteHistoryDTO, error)

	// GetByNoteIDAndHash 根据笔记ID和内容哈希获取历史记录
	GetByNoteIDAndHash(ctx context.Context, uid int64, noteID int64, contentHash string) (*NoteHistoryDTO, error)

	// List 获取指定笔记的历史版本列表
	List(ctx context.Context, uid int64, params *dto.NoteHistoryListRequest, pager *app.Pager) ([]*NoteHistoryNoContentDTO, int64, error)

	// ProcessDelay 延时处理笔记历史（计算 diff 并保存补丁版本）
	ProcessDelay(ctx context.Context, noteID int64, uid int64) error

	// Migrate 处理笔记历史迁移
	Migrate(ctx context.Context, oldNoteID, newNoteID int64, uid int64) error
}

// NoteHistoryDTO 笔记历史数据传输对象
type NoteHistoryDTO struct {
	ID          int64                 `json:"id" form:"id"`
	NoteID      int64                 `json:"noteId" form:"noteId"`
	VaultID     int64                 `json:"vaultId" form:"vaultId"`
	Path        string                `json:"path" form:"path"`
	Diffs       []diffmatchpatch.Diff `json:"diffs"`
	Content     string                `json:"content" form:"content"`
	ContentHash string                `json:"contentHash" form:"contentHash"`
	ClientName  string                `json:"clientName" form:"clientName"`
	Version     int64                 `json:"version" form:"version"`
	CreatedAt   timex.Time            `json:"createdAt" form:"createdAt"`
}

// NoteHistoryNoContentDTO 不包含内容的笔记历史 DTO
type NoteHistoryNoContentDTO struct {
	ID         int64      `json:"id" form:"id"`
	NoteID     int64      `json:"noteId" form:"noteId"`
	VaultID    int64      `json:"vaultId" form:"vaultId"`
	Path       string     `json:"path" form:"path"`
	ClientName string     `json:"clientName" form:"clientName"`
	Version    int64      `json:"version" form:"version"`
	CreatedAt  timex.Time `json:"createdAt" form:"createdAt"`
}

// noteHistoryService 实现 NoteHistoryService 接口
type noteHistoryService struct {
	historyRepo  domain.NoteHistoryRepository
	noteRepo     domain.NoteRepository
	vaultService VaultService
	sf           *singleflight.Group
}

// NewNoteHistoryService 创建 NoteHistoryService 实例
func NewNoteHistoryService(historyRepo domain.NoteHistoryRepository, noteRepo domain.NoteRepository, vaultSvc VaultService) NoteHistoryService {
	return &noteHistoryService{
		historyRepo:  historyRepo,
		noteRepo:     noteRepo,
		vaultService: vaultSvc,
		sf:           &singleflight.Group{},
	}
}

// domainToDTO 将领域模型转换为 DTO（包含 diff 计算）
func (s *noteHistoryService) domainToDTO(history *domain.NoteHistory) *NoteHistoryDTO {
	if history == nil {
		return nil
	}

	dmp := diffmatchpatch.New()
	parsedPatches, _ := dmp.PatchFromText(history.DiffPatch)
	restoredNewVersion, _ := dmp.PatchApply(parsedPatches, history.Content)
	diffResults := dmp.DiffMain(history.Content, restoredNewVersion, false)

	return &NoteHistoryDTO{
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
func (s *noteHistoryService) domainToNoContentDTO(history *domain.NoteHistory) *NoteHistoryNoContentDTO {
	if history == nil {
		return nil
	}
	return &NoteHistoryNoContentDTO{
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
func (s *noteHistoryService) Get(ctx context.Context, uid int64, id int64) (*NoteHistoryDTO, error) {
	history, err := s.historyRepo.GetByID(ctx, id, uid)
	if err != nil {
		return nil, err
	}
	return s.domainToDTO(history), nil
}

// GetByNoteIDAndHash 根据笔记ID和内容哈希获取历史记录
func (s *noteHistoryService) GetByNoteIDAndHash(ctx context.Context, uid int64, noteID int64, contentHash string) (*NoteHistoryDTO, error) {
	history, err := s.historyRepo.GetByNoteIDAndHash(ctx, noteID, contentHash, uid)
	if err != nil {
		return nil, err
	}
	if history == nil {
		return nil, nil
	}
	return s.domainToDTO(history), nil
}

// List 获取指定笔记的历史版本列表
func (s *noteHistoryService) List(ctx context.Context, uid int64, params *dto.NoteHistoryListRequest, pager *app.Pager) ([]*NoteHistoryNoContentDTO, int64, error) {
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

	var results []*NoteHistoryNoContentDTO
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

// 确保 noteHistoryService 实现了 NoteHistoryService 接口
var _ NoteHistoryService = (*noteHistoryService)(nil)
