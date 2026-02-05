// Package service implements the business logic layer
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

// NoteService defines the note business service interface
// NoteService 定义笔记业务服务接口
type NoteService interface {
	// Get retrieves a single note
	// Get 获取单条笔记
	Get(ctx context.Context, uid int64, params *dto.NoteGetRequest) (*dto.NoteDTO, error)

	// UpdateCheck checks if note needs updating
	// UpdateCheck 检查笔记是否需要更新
	UpdateCheck(ctx context.Context, uid int64, params *dto.NoteUpdateCheckRequest) (string, *dto.NoteDTO, error)

	// ModifyOrCreate creates or modifies a note
	// ModifyOrCreate 创建或修改笔记
	ModifyOrCreate(ctx context.Context, uid int64, params *dto.NoteModifyOrCreateRequest, mtimeCheck bool) (bool, *dto.NoteDTO, error)

	// Delete deletes a note
	// Delete 删除笔记
	Delete(ctx context.Context, uid int64, params *dto.NoteDeleteRequest) (*dto.NoteDTO, error)

	// Restore restores a note (from recycle bin)
	// Restore 恢复笔记（从回收站恢复）
	Restore(ctx context.Context, uid int64, params *dto.NoteRestoreRequest) (*dto.NoteDTO, error)

	// Rename renames a note
	// Rename 重命名笔记
	Rename(ctx context.Context, uid int64, params *dto.NoteRenameRequest) (*dto.NoteDTO, *dto.NoteDTO, error)

	// List retrieves note list
	// List 获取笔记列表
	List(ctx context.Context, uid int64, params *dto.NoteListRequest, pager *app.Pager) ([]*dto.NoteNoContentDTO, int, error)

	// ListByLastTime retrieves notes updated after lastTime
	// ListByLastTime 获取在 lastTime 之后更新的笔记
	ListByLastTime(ctx context.Context, uid int64, params *dto.NoteSyncRequest) ([]*dto.NoteDTO, error)

	// Sync syncs notes (alias for ListByLastTime, used for WebSocket sync)
	// Sync 同步笔记（ListByLastTime 的别名，用于 WebSocket 同步）
	Sync(ctx context.Context, uid int64, params *dto.NoteSyncRequest) ([]*dto.NoteDTO, error)

	// CountSizeSum counts total number and size of notes in a vault
	// CountSizeSum 统计 vault 中笔记总数与总大小
	CountSizeSum(ctx context.Context, vaultID int64, uid int64) error

	// Cleanup cleans up expired soft-deleted notes
	// Cleanup 清理过期的软删除笔记
	Cleanup(ctx context.Context, uid int64) error

	// CleanupByTime cleans up expired soft-deleted notes for all users by cutoff time
	// CleanupByTime 按截止时间清理所有用户的过期软删除笔记
	CleanupByTime(ctx context.Context, cutoffTime int64) error

	// ListNeedSnapshot retrieves notes that need snapshot
	// ListNeedSnapshot 获取需要快照的笔记
	ListNeedSnapshot(ctx context.Context, uid int64) ([]*dto.NoteDTO, error)

	// Migrate migrates note history records
	// Migrate 迁移笔记历史记录
	Migrate(ctx context.Context, oldNoteID, newNoteID int64, uid int64) error

	// MigratePush submits note migration task
	// MigratePush 提交笔记迁移任务
	MigratePush(oldNoteID, newNoteID int64, uid int64)

	// WithClient sets client info
	// WithClient 设置客户端信息
	WithClient(name, version string) NoteService

	// PatchFrontmatter patches note frontmatter
	// PatchFrontmatter 修改笔记 Frontmatter
	PatchFrontmatter(ctx context.Context, uid int64, params *dto.NotePatchFrontmatterRequest) (*dto.NoteDTO, error)

	// AppendContent appends content to a note
	// AppendContent 在笔记末尾追加内容
	AppendContent(ctx context.Context, uid int64, params *dto.NoteAppendRequest) (*dto.NoteDTO, error)

	// PrependContent prepends content to a note
	// PrependContent 在笔记开头插入内容
	PrependContent(ctx context.Context, uid int64, params *dto.NotePrependRequest) (*dto.NoteDTO, error)

	// ReplaceContent performs find/replace in a note
	// ReplaceContent 在笔记中执行替换
	ReplaceContent(ctx context.Context, uid int64, params *dto.NoteReplaceRequest) (*dto.NoteReplaceResponse, error)

	// Move moves a note to a new path
	// Move 移动笔记
	Move(ctx context.Context, uid int64, params *dto.NoteMoveRequest) (*dto.NoteDTO, error)

	// UpdateNoteLinks extracts wiki links from content and updates the link index
	// UpdateNoteLinks 从内容中提取 Wiki 链接并更新链接索引
	UpdateNoteLinks(ctx context.Context, noteID int64, content string, vaultID, uid int64)
}

// noteService implementation of NoteService interface
// noteService 实现 NoteService 接口
type noteService struct {
	noteRepo      domain.NoteRepository     // Note repository // 笔记仓库
	noteLinkRepo  domain.NoteLinkRepository // Note link repository // 笔记链接仓库
	fileRepo      domain.FileRepository     // File repository // 文件仓库
	vaultService  VaultService              // Vault service // 仓库服务
	folderService FolderService             // Folder service // 文件夹服务
	sf            *singleflight.Group       // Singleflight group // 并发请求合并组
	clientName    string                    // Client name // 客户端名称
	clientVer     string                    // Client version // 客户端版本
	config        *ServiceConfig            // Service configuration // 服务配置
}

// NewNoteService creates NoteService instance
// NewNoteService 创建 NoteService 实例
func NewNoteService(noteRepo domain.NoteRepository, noteLinkRepo domain.NoteLinkRepository, fileRepo domain.FileRepository, vaultSvc VaultService, folderSvc FolderService, config *ServiceConfig) NoteService {
	return &noteService{
		noteRepo:      noteRepo,
		noteLinkRepo:  noteLinkRepo,
		fileRepo:      fileRepo,
		vaultService:  vaultSvc,
		folderService: folderSvc,
		sf:            &singleflight.Group{},
		config:        config,
	}
}

// WithClient sets client info, returns new NoteService instance
// WithClient 设置客户端信息，返回新 NoteService 实例
func (s *noteService) WithClient(name, version string) NoteService {
	return &noteService{
		noteRepo:      s.noteRepo,
		noteLinkRepo:  s.noteLinkRepo,
		fileRepo:      s.fileRepo,
		vaultService:  s.vaultService,
		folderService: s.folderService,
		sf:            s.sf,
		clientName:    name,
		clientVer:     version,
		config:        s.config,
	}
}

// domainToDTO converts domain model to DTO
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
		Size:             note.Size,
		Ctime:            note.Ctime,
		Mtime:            note.Mtime,
		UpdatedTimestamp: note.UpdatedTimestamp,
		UpdatedAt:        timex.Time(note.UpdatedAt),
		CreatedAt:        timex.Time(note.CreatedAt),
	}
}

// domainToNoContentDTO converts domain model to DTO without content
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
		Size:             note.Size,
		Ctime:            note.Ctime,
		Mtime:            note.Mtime,
		UpdatedTimestamp: note.UpdatedTimestamp,
		UpdatedAt:        timex.Time(note.UpdatedAt),
		CreatedAt:        timex.Time(note.CreatedAt),
	}
}

// Get retrieves a single note
// Get 获取单条笔记
func (s *noteService) Get(ctx context.Context, uid int64, params *dto.NoteGetRequest) (*dto.NoteDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID
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

// UpdateCheck checks if note needs updating
// UpdateCheck 检查笔记是否需要更新
func (s *noteService) UpdateCheck(ctx context.Context, uid int64, params *dto.NoteUpdateCheckRequest) (string, *dto.NoteDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return "", nil, err
	}

	note, _ := s.noteRepo.GetAllByPathHash(ctx, params.PathHash, vaultID, uid)

	if note != nil {
		noteDTO := s.domainToDTO(note)
		// Check if content is consistent
		// 检查内容是否一致
		if note.Action == "delete" {
			return "Create", nil, nil
		}
		if note.ContentHash == params.ContentHash {
			// Notify user to update mtime when user mtime is less than server mtime
			// 当用户 mtime 小于服务端 mtime 时，通知用户更新 mtime
			if params.Mtime < note.Mtime {
				return "UpdateMtime", noteDTO, nil
			} else if params.Mtime > note.Mtime {
				if err := s.noteRepo.UpdateMtime(ctx, params.Mtime, note.ID, uid); err != nil {
					// Non-critical update failed, log warning but do not block flow
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

// ModifyOrCreate creates or modifies a note
// ModifyOrCreate 创建或修改笔记
func (s *noteService) ModifyOrCreate(ctx context.Context, uid int64, params *dto.NoteModifyOrCreateRequest, mtimeCheck bool) (bool, *dto.NoteDTO, error) {
	var isNew bool

	// Use VaultService.MustGetID to retrieve VaultID
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
		// If content is consistent but modification time is different, only update modification time
		// 检查内容是否一致但修改时间不同，则只更新修改时间
		if mtimeCheck && note.Mtime < params.Mtime && note.ContentHash == params.ContentHash {
			err := s.noteRepo.UpdateActionMtime(ctx, domain.NoteActionModify, params.Mtime, note.ID, uid)
			if err != nil {
				return isNew, nil, code.ErrorDBQuery.WithDetails(err.Error())
			}
			note.Mtime = params.Mtime
			return isNew, s.domainToDTO(note), nil
		}

		// Set action
		// 设置 action
		var action domain.NoteAction
		if note.Action == domain.NoteActionDelete {
			action = domain.NoteActionCreate
		} else {
			action = domain.NoteActionModify
		}

		// Update note
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
		note.Version++ // Increment version on content change // 内容变更时递增版本号

		updated, err := s.noteRepo.Update(ctx, note, uid)
		if err != nil {
			return isNew, nil, code.ErrorDBQuery.WithDetails(err.Error())
		}

		go s.folderService.SyncResourceFID(context.Background(), uid, vaultID, []int64{updated.ID}, nil)
		go s.CountSizeSum(context.Background(), vaultID, uid)
		go s.UpdateNoteLinks(context.Background(), updated.ID, params.Content, vaultID, uid)
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

	go s.folderService.SyncResourceFID(context.Background(), uid, vaultID, []int64{created.ID}, nil)
	go s.CountSizeSum(context.Background(), vaultID, uid)
	go s.UpdateNoteLinks(context.Background(), created.ID, params.Content, vaultID, uid)
	NoteHistoryDelayPush(created.ID, uid)

	return isNew, s.domainToDTO(created), nil
}

// Delete deletes a note
// Delete 删除笔记
func (s *noteService) Delete(ctx context.Context, uid int64, params *dto.NoteDeleteRequest) (*dto.NoteDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID
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

	// Update to deleted status
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

// Restore restores a note (from recycle bin)
// Restore 恢复笔记（从回收站恢复）
func (s *noteService) Restore(ctx context.Context, uid int64, params *dto.NoteRestoreRequest) (*dto.NoteDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err // VaultService 已返回 code.Error
	}

	// Get note from recycle bin
	// 从回收站获取笔记
	note, err := s.noteRepo.GetByPathHashIncludeRecycle(ctx, params.PathHash, vaultID, uid, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorNoteNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Check if note is deleted
	// 检查笔记是否已删除
	if note.Action != domain.NoteActionDelete {
		return nil, code.ErrorNoteNotFound
	}

	// Update to modified status and update modification time
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

	go s.folderService.SyncResourceFID(context.Background(), uid, vaultID, []int64{updated.ID}, nil)
	go s.CountSizeSum(context.Background(), vaultID, uid)
	go s.UpdateNoteLinks(context.Background(), updated.ID, updated.Content, vaultID, uid)

	NoteHistoryDelayPush(updated.ID, uid)

	return s.domainToDTO(updated), nil
}

// Rename renames a note
// Rename 重命名笔记
func (s *noteService) Rename(ctx context.Context, uid int64, params *dto.NoteRenameRequest) (*dto.NoteDTO, *dto.NoteDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, nil, err
	}

	newPath := strings.Trim(params.Path, "/")
	newPathHash := params.PathHash
	if newPathHash == "" {
		newPathHash = util.EncodeHash32(newPath)
	}

	// 1. 判断目标路径是否存在有效笔记
	existNote, _ := s.noteRepo.GetAllByPathHash(ctx, newPathHash, vaultID, uid)
	if existNote != nil && existNote.Action != domain.NoteActionDelete {
		return nil, nil, code.ErrorNoteExist
	}

	oldPath := strings.Trim(params.OldPath, "/")
	oldPathHash := params.OldPathHash
	if oldPathHash == "" {
		oldPathHash = util.EncodeHash32(oldPath)
	}

	// 2. 获取旧笔记
	n, err := s.noteRepo.GetByPathHash(ctx, oldPathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, code.ErrorNoteNotFound
		}
		return nil, nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 3. 标记旧笔记删除
	n.Action = domain.NoteActionDelete
	n.UpdatedTimestamp = timex.Now().UnixMilli()
	oldNote, err := s.noteRepo.Update(ctx, n, uid)
	if err != nil {
		return nil, nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 4. 新建或复用笔记记录
	var newNoteCreated *domain.Note
	if existNote != nil {
		// 复用已删除的记录
		existNote.Action = domain.NoteActionCreate
		existNote.Path = newPath
		existNote.PathHash = newPathHash
		existNote.FID, _ = s.folderService.EnsurePathFID(ctx, uid, vaultID, newPath)
		existNote.Content = n.Content
		existNote.ContentHash = n.ContentHash
		existNote.Version = n.Version
		existNote.Mtime = timex.Now().UnixMilli()
		existNote.UpdatedTimestamp = timex.Now().UnixMilli()
		newNoteCreated, err = s.noteRepo.Update(ctx, existNote, uid)
	} else {
		// 创建新记录
		newNote := &domain.Note{
			VaultID:          vaultID,
			Action:           domain.NoteActionCreate,
			Path:             newPath,
			PathHash:         newPathHash,
			FID:              n.FID,
			Ctime:            n.Ctime,
			Mtime:            timex.Now().UnixMilli(),
			UpdatedTimestamp: timex.Now().UnixMilli(),
			Content:          n.Content,
			ContentHash:      n.ContentHash,
			Version:          n.Version,
		}
		// 确保 FID 正确
		newNote.FID, _ = s.folderService.EnsurePathFID(ctx, uid, vaultID, newPath)
		newNoteCreated, err = s.noteRepo.Create(ctx, newNote, uid)
	}

	if err != nil {
		return nil, nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	go s.folderService.SyncResourceFID(context.Background(), uid, vaultID, []int64{newNoteCreated.ID}, nil)
	go s.Migrate(context.Background(), n.ID, newNoteCreated.ID, uid)

	return s.domainToDTO(oldNote), s.domainToDTO(newNoteCreated), nil
}

// List retrieves note list
// List 获取笔记列表
func (s *noteService) List(ctx context.Context, uid int64, params *dto.NoteListRequest, pager *app.Pager) ([]*dto.NoteNoContentDTO, int, error) {
	// Use VaultService.MustGetID to retrieve VaultID
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

// ListByLastTime retrieves notes updated after lastTime
// ListByLastTime 获取在 lastTime 之后更新的笔记
func (s *noteService) ListByLastTime(ctx context.Context, uid int64, params *dto.NoteSyncRequest) ([]*dto.NoteDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID
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

// CountSizeSum counts total number and total size of notes in a vault
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

// Cleanup cleans up expired soft-deleted notes
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

// CleanupByTime cleans up expired soft-deleted notes for all users by cutoff time
// CleanupByTime 按截止时间清理所有用户的过期软删除笔记
func (s *noteService) CleanupByTime(ctx context.Context, cutoffTime int64) error {
	return s.noteRepo.DeletePhysicalByTimeAll(ctx, cutoffTime)
}

// ListNeedSnapshot retrieves notes that need snapshot
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

// Migrate migrates note history records
// Migrate 迁移笔记历史记录
func (s *noteService) Migrate(ctx context.Context, oldNoteID, newNoteID int64, uid int64) error {
	// Get old note information
	// 获取旧笔记信息
	oldNote, err := s.noteRepo.GetByID(ctx, oldNoteID, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Migrate ContentLastSnapshot and Version from old note to new note
	// 将旧笔记的 ContentLastSnapshot 和 Version 迁移到新笔记
	err = s.noteRepo.UpdateSnapshot(ctx, oldNote.ContentLastSnapshot, oldNote.ContentLastSnapshotHash, oldNote.Version, newNoteID, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Mark old note as deleted, and mark as rename deleted
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

// MigratePush submits note migration task
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
// AppendContent 在笔记末尾追加内容
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
	// 追加内容
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
// PrependContent 在笔记开头插入内容（如果存在 Frontmatter 则在之后插入）
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
	// 解析 Frontmatter 以保留它
	yamlData, body, hasFrontmatter := util.ParseFrontmatter(note.Content)

	// Prepend content to body
	// 在正文开头插入内容
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
// ReplaceContent 在笔记中执行查找/替换
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
		// 正则模式
		re, err := regexp.Compile(params.Find)
		if err != nil {
			return nil, code.ErrorInvalidRegex.WithDetails(err.Error())
		}

		matches := re.FindAllStringIndex(note.Content, -1)
		matchCount = len(matches)

		if params.All {
			newContent = re.ReplaceAllString(note.Content, params.Replace)
		} else if matchCount > 0 {
			// Only replace first match
			// 仅替换第一个匹配项
			loc := re.FindStringIndex(note.Content)
			if loc != nil {
				newContent = note.Content[:loc[0]] + params.Replace + note.Content[loc[1]:]
			}
		} else {
			newContent = note.Content
		}
	} else {
		// Plain text mode
		// 纯文本模式
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
	// 检查是否未找到匹配项且设置了失败标志
	if matchCount == 0 && params.FailIfNoMatch {
		return nil, code.ErrorNoMatchFound
	}

	// If no changes, return early
	// 如果没有变化，提前返回
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
// Move 移动笔记
func (s *noteService) Move(ctx context.Context, uid int64, params *dto.NoteMoveRequest) (*dto.NoteDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// Get source note
	// 获取源笔记
	sourceNote, err := s.noteRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorNoteNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	destPathHash := util.EncodeHash32(params.Destination)

	// Check if destination exists
	// 检查目标位置是否已存在
	destNote, _ := s.noteRepo.GetByPathHash(ctx, destPathHash, vaultID, uid)
	if destNote != nil {
		if !params.Overwrite {
			return nil, code.ErrorNoteConflict
		}
		// Delete destination if overwrite is allowed
		// 如果允许覆盖，则删除目标笔记
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
	// 在目标位置创建同内容的新笔记
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
	// 删除源笔记
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
	// 触发历史记录迁移
	if result != nil {
		s.MigratePush(sourceNote.ID, result.ID, uid)
	}

	return result, nil
}

// UpdateNoteLinks extracts wiki links from content and updates the link index
// UpdateNoteLinks 从内容中提取 Wiki 链接并更新链接索引
func (s *noteService) UpdateNoteLinks(ctx context.Context, noteID int64, content string, vaultID, uid int64) {
	if s.noteLinkRepo == nil {
		return
	}

	// Delete existing links for this note
	// 删除该笔记现有的链接
	_ = s.noteLinkRepo.DeleteBySourceNoteID(ctx, noteID, uid)

	// Parse wiki links from content
	// 从内容中解析 Wiki 链接
	links := util.ParseWikiLinks(content)
	if len(links) == 0 {
		return
	}

	// Create new link records
	// 创建新链接记录
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

// Verify noteService implements NoteService interface
// 确保 noteService 实现了 NoteService interface
var _ NoteService = (*noteService)(nil)
