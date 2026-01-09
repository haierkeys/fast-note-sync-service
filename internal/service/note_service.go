// Package service 实现业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
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
	Get(ctx context.Context, uid int64, params *dto.NoteGetRequest) (*NoteDTO, error)

	// UpdateCheck 检查笔记是否需要更新
	UpdateCheck(ctx context.Context, uid int64, params *dto.NoteUpdateCheckRequest) (string, *NoteDTO, error)

	// ModifyOrCreate 创建或修改笔记
	ModifyOrCreate(ctx context.Context, uid int64, params *dto.NoteModifyOrCreateRequest, mtimeCheck bool) (bool, *NoteDTO, error)

	// Delete 删除笔记
	Delete(ctx context.Context, uid int64, params *dto.NoteDeleteRequest) (*NoteDTO, error)

	// Rename 重命名笔记
	Rename(ctx context.Context, uid int64, params *dto.NoteRenameRequest) error

	// List 获取笔记列表
	List(ctx context.Context, uid int64, params *dto.NoteListRequest, pager *app.Pager) ([]*NoteNoContentDTO, int, error)

	// ListByLastTime 获取在 lastTime 之后更新的笔记
	ListByLastTime(ctx context.Context, uid int64, params *dto.NoteSyncRequest) ([]*NoteDTO, error)

	// Sync 同步笔记（ListByLastTime 的别名，用于 WebSocket 同步）
	Sync(ctx context.Context, uid int64, params *dto.NoteSyncRequest) ([]*NoteDTO, error)

	// GetFileContent 获取笔记或附件文件的原始内容
	GetFileContent(ctx context.Context, uid int64, params *dto.NoteGetRequest) ([]byte, string, int64, string, error)

	// CountSizeSum 统计 vault 中笔记总数与总大小
	CountSizeSum(ctx context.Context, vaultID int64, uid int64) error

	// Cleanup 清理过期的软删除笔记
	Cleanup(ctx context.Context, uid int64) error

	// CleanupAll 清理所有用户的过期软删除笔记
	CleanupAll(ctx context.Context) error

	// ListNeedSnapshot 获取需要快照的笔记
	ListNeedSnapshot(ctx context.Context, uid int64) ([]*NoteDTO, error)

	// Migrate 迁移笔记历史记录
	Migrate(ctx context.Context, oldNoteID, newNoteID int64, uid int64) error

	// MigratePush 提交笔记迁移任务
	MigratePush(oldNoteID, newNoteID int64, uid int64)

	// WithClient 设置客户端信息
	WithClient(name, version string) NoteService
}

// NoteDTO 笔记数据传输对象
type NoteDTO struct {
	ID               int64      `json:"id" form:"id"`
	Action           string     `json:"-" form:"action"`
	Path             string     `json:"path" form:"path"`
	PathHash         string     `json:"pathHash" form:"pathHash"`
	Content          string     `json:"content" form:"content"`
	ContentHash      string     `json:"contentHash" form:"contentHash"`
	Version          int64      `json:"version" form:"version"`
	Ctime            int64      `json:"ctime" form:"ctime"`
	Mtime            int64      `json:"mtime" form:"mtime"`
	UpdatedTimestamp int64      `json:"lastTime" form:"updatedTimestamp"`
	UpdatedAt        timex.Time `json:"-"`
	CreatedAt        timex.Time `json:"-"`
}

// NoteNoContentDTO 不包含内容的笔记 DTO
type NoteNoContentDTO struct {
	ID               int64      `json:"id" form:"id"`
	Action           string     `json:"action" form:"action"`
	Path             string     `json:"path" form:"path"`
	PathHash         string     `json:"pathHash" form:"pathHash"`
	Version          int64      `json:"version" form:"version"`
	Ctime            int64      `json:"ctime" form:"ctime"`
	Mtime            int64      `json:"mtime" form:"mtime"`
	UpdatedTimestamp int64      `json:"updatedTimestamp" form:"updatedTimestamp"`
	UpdatedAt        timex.Time `json:"updatedAt"`
	CreatedAt        timex.Time `json:"createdAt"`
}

// noteService 实现 NoteService 接口
type noteService struct {
	noteRepo     domain.NoteRepository
	fileRepo     domain.FileRepository
	vaultService VaultService
	sf           *singleflight.Group
	clientName   string
	clientVer    string
	config       *ServiceConfig

	// 清理相关
	lastCleanupTime time.Time
	cleanupMutex    sync.Mutex
}

// NewNoteService 创建 NoteService 实例
func NewNoteService(noteRepo domain.NoteRepository, fileRepo domain.FileRepository, vaultSvc VaultService, config *ServiceConfig) NoteService {
	return &noteService{
		noteRepo:     noteRepo,
		fileRepo:     fileRepo,
		vaultService: vaultSvc,
		sf:           &singleflight.Group{},
		config:       config,
	}
}

// WithClient 设置客户端信息，返回新的 NoteService 实例
func (s *noteService) WithClient(name, version string) NoteService {
	return &noteService{
		noteRepo:     s.noteRepo,
		fileRepo:     s.fileRepo,
		vaultService: s.vaultService,
		sf:           s.sf,
		clientName:   name,
		clientVer:    version,
		config:       s.config,
	}
}

// domainToDTO 将领域模型转换为 DTO
func (s *noteService) domainToDTO(note *domain.Note) *NoteDTO {
	if note == nil {
		return nil
	}
	return &NoteDTO{
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
func (s *noteService) domainToNoContentDTO(note *domain.Note) *NoteNoContentDTO {
	if note == nil {
		return nil
	}
	return &NoteNoContentDTO{
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
func (s *noteService) Get(ctx context.Context, uid int64, params *dto.NoteGetRequest) (*NoteDTO, error) {
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
func (s *noteService) UpdateCheck(ctx context.Context, uid int64, params *dto.NoteUpdateCheckRequest) (string, *NoteDTO, error) {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return "", nil, err
	}

	note, _ := s.noteRepo.GetAllByPathHash(ctx, params.PathHash, vaultID, uid)
	if note != nil {
		noteDTO := s.domainToDTO(note)
		// 检查内容是否一致
		if note.ContentHash == params.ContentHash {
			// 当用户 mtime 小于服务端 mtime 时，通知用户更新 mtime
			if params.Mtime < note.Mtime {
				return "UpdateMtime", noteDTO, nil
			} else if params.Mtime > note.Mtime {
				if err := s.noteRepo.UpdateMtime(ctx, params.Mtime, note.ID, uid); err != nil {
					// 非关键更新失败，记录警告日志但不阻断流程
					global.Logger.Warn("UpdateMtime failed for note",
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
func (s *noteService) ModifyOrCreate(ctx context.Context, uid int64, params *dto.NoteModifyOrCreateRequest, mtimeCheck bool) (bool, *NoteDTO, error) {
	var isNew bool

	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return isNew, nil, err
	}

	note, _ := s.noteRepo.GetAllByPathHash(ctx, params.PathHash, vaultID, uid)
	if note != nil {
		isNew = false
		// 检查内容是否一致
		if mtimeCheck && note.Mtime == params.Mtime && note.ContentHash == params.ContentHash {
			return isNew, nil, nil
		}
		// 检查内容是否一致但修改时间不同，则只更新修改时间
		if mtimeCheck && note.Mtime < params.Mtime && note.ContentHash == params.ContentHash {
			err := s.noteRepo.UpdateMtime(ctx, params.Mtime, note.ID, uid)
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

		updated, err := s.noteRepo.Update(ctx, note, uid)
		if err != nil {
			return isNew, nil, code.ErrorDBQuery.WithDetails(err.Error())
		}

		go s.CountSizeSum(ctx, vaultID, uid)
		NoteHistoryDelayPush(updated.ID, uid)

		return isNew, s.domainToDTO(updated), nil
	}

	// 创建新笔记
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

	go s.CountSizeSum(ctx, vaultID, uid)
	NoteHistoryDelayPush(created.ID, uid)

	return isNew, s.domainToDTO(created), nil
}

// Delete 删除笔记
func (s *noteService) Delete(ctx context.Context, uid int64, params *dto.NoteDeleteRequest) (*NoteDTO, error) {
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

	go s.CountSizeSum(ctx, vaultID, uid)
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
	oldNote, err := s.noteRepo.GetByPathHashIncludeRecycle(ctx, params.OldPathHash, vaultID, uid, true)
	if err != nil {
		return fmt.Errorf("old note not found: %w", err)
	}

	// 获取新笔记
	newNote, err := s.noteRepo.GetByPathHashIncludeRecycle(ctx, params.PathHash, vaultID, uid, false)
	if err != nil {
		return fmt.Errorf("new note not found: %w", err)
	}

	// 触发历史记录迁移
	s.MigratePush(oldNote.ID, newNote.ID, uid)

	return nil
}

// List 获取笔记列表
func (s *noteService) List(ctx context.Context, uid int64, params *dto.NoteListRequest, pager *app.Pager) ([]*NoteNoContentDTO, int, error) {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, 0, err
	}

	notes, err := s.noteRepo.List(ctx, vaultID, pager.Page, pager.PageSize, uid, params.Keyword, params.IsRecycle)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}

	count, err := s.noteRepo.ListCount(ctx, vaultID, uid, params.Keyword, params.IsRecycle)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var result []*NoteNoContentDTO
	for _, n := range notes {
		result = append(result, s.domainToNoContentDTO(n))
	}

	return result, int(count), nil
}

// ListByLastTime 获取在 lastTime 之后更新的笔记
func (s *noteService) ListByLastTime(ctx context.Context, uid int64, params *dto.NoteSyncRequest) ([]*NoteDTO, error) {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err // VaultService 已返回 code.Error
	}

	notes, err := s.noteRepo.ListByUpdatedTimestamp(ctx, params.LastTime, vaultID, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var results []*NoteDTO
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
		return err
	}

	if retentionDuration <= 0 {
		return nil
	}

	cutoffTime := time.Now().Add(-retentionDuration).UnixMilli()
	return s.noteRepo.DeletePhysicalByTime(ctx, cutoffTime, uid)
}

// CleanupAll 清理所有用户的过期软删除笔记
func (s *noteService) CleanupAll(ctx context.Context) error {
	if s.config == nil {
		return nil
	}
	retentionTimeStr := s.config.App.SoftDeleteRetentionTime
	if retentionTimeStr == "" || retentionTimeStr == "0" {
		return nil
	}

	retentionDuration, err := util.ParseDuration(retentionTimeStr)
	if err != nil {
		return err
	}

	if retentionDuration <= 0 {
		return nil
	}

	s.cleanupMutex.Lock()
	defer s.cleanupMutex.Unlock()

	// 动态计算检查间隔
	var checkInterval time.Duration
	if retentionDuration < time.Hour {
		checkInterval = time.Minute
	} else {
		checkInterval = retentionDuration / 10
		if checkInterval > time.Hour {
			checkInterval = time.Hour
		}
		if checkInterval < time.Minute {
			checkInterval = time.Minute
		}
	}

	// 如果距离上次清理时间不足检查间隔，则跳过
	if time.Since(s.lastCleanupTime) < checkInterval {
		return nil
	}

	// 注意：这里需要获取所有用户 UID，但 NoteService 不应该直接访问 UserRepository
	// 这个方法应该由上层调用者提供 UID 列表，或者通过其他方式获取
	// 暂时保留此方法签名，实际实现需要调整架构
	s.lastCleanupTime = time.Now()
	return nil
}

// ListNeedSnapshot 获取需要快照的笔记
func (s *noteService) ListNeedSnapshot(ctx context.Context, uid int64) ([]*NoteDTO, error) {
	list, err := s.noteRepo.ListContentUnchanged(ctx, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var result []*NoteDTO
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

	go s.CountSizeSum(ctx, oldNote.VaultID, uid)
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

// Sync 同步笔记（ListByLastTime 的别名，用于 WebSocket 同步）
func (s *noteService) Sync(ctx context.Context, uid int64, params *dto.NoteSyncRequest) ([]*NoteDTO, error) {
	return s.ListByLastTime(ctx, uid, params)
}

// GetFileContent 获取笔记或附件文件的原始内容
// 返回值说明:
//   - []byte: 文件原始数据
//   - string: MIME 类型 (Content-Type)
//   - int64: mtime (Last-Modified)
//   - string: etag (Content-Hash)
//   - error: 出错时返回错误
func (s *noteService) GetFileContent(ctx context.Context, uid int64, params *dto.NoteGetRequest) ([]byte, string, int64, string, error) {
	// 1. 获取仓库 ID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, "", 0, "", err
	}

	// 2. 确认路径哈希
	pathHash := params.PathHash
	if pathHash == "" {
		pathHash = util.EncodeHash32(params.Path)
	}

	// 3. 优先尝试从 Note 表获取 (笔记/文本内容)
	note, err := s.noteRepo.GetAllByPathHash(ctx, pathHash, vaultID, uid)
	if err == nil && note != nil {
		// 笔记内容固定识别为 markdown
		return []byte(note.Content), "text/markdown; charset=utf-8", note.Mtime, note.ContentHash, nil
	}

	// 4. 尝试从 File 表获取 (附件/二进制文件)
	if s.fileRepo != nil {
		file, err := s.fileRepo.GetByPathHash(ctx, pathHash, vaultID, uid)
		if err == nil && file != nil && file.Action != domain.FileActionDelete {
			// 读取物理文件内容
			content, err := os.ReadFile(file.SavePath)
			if err != nil {
				return nil, "", 0, "", fmt.Errorf("读取物理文件失败: %w", err)
			}

			// 识别文件 MIME 类型
			ext := filepath.Ext(params.Path)
			contentType := mime.TypeByExtension(ext)
			if contentType == "" {
				// 如果扩展名识别不到, 进行内容嗅探
				contentType = http.DetectContentType(content)
			}

			// File 表没有 ContentHash 或不确定, 实时计算
			etag := util.EncodeHash32(string(content))

			return content, contentType, file.Mtime, etag, nil
		}
	}

	return nil, "", 0, "", code.ErrorNoteNotFound
}

// 确保 noteService 实现了 NoteService 接口
var _ NoteService = (*noteService)(nil)
