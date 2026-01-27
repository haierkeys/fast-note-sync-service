// Package service 实现业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

// NoteService 定义笔记业务服务接口
type NoteService interface {
	// Get 获取单条笔记
	Get(ctx context.Context, uid int64, params *dto.NoteGetRequest) (*dto.NoteDTO, error)

	// UpdateCheck 检查笔记是否需要更新
	UpdateCheck(ctx context.Context, uid int64, params *dto.NoteUpdateCheckRequest) (string, *dto.NoteDTO, error)

	// ModifyOrCreate 创建或修改笔记
	ModifyOrCreate(ctx context.Context, uid int64, params *dto.NoteModifyOrCreateRequest, mtimeCheck bool) (bool, *dto.NoteDTO, error)

	// Delete 删除笔记
	Delete(ctx context.Context, uid int64, params *dto.NoteDeleteRequest) (*dto.NoteDTO, error)

	// Restore 恢复笔记（从回收站恢复）
	Restore(ctx context.Context, uid int64, params *dto.NoteRestoreRequest) (*dto.NoteDTO, error)

	// Rename 重命名笔记
	Rename(ctx context.Context, uid int64, params *dto.NoteRenameRequest) error

	// List 获取笔记列表
	List(ctx context.Context, uid int64, params *dto.NoteListRequest, pager *app.Pager) ([]*dto.NoteNoContentDTO, int, error)

	// ListByLastTime 获取在 lastTime 之后更新的笔记
	ListByLastTime(ctx context.Context, uid int64, params *dto.NoteSyncRequest) ([]*dto.NoteDTO, error)

	// Sync 同步笔记（ListByLastTime 的别名，用于 WebSocket 同步）
	Sync(ctx context.Context, uid int64, params *dto.NoteSyncRequest) ([]*dto.NoteDTO, error)

	// CountSizeSum 统计 vault 中笔记总数与总大小
	CountSizeSum(ctx context.Context, vaultID int64, uid int64) error

	// Cleanup 清理过期的软删除笔记
	Cleanup(ctx context.Context, uid int64) error

	// CleanupByTime 按截止时间清理所有用户的过期软删除笔记
	CleanupByTime(ctx context.Context, cutoffTime int64) error

	// ListNeedSnapshot 获取需要快照的笔记
	ListNeedSnapshot(ctx context.Context, uid int64) ([]*dto.NoteDTO, error)

	// Migrate 迁移笔记历史记录
	Migrate(ctx context.Context, oldNoteID, newNoteID int64, uid int64) error

	// MigratePush submits note migration task
	MigratePush(oldNoteID, newNoteID int64, uid int64)

	// WithClient sets client info
	WithClient(name, version string) NoteService

	// PatchFrontmatter patches note frontmatter
	PatchFrontmatter(ctx context.Context, uid int64, params *dto.NotePatchFrontmatterRequest) (*dto.NoteDTO, error)

	// AppendContent appends content to a note
	AppendContent(ctx context.Context, uid int64, params *dto.NoteAppendRequest) (*dto.NoteDTO, error)

	// PrependContent prepends content to a note
	PrependContent(ctx context.Context, uid int64, params *dto.NotePrependRequest) (*dto.NoteDTO, error)

	// ReplaceContent performs find/replace in a note
	ReplaceContent(ctx context.Context, uid int64, params *dto.NoteReplaceRequest) (*dto.NoteReplaceResponse, error)

	// Move moves a note to a new path
	Move(ctx context.Context, uid int64, params *dto.NoteMoveRequest) (*dto.NoteDTO, error)
}

// noteService implements NoteService interface
type noteService struct {
	noteRepo     domain.NoteRepository
	noteLinkRepo domain.NoteLinkRepository
	fileRepo     domain.FileRepository
	vaultService VaultService
	sf           *singleflight.Group
	clientName   string
	clientVer    string
	config       *ServiceConfig
}

// NewNoteService creates NoteService instance
func NewNoteService(noteRepo domain.NoteRepository, noteLinkRepo domain.NoteLinkRepository, fileRepo domain.FileRepository, vaultSvc VaultService, config *ServiceConfig) NoteService {
	return &noteService{
		noteRepo:     noteRepo,
		noteLinkRepo: noteLinkRepo,
		fileRepo:     fileRepo,
		vaultService: vaultSvc,
		sf:           &singleflight.Group{},
		config:       config,
	}
}

// WithClient sets client info, returns new NoteService instance
func (s *noteService) WithClient(name, version string) NoteService {
	return &noteService{
		noteRepo:     s.noteRepo,
		noteLinkRepo: s.noteLinkRepo,
		fileRepo:     s.fileRepo,
		vaultService: s.vaultService,
		sf:           s.sf,
		clientName:   name,
		clientVer:    version,
		config:       s.config,
	}
}

// domainToDTO 将领域模型转换为 DTO
func (s *noteService) domainToDTO(note *domain.Note) *dto.NoteDTO {
	if note == nil {
		return nil
	}
	return &dto.NoteDTO{
		ID:               note.ID,
		Action:           string(note.Action),
		Path:             note.Path,
		PathHash:         note.PathHash,
		Content:          note.Content,
		ContentHash:      note.ContentHash,
		Version:          note.Version,
		Ctime:            note.Ctime,
		Mtime:            note.Mtime,
		UpdatedTimestamp: note.UpdatedTimestamp,
		UpdatedAt:        timex.Time(note.UpdatedAt),
		CreatedAt:        timex.Time(note.CreatedAt),
	}
}

// domainToNoContentDTO 将领域模型转换为不含内容的 DTO
func (s *noteService) domainToNoContentDTO(note *domain.Note) *dto.NoteNoContentDTO {
	if note == nil {
		return nil
	}
	return &dto.NoteNoContentDTO{
		ID:               note.ID,
		Action:           string(note.Action),
		Path:             note.Path,
		PathHash:         note.PathHash,
		Version:          note.Version,
		Ctime:            note.Ctime,
		Mtime:            note.Mtime,
		UpdatedTimestamp: note.UpdatedTimestamp,
		UpdatedAt:        timex.Time(note.UpdatedAt),
		CreatedAt:        timex.Time(note.CreatedAt),
	}
}

// Get 获取单条笔记
func (s *noteService) Get(ctx context.Context, uid int64, params *dto.NoteGetRequest) (*dto.NoteDTO, error) {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	note, err := s.noteRepo.GetByPathHashIncludeRecycle(ctx, params.PathHash, vaultID, uid, params.IsRecycle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorNoteNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	return s.domainToDTO(note), nil
}

// UpdateCheck 检查笔记是否需要更新
func (s *noteService) UpdateCheck(ctx context.Context, uid int64, params *dto.NoteUpdateCheckRequest) (string, *dto.NoteDTO, error) {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return "", nil, err
	}

	note, _ := s.noteRepo.GetAllByPathHash(ctx, params.PathHash, vaultID, uid)

	if note != nil {
		noteDTO := s.domainToDTO(note)
		// 检查内容是否一致
		if note.Action == "delete" {
			return "Create", nil, nil
		}
		if note.ContentHash == params.ContentHash {
			// 当用户 mtime 小于服务端 mtime 时，通知用户更新 mtime
			if params.Mtime < note.Mtime {
				return "UpdateMtime", noteDTO, nil
			} else if params.Mtime > note.Mtime {
				if err := s.noteRepo.UpdateMtime(ctx, params.Mtime, note.ID, uid); err != nil {
					// 非关键更新失败，记录警告日志但不阻断流程
					zap.L().Warn("UpdateMtime failed for note",
						zap.Int64(logger.FieldUID, uid),
						zap.Int64("noteId", note.ID),
						zap.Int64("mtime", params.Mtime),
						zap.String(logger.FieldMethod, "NoteService.UpdateCheck"),
						zap.Error(err),
					)
				}
			}
			return "", noteDTO, nil
		}
		return "UpdateContent", noteDTO, nil
	}
	return "Create", nil, nil
}

// ModifyOrCreate 创建或修改笔记
func (s *noteService) ModifyOrCreate(ctx context.Context, uid int64, params *dto.NoteModifyOrCreateRequest, mtimeCheck bool) (bool, *dto.NoteDTO, error) {
	var isNew bool

	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return isNew, nil, err
	}

	note, _ := s.noteRepo.GetAllByPathHash(ctx, params.PathHash, vaultID, uid)

	if note != nil {
		isNew = false

		// If createOnly is set and note exists (not deleted), return error
		if note.Action != domain.NoteActionDelete && params.CreateOnly {
			return false, nil, code.ErrorNoteExist
		}

		// Check if content is consistent, excluding notes marked as deleted
		if mtimeCheck && note.Action != domain.NoteActionDelete && note.Mtime == params.Mtime && note.ContentHash == params.ContentHash {
			return isNew, nil, nil
		}
		// 检查内容是否一致但修改时间不同，则只更新修改时间
		if mtimeCheck && note.Mtime < params.Mtime && note.ContentHash == params.ContentHash {
			err := s.noteRepo.UpdateActionMtime(ctx, domain.NoteActionModify, params.Mtime, note.ID, uid)
			if err != nil {
				return isNew, nil, code.ErrorDBQuery.WithDetails(err.Error())
			}
			note.Mtime = params.Mtime
			return isNew, s.domainToDTO(note), nil
		}

		// 设置 action
		var action domain.NoteAction
		if note.Action == domain.NoteActionDelete {
			action = domain.NoteActionCreate
		} else {
			action = domain.NoteActionModify
		}

		// 更新笔记
		note.VaultID = vaultID
		note.Path = params.Path
		note.PathHash = params.PathHash
		note.Content = params.Content
		note.ContentHash = params.ContentHash
		note.ClientName = s.clientName
		note.Size = int64(len(params.Content))
		note.Mtime = params.Mtime
		note.Ctime = params.Ctime
		note.Action = action
		note.Rename = 0
		note.Version++ // 内容变更时递增版本号

		updated, err := s.noteRepo.Update(ctx, note, uid)
		if err != nil {
			return isNew, nil, code.ErrorDBQuery.WithDetails(err.Error())
		}

		go s.CountSizeSum(context.Background(), vaultID, uid)
		go s.updateNoteLinks(context.Background(), updated.ID, params.Content, vaultID, uid)
		NoteHistoryDelayPush(updated.ID, uid)

		return isNew, s.domainToDTO(updated), nil
	}

	// Create new note
	isNew = true
	newNote := &domain.Note{
		VaultID:     vaultID,
		Path:        params.Path,
		PathHash:    params.PathHash,
		Content:     params.Content,
		ContentHash: params.ContentHash,
		ClientName:  s.clientName,
		Size:        int64(len(params.Content)),
		Mtime:       params.Mtime,
		Ctime:       params.Ctime,
		Action:      domain.NoteActionCreate,
	}

	created, err := s.noteRepo.Create(ctx, newNote, uid)
	if err != nil {
		return isNew, nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	go s.CountSizeSum(context.Background(), vaultID, uid)
	go s.updateNoteLinks(context.Background(), created.ID, params.Content, vaultID, uid)
	NoteHistoryDelayPush(created.ID, uid)

	return isNew, s.domainToDTO(created), nil
}

// Delete deletes a note
func (s *noteService) Delete(ctx context.Context, uid int64, params *dto.NoteDeleteRequest) (*dto.NoteDTO, error) {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err // VaultService 已返回 code.Error
	}

	note, err := s.noteRepo.GetByPathHashIncludeRecycle(ctx, params.PathHash, vaultID, uid, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorNoteNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 更新为删除状态
	note.Action = domain.NoteActionDelete
	note.ClientName = s.clientName
	note.Rename = 0

	err = s.noteRepo.UpdateDelete(ctx, note, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 重新获取更新后的笔记
	updated, err := s.noteRepo.GetByID(ctx, note.ID, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	go s.CountSizeSum(context.Background(), vaultID, uid)
	NoteHistoryDelayPush(updated.ID, uid)

	return s.domainToDTO(updated), nil
}

// Restore 恢复笔记（从回收站恢复）
func (s *noteService) Restore(ctx context.Context, uid int64, params *dto.NoteRestoreRequest) (*dto.NoteDTO, error) {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err // VaultService 已返回 code.Error
	}

	// 从回收站获取笔记
	note, err := s.noteRepo.GetByPathHashIncludeRecycle(ctx, params.PathHash, vaultID, uid, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorNoteNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 检查笔记是否已删除
	if note.Action != domain.NoteActionDelete {
		return nil, code.ErrorNoteNotFound
	}

	// 更新为修改状态 并更新修改时间
	note.Action = domain.NoteActionModify
	note.ClientName = s.clientName
	note.Mtime = time.Now().UnixMilli()
	note.Rename = 0

	err = s.noteRepo.UpdateDelete(ctx, note, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 重新获取更新后的笔记
	updated, err := s.noteRepo.GetByID(ctx, note.ID, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	go s.CountSizeSum(context.Background(), vaultID, uid)
	NoteHistoryDelayPush(updated.ID, uid)

	return s.domainToDTO(updated), nil
}

// Rename 重命名笔记
func (s *noteService) Rename(ctx context.Context, uid int64, params *dto.NoteRenameRequest) error {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return err
	}

	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}
	if params.OldPathHash == "" {
		params.OldPathHash = util.EncodeHash32(params.OldPath)
	}

	// 获取旧笔记
	oldNote, err := s.noteRepo.GetAllByPathHash(ctx, params.OldPathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.ErrorNoteNotFound
		}
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 获取新笔记
	newNote, err := s.noteRepo.GetAllByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.ErrorNoteNotFound
		}
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 触发历史记录迁移
	s.MigratePush(oldNote.ID, newNote.ID, uid)

	return nil
}

// List 获取笔记列表
func (s *noteService) List(ctx context.Context, uid int64, params *dto.NoteListRequest, pager *app.Pager) ([]*dto.NoteNoContentDTO, int, error) {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, 0, err
	}

	notes, err := s.noteRepo.List(ctx, vaultID, pager.Page, pager.PageSize, uid, params.Keyword, params.IsRecycle, params.SearchMode, params.SearchContent, params.SortBy, params.SortOrder)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}

	count, err := s.noteRepo.ListCount(ctx, vaultID, uid, params.Keyword, params.IsRecycle, params.SearchMode, params.SearchContent)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var result []*dto.NoteNoContentDTO
	for _, n := range notes {
		result = append(result, s.domainToNoContentDTO(n))
	}

	return result, int(count), nil
}

// ListByLastTime 获取在 lastTime 之后更新的笔记
func (s *noteService) ListByLastTime(ctx context.Context, uid int64, params *dto.NoteSyncRequest) ([]*dto.NoteDTO, error) {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err // VaultService 已返回 code.Error
	}

	notes, err := s.noteRepo.ListByUpdatedTimestamp(ctx, params.LastTime, vaultID, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var results []*dto.NoteDTO
	cacheList := make(map[string]bool)
	for _, note := range notes {
		if cacheList[note.PathHash] {
			continue
		}
		results = append(results, s.domainToDTO(note))
		cacheList[note.PathHash] = true
	}

	return results, nil
}

// CountSizeSum 统计 vault 中笔记总数与总大小
func (s *noteService) CountSizeSum(ctx context.Context, vaultID int64, uid int64) error {
	key := fmt.Sprintf("note_count_size_sum_%d_%d", uid, vaultID)
	_, err, _ := s.sf.Do(key, func() (any, error) {
		result, err := s.noteRepo.CountSizeSum(ctx, vaultID, uid)
		if err != nil {
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}
		return nil, s.vaultService.UpdateNoteStats(ctx, result.Size, result.Count, vaultID, uid)
	})
	return err
}

// Cleanup 清理过期的软删除笔记
func (s *noteService) Cleanup(ctx context.Context, uid int64) error {
	if s.config == nil {
		return nil
	}
	retentionTimeStr := s.config.App.SoftDeleteRetentionTime
	if retentionTimeStr == "" || retentionTimeStr == "0" {
		return nil
	}

	retentionDuration, err := util.ParseDuration(retentionTimeStr)
	if err != nil {
		return code.ErrorInvalidParams.WithDetails("invalid SoftDeleteRetentionTime")
	}

	if retentionDuration <= 0 {
		return nil
	}

	cutoffTime := time.Now().Add(-retentionDuration).UnixMilli()
	err = s.noteRepo.DeletePhysicalByTime(ctx, cutoffTime, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}
	return nil
}

// CleanupByTime 按截止时间清理所有用户的过期软删除笔记
func (s *noteService) CleanupByTime(ctx context.Context, cutoffTime int64) error {
	return s.noteRepo.DeletePhysicalByTimeAll(ctx, cutoffTime)
}

// ListNeedSnapshot 获取需要快照的笔记
func (s *noteService) ListNeedSnapshot(ctx context.Context, uid int64) ([]*dto.NoteDTO, error) {
	list, err := s.noteRepo.ListContentUnchanged(ctx, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var result []*dto.NoteDTO
	for _, n := range list {
		result = append(result, s.domainToDTO(n))
	}
	return result, nil
}

// Migrate 迁移笔记历史记录
func (s *noteService) Migrate(ctx context.Context, oldNoteID, newNoteID int64, uid int64) error {
	// 获取旧笔记信息
	oldNote, err := s.noteRepo.GetByID(ctx, oldNoteID, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 将旧笔记的 ContentLastSnapshot 和 Version 迁移到新笔记
	err = s.noteRepo.UpdateSnapshot(ctx, oldNote.ContentLastSnapshot, oldNote.ContentLastSnapshotHash, oldNote.Version, newNoteID, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 标记删除旧笔记，并标记是 rename 删除的笔记
	oldNote.Action = domain.NoteActionDelete
	oldNote.Rename = 1

	err = s.noteRepo.UpdateDelete(ctx, oldNote, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	go s.CountSizeSum(context.Background(), oldNote.VaultID, uid)
	return nil
}

// MigratePush 提交笔记迁移任务
func (s *noteService) MigratePush(oldNoteID, newNoteID int64, uid int64) {
	NoteMigrateChannel <- NoteMigrateMsg{
		OldNoteID: oldNoteID,
		NewNoteID: newNoteID,
		UID:       uid,
	}
}

// Sync syncs notes (alias for ListByLastTime, used for WebSocket sync)
func (s *noteService) Sync(ctx context.Context, uid int64, params *dto.NoteSyncRequest) ([]*dto.NoteDTO, error) {
	return s.ListByLastTime(ctx, uid, params)
}

// PatchFrontmatter patches note frontmatter with updates and removes specified keys
func (s *noteService) PatchFrontmatter(ctx context.Context, uid int64, params *dto.NotePatchFrontmatterRequest) (*dto.NoteDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	note, err := s.noteRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorNoteNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Parse existing frontmatter
	existingYaml, body, _ := util.ParseFrontmatter(note.Content)
	if existingYaml == nil {
		existingYaml = make(map[string]interface{})
	}

	// Merge updates
	newYaml := util.MergeFrontmatter(existingYaml, params.Updates, params.Remove)

	// Reconstruct content
	newContent := util.ReconstructContent(newYaml, body)

	// Save via ModifyOrCreate
	modifyParams := &dto.NoteModifyOrCreateRequest{
		Vault:       params.Vault,
		Path:        params.Path,
		PathHash:    params.PathHash,
		Content:     newContent,
		ContentHash: util.EncodeHash32(newContent),
		Mtime:       time.Now().UnixMilli(),
		Ctime:       note.Ctime,
	}

	_, result, err := s.ModifyOrCreate(ctx, uid, modifyParams, false)
	return result, err
}

// AppendContent appends content to the end of a note
func (s *noteService) AppendContent(ctx context.Context, uid int64, params *dto.NoteAppendRequest) (*dto.NoteDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	note, err := s.noteRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorNoteNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Append content
	newContent := note.Content + params.Content

	// Save via ModifyOrCreate
	modifyParams := &dto.NoteModifyOrCreateRequest{
		Vault:       params.Vault,
		Path:        params.Path,
		PathHash:    params.PathHash,
		Content:     newContent,
		ContentHash: util.EncodeHash32(newContent),
		Mtime:       time.Now().UnixMilli(),
		Ctime:       note.Ctime,
	}

	_, result, err := s.ModifyOrCreate(ctx, uid, modifyParams, false)
	return result, err
}

// PrependContent prepends content to a note (after frontmatter if present)
func (s *noteService) PrependContent(ctx context.Context, uid int64, params *dto.NotePrependRequest) (*dto.NoteDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	note, err := s.noteRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorNoteNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Parse frontmatter to preserve it
	yamlData, body, hasFrontmatter := util.ParseFrontmatter(note.Content)

	// Prepend content to body
	newBody := params.Content + body

	// Reconstruct content
	var newContent string
	if hasFrontmatter {
		newContent = util.ReconstructContent(yamlData, newBody)
	} else {
		newContent = newBody
	}

	// Save via ModifyOrCreate
	modifyParams := &dto.NoteModifyOrCreateRequest{
		Vault:       params.Vault,
		Path:        params.Path,
		PathHash:    params.PathHash,
		Content:     newContent,
		ContentHash: util.EncodeHash32(newContent),
		Mtime:       time.Now().UnixMilli(),
		Ctime:       note.Ctime,
	}

	_, result, err := s.ModifyOrCreate(ctx, uid, modifyParams, false)
	return result, err
}

// ReplaceContent performs find/replace in a note
func (s *noteService) ReplaceContent(ctx context.Context, uid int64, params *dto.NoteReplaceRequest) (*dto.NoteReplaceResponse, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	note, err := s.noteRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorNoteNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var matchCount int
	var newContent string

	if params.Regex {
		// Regex mode
		re, err := regexp.Compile(params.Find)
		if err != nil {
			return nil, code.ErrorInvalidRegex.WithDetails(err.Error())
		}

		matches := re.FindAllStringIndex(note.Content, -1)
		matchCount = len(matches)

		if params.All {
			newContent = re.ReplaceAllString(note.Content, params.Replace)
		} else if matchCount > 0 {
			newContent = re.ReplaceAllStringFunc(note.Content, func(s string) string {
				return params.Replace
			})
			// Only replace first match
			loc := re.FindStringIndex(note.Content)
			if loc != nil {
				newContent = note.Content[:loc[0]] + params.Replace + note.Content[loc[1]:]
			}
		} else {
			newContent = note.Content
		}
	} else {
		// Plain text mode
		matchCount = strings.Count(note.Content, params.Find)

		if params.All {
			newContent = strings.ReplaceAll(note.Content, params.Find, params.Replace)
		} else if matchCount > 0 {
			newContent = strings.Replace(note.Content, params.Find, params.Replace, 1)
		} else {
			newContent = note.Content
		}
	}

	// Check if no match found and fail flag is set
	if matchCount == 0 && params.FailIfNoMatch {
		return nil, code.ErrorNoMatchFound
	}

	// If no changes, return early
	if newContent == note.Content {
		return &dto.NoteReplaceResponse{
			MatchCount: matchCount,
			Note:       s.domainToDTO(note),
		}, nil
	}

	// Save via ModifyOrCreate
	modifyParams := &dto.NoteModifyOrCreateRequest{
		Vault:       params.Vault,
		Path:        params.Path,
		PathHash:    params.PathHash,
		Content:     newContent,
		ContentHash: util.EncodeHash32(newContent),
		Mtime:       time.Now().UnixMilli(),
		Ctime:       note.Ctime,
	}

	_, result, err := s.ModifyOrCreate(ctx, uid, modifyParams, false)
	if err != nil {
		return nil, err
	}

	return &dto.NoteReplaceResponse{
		MatchCount: matchCount,
		Note:       result,
	}, nil
}

// Move moves a note to a new path
func (s *noteService) Move(ctx context.Context, uid int64, params *dto.NoteMoveRequest) (*dto.NoteDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// Get source note
	sourceNote, err := s.noteRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorNoteNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	destPathHash := util.EncodeHash32(params.Destination)

	// Check if destination exists
	destNote, _ := s.noteRepo.GetByPathHash(ctx, destPathHash, vaultID, uid)
	if destNote != nil {
		if !params.Overwrite {
			return nil, code.ErrorNoteConflict
		}
		// Delete destination if overwrite is allowed
		deleteParams := &dto.NoteDeleteRequest{
			Vault:    params.Vault,
			Path:     params.Destination,
			PathHash: destPathHash,
		}
		_, err = s.Delete(ctx, uid, deleteParams)
		if err != nil {
			return nil, err
		}
	}

	// Create new note at destination with same content
	modifyParams := &dto.NoteModifyOrCreateRequest{
		Vault:       params.Vault,
		Path:        params.Destination,
		PathHash:    destPathHash,
		Content:     sourceNote.Content,
		ContentHash: sourceNote.ContentHash,
		Mtime:       time.Now().UnixMilli(),
		Ctime:       sourceNote.Ctime,
	}

	_, result, err := s.ModifyOrCreate(ctx, uid, modifyParams, false)
	if err != nil {
		return nil, err
	}

	// Delete source note
	deleteParams := &dto.NoteDeleteRequest{
		Vault:    params.Vault,
		Path:     params.Path,
		PathHash: params.PathHash,
	}
	_, err = s.Delete(ctx, uid, deleteParams)
	if err != nil {
		// Log but don't fail - the move already succeeded
		zap.L().Warn("Move: failed to delete source note",
			zap.Int64(logger.FieldUID, uid),
			zap.String("path", params.Path),
			zap.Error(err),
		)
	}

	// Trigger history migration
	if result != nil {
		s.MigratePush(sourceNote.ID, result.ID, uid)
	}

	return result, nil
}

// updateNoteLinks extracts wiki links from content and updates the link index
func (s *noteService) updateNoteLinks(ctx context.Context, noteID int64, content string, vaultID, uid int64) {
	if s.noteLinkRepo == nil {
		return
	}

	// Delete existing links for this note
	_ = s.noteLinkRepo.DeleteBySourceNoteID(ctx, noteID, uid)

	// Parse wiki links from content
	links := util.ParseWikiLinks(content)
	if len(links) == 0 {
		return
	}

	// Create new link records
	var noteLinks []*domain.NoteLink
	for _, link := range links {
		noteLinks = append(noteLinks, &domain.NoteLink{
			SourceNoteID:   noteID,
			TargetPath:     link.Path,
			TargetPathHash: util.EncodeHash32(link.Path),
			LinkText:       link.Alias,
			IsEmbed:        link.IsEmbed,
			VaultID:        vaultID,
		})
	}

	_ = s.noteLinkRepo.CreateBatch(ctx, noteLinks, uid)
}

// Ensure noteService implements NoteService interface
var _ NoteService = (*noteService)(nil)
