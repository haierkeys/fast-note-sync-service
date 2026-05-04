// Package service implements the business logic layer
// Package service 实现业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
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
	WithClient(clientType, name, version string) NoteService

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

	// UpdateNoteLinks extracts wiki links from content and updates the link index
	// UpdateNoteLinks 从内容中提取 Wiki 链接并更新链接索引
	UpdateNoteLinks(ctx context.Context, noteID int64, content string, vaultID, uid int64)

	// RecycleClear cleans up the recycle bin
	// RecycleClear 清理回收站
	RecycleClear(ctx context.Context, uid int64, params *dto.NoteRecycleClearRequest) error

	// CleanDuplicateNotes cleans up duplicate note records
	// CleanDuplicateNotes 清理重复的笔记记录
	CleanDuplicateNotes(ctx context.Context, uid int64, vaultID int64) error

	// CleanDuplicateNotesAll cleans up duplicate note records for all users
	// CleanDuplicateNotesAll 清理所有用户的重复笔记记录
	CleanDuplicateNotesAll(ctx context.Context) error
}

// noteService implementation of NoteService interface
// noteService 实现 NoteService 接口
type noteService struct {
	userRepo       domain.UserRepository      // User repository // 用户仓库
	noteRepo       domain.NoteRepository      // Note repository // 笔记仓库
	noteLinkRepo   domain.NoteLinkRepository  // Note link repository // 笔记链接仓库
	fileRepo       domain.FileRepository      // File repository // 文件仓库
	shareRepo      domain.UserShareRepository // Share repository for auto-revoke on delete // 分享仓库（删除时自动撤销）
	vaultService   VaultService               // Vault service // 仓库服务
	folderService  FolderService              // Folder service // 文件夹服务
	syncLogService SyncLogService             // Sync log service // 同步日志服务
	sf             *singleflight.Group        // Singleflight group // 并发请求合并组
	clientType     string                     // Client type // 客户端类型
	clientName     string                     // Client name // 客户端名称
	clientVer      string                     // Client version // 客户端版本
	config         *ServiceConfig             // Service configuration // 服务配置
	backupService  BackupService              // Backup service // 备份服务
	gitSyncService GitSyncService             // Git sync service // Git 同步服务
	countTimers    *sync.Map                  // Timers for CountSizeSum debounce // CountSizeSum 防抖计时器
}

// NewNoteService creates NoteService instance
// NewNoteService 创建 NoteService 实例
func NewNoteService(userRepo domain.UserRepository, noteRepo domain.NoteRepository, noteLinkRepo domain.NoteLinkRepository, fileRepo domain.FileRepository, shareRepo domain.UserShareRepository, vaultSvc VaultService, folderSvc FolderService, backupSvc BackupService, gitSyncSvc GitSyncService, syncLogSvc SyncLogService, config *ServiceConfig) NoteService {
	return &noteService{
		userRepo:       userRepo,
		noteRepo:       noteRepo,
		noteLinkRepo:   noteLinkRepo,
		fileRepo:       fileRepo,
		shareRepo:      shareRepo,
		vaultService:   vaultSvc,
		folderService:  folderSvc,
		backupService:  backupSvc,
		gitSyncService: gitSyncSvc,
		syncLogService: syncLogSvc,
		sf:             &singleflight.Group{},
		config:         config,
		countTimers:    &sync.Map{},
	}
}

// WithClient sets client info, returns new NoteService instance
// WithClient 设置客户端信息，返回新 NoteService 实例
func (s *noteService) WithClient(clientType, name, version string) NoteService {
	return &noteService{
		noteRepo:       s.noteRepo,
		noteLinkRepo:   s.noteLinkRepo,
		fileRepo:       s.fileRepo,
		shareRepo:      s.shareRepo,
		vaultService:   s.vaultService,
		folderService:  s.folderService,
		syncLogService: s.syncLogService,
		sf:             s.sf,
		clientType:     clientType,
		clientName:     name,
		clientVer:      version,
		config:         s.config,
		backupService:  s.backupService,
		gitSyncService: s.gitSyncService,
		countTimers:    s.countTimers, // Share the same timer map // 共享同一个计时器 map
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
		ClientName:       note.ClientName,
		ClientType:       note.ClientType,
		ClientVersion:    note.ClientVersion,
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
		ClientName:       note.ClientName,
		ClientType:       note.ClientType,
		ClientVersion:    note.ClientVersion,
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
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return false, nil, err
	}

	key := fmt.Sprintf("modify_or_create_%d_%d_%s", uid, vaultID, params.PathHash)
	type result struct {
		isNew bool
		dto   *dto.NoteDTO
	}

	val, err, _ := s.sf.Do(key, func() (any, error) {
		var isNew bool
		note, _ := s.noteRepo.GetAllByPathHash(ctx, params.PathHash, vaultID, uid)

		if note != nil {
			isNew = false

			// If createOnly is set and note exists (not deleted), return error
			if note.Action != domain.NoteActionDelete && params.CreateOnly {
				return nil, code.ErrorNoteExist
			}

			// Check if content is consistent, excluding notes marked as deleted
			if mtimeCheck && note.Action != domain.NoteActionDelete && note.Mtime == params.Mtime && note.ContentHash == params.ContentHash {
				return &result{isNew: isNew, dto: nil}, nil
			}
			// If content is consistent but modification time is different, only update modification time
			// 检查内容是否一致但修改时间不同，则只更新修改时间
			if mtimeCheck && note.Mtime < params.Mtime && note.ContentHash == params.ContentHash {
				err := s.noteRepo.UpdateActionMtime(ctx, domain.NoteActionModify, params.Mtime, note.ID, uid)
				if err != nil {
					return nil, code.ErrorDBQuery.WithDetails(err.Error())
				}
				note.Mtime = params.Mtime
				// Log mtime-only update // 记录仅 mtime 变更日志
				if s.syncLogService != nil {
					s.syncLogService.Log(uid, vaultID, domain.SyncLogTypeNote, domain.SyncLogActionModify, "mtime", note.Path, note.PathHash, s.clientType, s.clientName, s.clientVer, note.Size)
				}
				return &result{isNew: isNew, dto: s.domainToDTO(note)}, nil
			}

			// Set action // 设置 action
			var action domain.NoteAction
			if note.Action == domain.NoteActionDelete {
				action = domain.NoteActionCreate
			} else {
				action = domain.NoteActionModify
			}

			// Update note // 更新笔记
			note.VaultID = vaultID
			note.Path = params.Path
			note.PathHash = params.PathHash
			note.Content = params.Content
			note.ContentHash = params.ContentHash
			note.ClientName = s.clientName
			note.ClientType = s.clientType
			note.ClientVersion = s.clientVer
			note.Size = int64(len(params.Content))
			note.Mtime = params.Mtime
			note.Ctime = params.Ctime
			note.Action = action
			note.Rename = 0
			note.Version++ // Increment version on content change // 内容变更时递增版本号

			updated, err := s.noteRepo.Update(ctx, note, uid)
			if err != nil {
				return nil, code.ErrorDBQuery.WithDetails(err.Error())
			}

			// Log content modify // 记录内容变更日志
			if s.syncLogService != nil {
				s.syncLogService.Log(uid, vaultID, domain.SyncLogTypeNote, domain.SyncLogActionModify, "content,mtime", updated.Path, updated.PathHash, s.clientType, s.clientName, s.clientVer, updated.Size)
			}

			go s.folderService.SyncResourceFID(context.Background(), uid, vaultID, []int64{updated.ID}, nil)
			go s.CountSizeSum(context.Background(), vaultID, uid)
			go s.UpdateNoteLinks(context.Background(), updated.ID, params.Content, vaultID, uid)
			NoteHistoryDelayPush(updated.ID, uid)

			if s.backupService != nil {
				go s.backupService.NotifyUpdated(uid)
			}
			if s.gitSyncService != nil {
				go s.gitSyncService.NotifyUpdated(uid, vaultID)
			}

			return &result{isNew: isNew, dto: s.domainToDTO(updated)}, nil
		}

		// Create new note // 创建新笔记
		isNew = true
		newNote := &domain.Note{
			VaultID:       vaultID,
			Path:          params.Path,
			PathHash:      params.PathHash,
			Content:       params.Content,
			ContentHash:   params.ContentHash,
			ClientName:    s.clientName,
			ClientType:    s.clientType,
			ClientVersion: s.clientVer,
			Size:          int64(len(params.Content)),
			Mtime:         params.Mtime,
			Ctime:         params.Ctime,
			Action:        domain.NoteActionCreate,
		}

		created, err := s.noteRepo.Create(ctx, newNote, uid)
		if err != nil {
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}

		// Log create // 记录新建日志
		if s.syncLogService != nil {
			s.syncLogService.Log(uid, vaultID, domain.SyncLogTypeNote, domain.SyncLogActionCreate, "", created.Path, created.PathHash, s.clientType, s.clientName, s.clientVer, created.Size)
		}

		go s.folderService.SyncResourceFID(context.Background(), uid, vaultID, []int64{created.ID}, nil)
		go s.CountSizeSum(context.Background(), vaultID, uid)
		go s.UpdateNoteLinks(context.Background(), created.ID, params.Content, vaultID, uid)
		NoteHistoryDelayPush(created.ID, uid)
		if s.backupService != nil {
			go s.backupService.NotifyUpdated(uid)
		}

		if s.gitSyncService != nil {
			go s.gitSyncService.NotifyUpdated(uid, vaultID)
		}

		return &result{isNew: isNew, dto: s.domainToDTO(created)}, nil
	})

	if err != nil {
		return false, nil, err
	}

	res := val.(*result)
	return res.isNew, res.dto, nil
}

// Delete deletes a note
// Delete 删除笔记
func (s *noteService) Delete(ctx context.Context, uid int64, params *dto.NoteDeleteRequest) (*dto.NoteDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID // 使用 VaultService.MustGetID 获取 VaultID
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

	// Update to deleted status // 更新为删除状态
	note.Action = domain.NoteActionDelete
	note.ClientName = s.clientName
	note.ClientType = s.clientType
	note.ClientVersion = s.clientVer
	note.Rename = 0

	err = s.noteRepo.UpdateDelete(ctx, note, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// If note has active share, automatically revoke (to prevent count residue) // 若笔记有 active 分享，自动撤销（防止计数残留）
	if err := s.shareRepo.UpdateStatusByRes(ctx, uid, "note", note.ID, domain.UserShareStatusRevoked); err != nil {
		zap.L().Warn("Failed to revoke share on note deletion",
			zap.Int64("uid", uid),
			zap.Int64("noteId", note.ID),
			zap.String("pathHash", params.PathHash),
			zap.Error(err),
		)
	}

	// Log soft delete // 记录软删除日志
	if s.syncLogService != nil {
		s.syncLogService.Log(uid, vaultID, domain.SyncLogTypeNote, domain.SyncLogActionSoftDelete, "", note.Path, note.PathHash, s.clientType, s.clientName, s.clientVer, note.Size)
	}

	// Re-fetch the updated note // 重新获取更新后的笔记
	updated, err := s.noteRepo.GetByID(ctx, note.ID, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	NoteHistoryDelayPush(updated.ID, uid)
	if s.backupService != nil {
		go s.backupService.NotifyUpdated(uid)
	}
	if s.gitSyncService != nil {
		go s.gitSyncService.NotifyUpdated(uid, vaultID)
	}

	return s.domainToDTO(updated), nil
}

// Restore restores a note (from recycle bin)
// Restore 恢复笔记（从回收站恢复）
func (s *noteService) Restore(ctx context.Context, uid int64, params *dto.NoteRestoreRequest) (*dto.NoteDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID // 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err // VaultService 已返回 code.Error
	}

	// Get note from recycle bin // 从回收站获取笔记
	note, err := s.noteRepo.GetByPathHashIncludeRecycle(ctx, params.PathHash, vaultID, uid, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorNoteNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Check if note is deleted // 检查笔记是否已删除
	if note.Action != domain.NoteActionDelete {
		return nil, code.ErrorNoteNotFound
	}

	// Update to modified status and update modification time // 更新为修改状态 并更新修改时间
	note.Action = domain.NoteActionModify
	note.ClientName = s.clientName
	note.ClientType = s.clientType
	note.ClientVersion = s.clientVer
	note.Mtime = time.Now().UnixMilli()
	note.Rename = 0

	err = s.noteRepo.UpdateDelete(ctx, note, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Log restore // 记录恢复日志
	if s.syncLogService != nil {
		s.syncLogService.Log(uid, vaultID, domain.SyncLogTypeNote, domain.SyncLogActionRestore, "", note.Path, note.PathHash, s.clientType, s.clientName, s.clientVer, note.Size)
	}

	// Re-fetch the updated note
	// 重新获取更新后的笔记
	updated, err := s.noteRepo.GetByID(ctx, note.ID, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	go s.folderService.SyncResourceFID(context.Background(), uid, vaultID, []int64{updated.ID}, nil)
	go s.CountSizeSum(context.Background(), vaultID, uid)
	go s.UpdateNoteLinks(context.Background(), updated.ID, updated.Content, vaultID, uid)

	NoteHistoryDelayPush(updated.ID, uid)
	if s.backupService != nil {
		go s.backupService.NotifyUpdated(uid)
	}
	if s.gitSyncService != nil {
		go s.gitSyncService.NotifyUpdated(uid, vaultID)
	}

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

	key := fmt.Sprintf("rename_%d_%d_%s_%s", uid, vaultID, params.OldPathHash, newPathHash)
	type result struct {
		oldNote *dto.NoteDTO
		newNote *dto.NoteDTO
	}

	val, err, _ := s.sf.Do(key, func() (any, error) {
		// 1. Check if target path has valid note
		// 1. 判断目标路径是否存在有效笔记
		existNote, _ := s.noteRepo.GetAllByPathHash(ctx, newPathHash, vaultID, uid)
		if existNote != nil && existNote.Action != domain.NoteActionDelete {
			return nil, code.ErrorNoteExist
		}

		oldPath := strings.Trim(params.OldPath, "/")
		oldPathHash := params.OldPathHash
		if oldPathHash == "" {
			oldPathHash = util.EncodeHash32(oldPath)
		}

		// 2. Get old note
		// 2. 获取旧笔记
		n, err := s.noteRepo.GetByPathHash(ctx, oldPathHash, vaultID, uid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, code.ErrorNoteNotFound
			}
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}

		// 3. Mark old note as deleted with rename flag
		// 3. 标记旧笔记删除并带上重命名标志
		n.Action = domain.NoteActionDelete
		n.Rename = 1
		n.ClientName = s.clientName
		n.ClientType = s.clientType
		n.ClientVersion = s.clientVer
		n.UpdatedTimestamp = timex.Now().UnixMilli()
		oldNote, err := s.noteRepo.Update(ctx, n, uid)
		if err != nil {
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}

		// 4. New or reuse note record
		// 4. 新建或复用笔记记录
		var newNoteCreated *domain.Note
		if existNote != nil {
			// Reuse deleted record // 复用已删除的记录
			existNote.Action = domain.NoteActionCreate
			existNote.Path = newPath
			existNote.PathHash = newPathHash
			newPathDir := ""
			if idx := strings.LastIndex(newPath, "/"); idx >= 0 {
				newPathDir = newPath[:idx]
			}
			existNote.FID, _ = s.folderService.EnsurePathFID(ctx, uid, vaultID, newPathDir)
			existNote.Content = n.Content
			existNote.ContentHash = n.ContentHash
			existNote.Version = n.Version
			existNote.Mtime = n.Mtime // Preserve original mtime // 保留原始修改时间
			existNote.ClientName = s.clientName
			existNote.ClientType = s.clientType
			existNote.ClientVersion = s.clientVer
			existNote.UpdatedTimestamp = timex.Now().UnixMilli()
			newNoteCreated, err = s.noteRepo.Update(ctx, existNote, uid)
		} else {
			// Create new record // 创建新记录
			newNote := &domain.Note{
				VaultID:          vaultID,
				Action:           domain.NoteActionCreate,
				Path:             newPath,
				PathHash:         newPathHash,
				FID:              n.FID,
				Ctime:            n.Ctime,
				Mtime:            n.Mtime, // Preserve original mtime // 保留原始修改时间
				ClientName:       s.clientName,
				ClientType:       s.clientType,
				ClientVersion:    s.clientVer,
				UpdatedTimestamp: timex.Now().UnixMilli(),
				Content:          n.Content,
				ContentHash:      n.ContentHash,
				Version:          n.Version,
			}
			newPathDir := ""
			if idx := strings.LastIndex(newPath, "/"); idx >= 0 {
				newPathDir = newPath[:idx]
			}
			// Ensure FID is correct // 确保 FID 正确
			newNote.FID, _ = s.folderService.EnsurePathFID(ctx, uid, vaultID, newPathDir)
			newNoteCreated, err = s.noteRepo.Create(ctx, newNote, uid)
		}

		if err != nil {
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}

		// Log rename // 记录重命名日志
		if s.syncLogService != nil {
			s.syncLogService.Log(uid, vaultID, domain.SyncLogTypeNote, domain.SyncLogActionRename, "path", newNoteCreated.Path, newNoteCreated.PathHash, s.clientType, s.clientName, s.clientVer, newNoteCreated.Size)
		}

		go s.folderService.SyncResourceFID(context.Background(), uid, vaultID, []int64{newNoteCreated.ID}, nil)
		go s.Migrate(context.Background(), n.ID, newNoteCreated.ID, uid)
		if s.backupService != nil {
			go s.backupService.NotifyUpdated(uid)
		}
		if s.gitSyncService != nil {
			go s.gitSyncService.NotifyUpdated(uid, vaultID)
		}

		return &result{oldNote: s.domainToDTO(oldNote), newNote: s.domainToDTO(newNoteCreated)}, nil
	})

	if err != nil {
		return nil, nil, err
	}

	res := val.(*result)
	return res.oldNote, res.newNote, nil
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

	// Parse paths parameter (comma-separated -> []string)
	// 解析 paths 参数（逗号分隔 → []string）
	var paths []string
	if params.Paths != "" {
		for _, p := range strings.Split(params.Paths, ",") {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				paths = append(paths, trimmed)
			}
		}
	}

	notes, err := s.noteRepo.List(ctx, vaultID, pager.Page, pager.PageSize, uid, params.Keyword, params.IsRecycle, params.SearchMode, params.SearchContent, params.SortBy, params.SortOrder, paths)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}

	count, err := s.noteRepo.ListCount(ctx, vaultID, uid, params.Keyword, params.IsRecycle, params.SearchMode, params.SearchContent, paths)
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
	key := fmt.Sprintf("%d_%d", uid, vaultID)

	// Debounce: 10 seconds delay. If a new request comes within 10s, reset the timer.
	// 防抖：10秒延迟。如果10秒内 house 有新请求，重置计时器。
	if timerOld, ok := s.countTimers.Load(key); ok {
		if t, ok := timerOld.(*time.Timer); ok {
			t.Stop()
		}
	}

	timer := time.AfterFunc(10*time.Second, func() {
		defer s.countTimers.Delete(key)

		// Use singleflight to ensure only one actual DB query runs for same key even if debounce period ends simultaneously
		// 使用 singleflight 确保即使防抖期同时结束，同一 key 也只有一个真实的 DB 查询
		s.sf.Do(key, func() (any, error) {
			result, err := s.noteRepo.CountSizeSum(context.Background(), vaultID, uid)
			if err != nil {
				return nil, code.ErrorDBQuery.WithDetails(err.Error())
			}
			return nil, s.vaultService.UpdateNoteStats(context.Background(), result.Size, result.Count, vaultID, uid)
		})
	})

	s.countTimers.Store(key, timer)
	return nil
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
	// Get old note information
	// 获取旧笔记信息
	oldNote, err := s.noteRepo.GetByID(ctx, oldNoteID, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Migrate ContentLastSnapshot and Version from old note to new note
	// Migrate ContentLastSnapshot and Version from old note to new note
	// 将旧笔记的 ContentLastSnapshot 和 Version 迁移到新笔记
	err = s.noteRepo.UpdateSnapshot(ctx, oldNote.ContentLastSnapshot, oldNote.ContentLastSnapshotHash, oldNote.Version, newNoteID, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Mark old note as deleted, and mark as rename deleted
	// Mark old note as deleted, and mark as rename deleted
	// 标记删除旧笔记，并标记是 rename 删除的笔记
	oldNote.Action = domain.NoteActionDelete
	oldNote.Rename = 1

	err = s.noteRepo.UpdateDelete(ctx, oldNote, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Migrate share records: update res_id and resources from old note ID to new note ID
	// Migrate share records: update res_id and resources from old note ID to new note ID
	// 迁移分享记录：将旧笔记 ID 的分享指向新笔记 ID
	if s.shareRepo != nil {
		if shareErr := s.shareRepo.MigrateResID(ctx, uid, oldNoteID, newNoteID); shareErr != nil {
			// Log but don't fail the rename operation
			zap.L().Warn("Migrate: failed to migrate share records",
				zap.Int64(logger.FieldUID, uid),
				zap.Int64("oldNoteID", oldNoteID),
				zap.Int64("newNoteID", newNoteID),
				zap.Error(shareErr))
		}
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
	// Parse frontmatter to preserve it
	// 解析 Frontmatter 以保留它
	yamlData, body, hasFrontmatter := util.ParseFrontmatter(note.Content)

	// Prepend content to body
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
	// Check if no match found and fail flag is set
	// 检查是否未找到匹配项且设置了失败标志
	if matchCount == 0 && params.FailIfNoMatch {
		return nil, code.ErrorNoMatchFound
	}

	// If no changes, return early
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

// UpdateNoteLinks extracts wiki links from content and updates the link index
// UpdateNoteLinks 从内容中提取 Wiki 链接并更新链接索引
func (s *noteService) UpdateNoteLinks(ctx context.Context, noteID int64, content string, vaultID, uid int64) {
	if s.noteLinkRepo == nil {
		return
	}

	// Delete existing links for this note
	// Delete existing links for this note
	// 删除该笔记现有的链接
	_ = s.noteLinkRepo.DeleteBySourceNoteID(ctx, noteID, uid)

	// Parse wiki links from content
	// Parse wiki links from content
	// 从内容中解析 Wiki 链接
	links := util.ParseWikiLinks(content)
	if len(links) == 0 {
		return
	}

	// Create new link records
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

// RecycleClear 清理回收站
func (s *noteService) RecycleClear(ctx context.Context, uid int64, params *dto.NoteRecycleClearRequest) error {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return err
	}

	if params.Path != "" && params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// Capture items to be deleted for detailed logging
	// 捕获待删除的项目以便进行详细日志记录
	var notesToDelete []*domain.Note
	if params.PathHash != "" {
		note, _ := s.noteRepo.GetByPathHashIncludeRecycle(ctx, params.PathHash, vaultID, uid, true)
		if note != nil {
			notesToDelete = append(notesToDelete, note)
			if params.Path == "" {
				params.Path = note.Path
			}
		}
	} else {
		// Clear all: retrieve all notes in recycle bin (using a large page size)
		// 清理全部：获取回收站中的所有笔记（使用较大的分页限制）
		notesToDelete, _ = s.noteRepo.List(ctx, vaultID, 1, 10000, uid, "", true, "", false, "", "", nil)
	}

	err = s.noteRepo.RecycleClear(ctx, params.Path, params.PathHash, vaultID, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Log permanent delete for each item // 为每一项记录彻底删除日志
	if s.syncLogService != nil {
		for _, n := range notesToDelete {
			s.syncLogService.Log(uid, vaultID, domain.SyncLogTypeNote, domain.SyncLogActionDelete, "", n.Path, n.PathHash, s.clientType, s.clientName, s.clientVer, n.Size)
		}
	}

	go s.CountSizeSum(context.Background(), vaultID, uid)
	return nil
}

// CleanDuplicateNotes 清理重复的笔记记录
func (s *noteService) CleanDuplicateNotes(ctx context.Context, uid int64, vaultID int64) error {
	// 获取所有笔记（包含已删除，以便全局去重）
	notes, err := s.noteRepo.ListByUpdatedTimestamp(ctx, 0, vaultID, uid)
	if err != nil {
		return err
	}

	// Group by Path
	// 按 Path 分组
	grouped := make(map[string][]*domain.Note)
	for _, n := range notes {
		grouped[n.Path] = append(grouped[n.Path], n)
	}

	for _, list := range grouped {
		if len(list) <= 1 {
			continue
		}

		// Retention rules:
		// 保留规则：
		// 1. Prioritize records with Action != delete
		// 1. 优先保留 Action != delete 的记录
		// 2. If multiple active records exist, keep the one with the largest UpdatedTimestamp (latest)
		// 2. 如果有多个活跃记录，保留 UpdatedTimestamp 最大（最新）的一条
		// 3. If timestamps are identical, keep the record with the largest ID
		// 3. 如果时间戳一致，保留 ID 最大的记录

		var bestNote *domain.Note
		for _, n := range list {
			if bestNote == nil {
				bestNote = n
				continue
			}

			// Comparison logic
			// 比较逻辑
			isBetter := false
			if n.Action != domain.NoteActionDelete && bestNote.Action == domain.NoteActionDelete {
				isBetter = true
			} else if n.Action == bestNote.Action {
				if n.UpdatedTimestamp > bestNote.UpdatedTimestamp {
					isBetter = true
				} else if n.UpdatedTimestamp == bestNote.UpdatedTimestamp && n.ID > bestNote.ID {
					isBetter = true
				}
			}

			if isBetter {
				bestNote = n
			}
		}

		// Delete all records except the bestNote
		// 删除非 bestNote 的所有记录
		for _, n := range list {
			if n.ID != bestNote.ID {
				// Clear singleflight cache to prevent residual data
				// 清除 singleflight 缓存，防止残留
				s.sf.Forget(fmt.Sprintf("modify_or_create_%d_%d_%s", uid, vaultID, n.PathHash))
				s.sf.Forget(fmt.Sprintf("rename_%d_%d_%s", uid, vaultID, n.PathHash))
				
				_ = s.noteRepo.Delete(ctx, n.ID, uid)
			}
		}
	}

	return nil
}

// CleanDuplicateNotesAll 清理所有用户的重复笔记记录
func (s *noteService) CleanDuplicateNotesAll(ctx context.Context) error {
	uids, err := s.userRepo.GetAllUIDs(ctx)
	if err != nil {
		return err
	}

	for _, uid := range uids {
		vaults, err := s.vaultService.List(ctx, uid)
		if err != nil {
			continue
		}
		for _, vault := range vaults {
			_ = s.CleanDuplicateNotes(ctx, uid, vault.ID)
		}
	}
	return nil
}

// Ensure noteService implements NoteService interface
// 确保 noteService 实现了 NoteService interface
var _ NoteService = (*noteService)(nil)
