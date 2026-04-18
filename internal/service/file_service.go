// Package service implements the business logic layer
// Package service 实现业务逻辑层
package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
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

// FileService defines the file business service interface
// FileService 定义文件业务服务接口
type FileService interface {
	// Get retrieves a single file
	// Get 获取单条文件
	Get(ctx context.Context, uid int64, params *dto.FileGetRequest) (*dto.FileDTO, error)

	// UpdateCheck checks if file needs updating
	// UpdateCheck 检查文件是否需要更新
	UpdateCheck(ctx context.Context, uid int64, params *dto.FileUpdateCheckRequest) (string, *dto.FileDTO, error)

	// UploadCheck checks file upload (alias for UpdateCheck, used for WebSocket upload check)
	// UploadCheck 检查文件上传（UpdateCheck 的别名，用于 WebSocket 上传检查）
	UploadCheck(ctx context.Context, uid int64, params *dto.FileUpdateCheckRequest) (string, *dto.FileDTO, error)

	// UpdateOrCreate creates or modifies a file
	// UpdateOrCreate 创建或修改文件
	UpdateOrCreate(ctx context.Context, uid int64, params *dto.FileUpdateRequest, mtimeCheck bool) (bool, *dto.FileDTO, error)

	// UploadComplete completes file upload (alias for UpdateOrCreate, used for WebSocket upload completion)
	// UploadComplete 完成文件上传（UpdateOrCreate 的别名，用于 WebSocket 上传完成）
	UploadComplete(ctx context.Context, uid int64, params *dto.FileUpdateRequest) (bool, *dto.FileDTO, error)

	// Delete deletes a file
	// Delete 删除文件
	Delete(ctx context.Context, uid int64, params *dto.FileDeleteRequest) (*dto.FileDTO, error)

	// List retrieves file list
	// List 获取文件列表
	List(ctx context.Context, uid int64, params *dto.FileListRequest, pager *app.Pager) ([]*dto.FileDTO, int, error)

	// ListByLastTime retrieves files updated after lastTime
	// ListByLastTime 获取在 lastTime 之后更新的文件
	ListByLastTime(ctx context.Context, uid int64, params *dto.FileSyncRequest) ([]*dto.FileDTO, error)

	// CountSizeSum counts total number and total size of files in a vault
	// CountSizeSum 统计 vault 中文件总数与总大小
	CountSizeSum(ctx context.Context, vaultID int64, uid int64) error

	// Cleanup cleans up expired soft-deleted files
	// Cleanup 清理过期的软删除文件
	Cleanup(ctx context.Context, uid int64) error

	// CleanupByTime cleans up expired soft-deleted files for all users by cutoff time
	// CleanupByTime 按截止时间清理所有用户的过期软删除文件
	CleanupByTime(ctx context.Context, cutoffTime int64) error

	// ResolveEmbedLinks resolves local file links in note content
	// ResolveEmbedLinks 解析笔记内容中的本地文件链接
	ResolveEmbedLinks(ctx context.Context, uid int64, vaultName string, notePath string, content string) (map[string]string, error)

	// GetContent retrieves raw content of note or attachment file
	// GetContent 获取笔记或附件文件的原始内容
	GetContent(ctx context.Context, uid int64, params *dto.FileGetRequest) (io.ReadCloser, string, int64, string, error)

	// GetContentInfo retrieves file metadata and path for zero-copy download
	// GetContentInfo 获取文件的元数据和路径，用于零拷贝下载
	GetContentInfo(ctx context.Context, uid int64, params *dto.FileGetRequest) (savePath string, contentType string, mtime int64, etag string, fileName string, err error)

	// Restore restores a file (from recycle bin)
	// Restore 恢复文件（从回收站恢复）
	Restore(ctx context.Context, uid int64, params *dto.FileRestoreRequest) (*dto.FileDTO, error)
	// Rename 重命名文件
	Rename(ctx context.Context, uid int64, params *dto.FileRenameRequest) (*dto.FileDTO, *dto.FileDTO, error)
	// WithClient sets client info
	// WithClient 设置客户端信息
	WithClient(name, version string) FileService

	// RecycleClear 清理回收站
	RecycleClear(ctx context.Context, uid int64, params *dto.FileRecycleClearRequest) error

	// CleanDuplicateFiles 清理重复的文件记录
	CleanDuplicateFiles(ctx context.Context, uid int64, vaultID int64) error
}

// fileService implementation of FileService interface
// fileService 实现 FileService 接口
type fileService struct {
	fileRepo       domain.FileRepository // File repository // 文件仓库
	noteRepo       domain.NoteRepository // Note repository // 笔记仓库
	vaultService   VaultService          // Vault service // 仓库服务
	folderService  FolderService         // Folder service // 文件夹服务
	sf             *singleflight.Group   // Singleflight group // 并发请求合并组
	clientName     string                // Client name // 客户端名称
	clientVer      string                // Client version // 客户端版本
	config         *ServiceConfig        // Service configuration // 服务配置
	backupService  BackupService         // Backup service // 备份服务
	gitSyncService GitSyncService        // Git sync service // Git 同步服务
	countTimers    *sync.Map             // Timers for CountSizeSum debounce // CountSizeSum 防抖计时器
}

// NewFileService creates FileService instance
// NewFileService 创建 FileService 实例
func NewFileService(fileRepo domain.FileRepository, noteRepo domain.NoteRepository, vaultSvc VaultService, folderSvc FolderService, backupSvc BackupService, gitSyncSvc GitSyncService, config *ServiceConfig) FileService {
	return &fileService{
		fileRepo:       fileRepo,
		noteRepo:       noteRepo,
		vaultService:   vaultSvc,
		folderService:  folderSvc,
		backupService:  backupSvc,
		gitSyncService: gitSyncSvc,
		sf:             &singleflight.Group{},
		config:         config,
		countTimers:    &sync.Map{},
	}
}

// domainToDTO converts domain model to DTO
// domainToDTO 将领域模型转换为 DTO
func (s *fileService) domainToDTO(file *domain.File) *dto.FileDTO {
	if file == nil {
		return nil
	}
	return &dto.FileDTO{
		ID:               file.ID,
		Action:           string(file.Action),
		Path:             file.Path,
		PathHash:         file.PathHash,
		ContentHash:      file.ContentHash,
		SavePath:         file.SavePath,
		Rename:           file.Rename,
		Size:             file.Size,
		Ctime:            file.Ctime,
		Mtime:            file.Mtime,
		UpdatedTimestamp: file.UpdatedTimestamp,
	}
}

// Get retrieves a single file
// Get 获取单条文件
func (s *fileService) Get(ctx context.Context, uid int64, params *dto.FileGetRequest) (*dto.FileDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	file, err := s.fileRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		return nil, err
	}

	return s.domainToDTO(file), nil
}

// UpdateCheck checks if file needs updating
// UpdateCheck 检查文件是否需要更新
func (s *fileService) UpdateCheck(ctx context.Context, uid int64, params *dto.FileUpdateCheckRequest) (string, *dto.FileDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return "", nil, err
	}

	file, _ := s.fileRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if file != nil {
		fileDTO := s.domainToDTO(file)

		// Check if file is deleted
		// 检查文件是否已删除
		if file.Action == domain.FileActionDelete {
			return "Create", nil, nil
		}

		// Check if content is consistent
		// 检查内容是否一致
		if file.ContentHash == params.ContentHash {
			// Notify user to update mtime when user mtime is less than server mtime
			// 当用户 mtime 小于服务端 mtime 时，通知用户更新 mtime
			if params.Mtime < file.Mtime {
				return "UpdateMtime", fileDTO, nil
			} else if params.Mtime > file.Mtime {
				if err := s.fileRepo.UpdateMtime(ctx, params.Mtime, file.ID, uid); err != nil {
					// Non-critical update failed, log warning but do not block flow
					// 非关键更新失败，记录警告日志但不阻断流程
					zap.L().Warn("UpdateMtime failed for file",
						zap.Int64(logger.FieldUID, uid),
						zap.Int64("fileId", file.ID),
						zap.Int64("mtime", params.Mtime),
						zap.String(logger.FieldMethod, "FileService.UpdateCheck"),
						zap.Error(err),
					)
				}
			}
			return "", fileDTO, nil
		}
		return "UpdateContent", fileDTO, nil
	}
	return "Create", nil, nil
}

// UpdateOrCreate creates or modifies a file
// UpdateOrCreate 创建或修改文件
func (s *fileService) UpdateOrCreate(ctx context.Context, uid int64, params *dto.FileUpdateRequest, mtimeCheck bool) (bool, *dto.FileDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return false, nil, err
	}

	key := fmt.Sprintf("update_or_create_%d_%d_%s", uid, vaultID, params.PathHash)
	type result struct {
		isNew bool
		dto   *dto.FileDTO
	}

	val, err, _ := s.sf.Do(key, func() (any, error) {
		var isNew bool
		file, _ := s.fileRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)

		if file != nil {
			isNew = false
			// Check if content is consistent, excluding files marked as deleted
			// 检查内容是否一致,排除掉已被标记删除的文件
			if mtimeCheck && file.Action != domain.FileActionDelete && file.Mtime == params.Mtime && file.ContentHash == params.ContentHash {
				return &result{isNew: isNew, dto: s.domainToDTO(file)}, nil
			}

			// If content is consistent but modification time is different, only update modification time
			// 检查内容是否一致但修改时间不同，则只更新修改时间
			if mtimeCheck && file.Mtime < params.Mtime && file.ContentHash == params.ContentHash {
				err := s.fileRepo.UpdateActionMtime(ctx, domain.FileActionModify, params.Mtime, file.ID, uid)
				if err != nil {
					return nil, code.ErrorDBQuery.WithDetails(err.Error())
				}
				file.Mtime = params.Mtime
				if s.backupService != nil {
					go s.backupService.NotifyUpdated(uid)
				}
				if s.gitSyncService != nil {
					go s.gitSyncService.NotifyUpdated(uid, vaultID)
				}
				return &result{isNew: isNew, dto: s.domainToDTO(file)}, nil
			}

			// Set action
			// 设置 action
			var action domain.FileAction
			if file.Action == domain.FileActionDelete {
				action = domain.FileActionCreate
			} else {
				action = domain.FileActionModify
			}

			// Update file
			// 更新文件
			file.VaultID = vaultID
			file.Path = params.Path
			file.PathHash = params.PathHash
			file.ContentHash = params.ContentHash
			file.SavePath = params.SavePath
			file.Size = params.Size
			file.Mtime = params.Mtime
			file.Ctime = params.Ctime
			file.Action = action
			file.Rename = 0

			updated, err := s.fileRepo.Update(ctx, file, uid)
			if err != nil {
				return nil, code.ErrorDBQuery.WithDetails(err.Error())
			}

			go s.CountSizeSum(context.Background(), vaultID, uid)
			go s.folderService.SyncResourceFID(context.Background(), uid, vaultID, nil, []int64{updated.ID})
			if s.backupService != nil {
				go s.backupService.NotifyUpdated(uid)
			}
			if s.gitSyncService != nil {
				go s.gitSyncService.NotifyUpdated(uid, vaultID)
			}
			return &result{isNew: isNew, dto: s.domainToDTO(updated)}, nil
		}

		// Create new file
		// 创建新文件
		isNew = true
		newFile := &domain.File{
			VaultID:     vaultID,
			Path:        params.Path,
			PathHash:    params.PathHash,
			ContentHash: params.ContentHash,
			SavePath:    params.SavePath,
			Size:        params.Size,
			Mtime:       params.Mtime,
			Ctime:       params.Ctime,
			Action:      domain.FileActionCreate,
		}

		created, err := s.fileRepo.Create(ctx, newFile, uid)
		if err != nil {
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}

		go s.CountSizeSum(context.Background(), vaultID, uid)
		go s.folderService.SyncResourceFID(context.Background(), uid, vaultID, nil, []int64{created.ID})
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

// Delete deletes a file
// Delete 删除文件
func (s *fileService) Delete(ctx context.Context, uid int64, params *dto.FileDeleteRequest) (*dto.FileDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	file, err := s.fileRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		return nil, err
	}

	// Update to deleted status
	// 更新为删除状态
	file.Action = domain.FileActionDelete
	file.Rename = 0

	updated, err := s.fileRepo.Update(ctx, file, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	go s.CountSizeSum(context.Background(), vaultID, uid)
	if s.backupService != nil {
		go s.backupService.NotifyUpdated(uid)
	}
	if s.gitSyncService != nil {
		go s.gitSyncService.NotifyUpdated(uid, vaultID)
	}
	return s.domainToDTO(updated), nil
}

// Restore restores a file (from recycle bin)
// Restore 恢复文件（从回收站恢复）
func (s *fileService) Restore(ctx context.Context, uid int64, params *dto.FileRestoreRequest) (*dto.FileDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	// Calculate PathHash if not provided
	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// Get file from recycle bin
	file, err := s.fileRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		return nil, code.ErrorNoteNotFound
	}

	// Check if file is deleted
	if file.Action != domain.FileActionDelete {
		return nil, code.ErrorNoteNotFound
	}

	// Update to modified status and update modification time
	file.Action = domain.FileActionModify
	file.Mtime = time.Now().UnixMilli()
	file.Rename = 0

	updated, err := s.fileRepo.Update(ctx, file, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	go s.CountSizeSum(context.Background(), vaultID, uid)
	if s.backupService != nil {
		go s.backupService.NotifyUpdated(uid)
	}
	if s.gitSyncService != nil {
		go s.gitSyncService.NotifyUpdated(uid, vaultID)
	}
	return s.domainToDTO(updated), nil
}

// List retrieves file list
// List 获取文件列表
func (s *fileService) List(ctx context.Context, uid int64, params *dto.FileListRequest, pager *app.Pager) ([]*dto.FileDTO, int, error) {
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, 0, err
	}

	files, err := s.fileRepo.List(ctx, vaultID, pager.Page, pager.PageSize, uid, params.Keyword, params.IsRecycle, params.SortBy, params.SortOrder)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}

	count, err := s.fileRepo.ListCount(ctx, vaultID, uid, params.Keyword, params.IsRecycle)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var result []*dto.FileDTO
	for _, f := range files {
		result = append(result, s.domainToDTO(f))
	}

	return result, int(count), nil
}

// ListByLastTime retrieves files updated after lastTime
// ListByLastTime 获取在 lastTime 之后更新的文件
func (s *fileService) ListByLastTime(ctx context.Context, uid int64, params *dto.FileSyncRequest) ([]*dto.FileDTO, error) {
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	files, err := s.fileRepo.ListByUpdatedTimestamp(ctx, params.LastTime, vaultID, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var results []*dto.FileDTO
	cacheList := make(map[string]bool)
	for _, file := range files {
		if cacheList[file.PathHash] {
			continue
		}
		results = append(results, s.domainToDTO(file))
		cacheList[file.PathHash] = true
	}

	return results, nil
}

// CountSizeSum counts total number and total size of files in a vault
// CountSizeSum 统计 vault 中文件总数与总大小
func (s *fileService) CountSizeSum(ctx context.Context, vaultID int64, uid int64) error {
	key := fmt.Sprintf("%d_%d", uid, vaultID)

	// Debounce: 10 seconds delay. If a new request comes within 10s, reset the timer.
	// 防抖：10秒延迟。如果10秒内有新请求，重置计时器。
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
			result, err := s.fileRepo.CountSizeSum(context.Background(), vaultID, uid)
			if err != nil {
				return nil, code.ErrorDBQuery.WithDetails(err.Error())
			}
			// Update vault stats, and removed the nested SyncResourceFID call
			// 更新仓库统计，并移除了嵌套的 SyncResourceFID 调用
			return nil, s.vaultService.UpdateFileStats(context.Background(), result.Size, result.Count, vaultID, uid)
		})
	})

	s.countTimers.Store(key, timer)
	return nil
}

// Cleanup cleans up expired soft-deleted files
// Cleanup 清理过期的软删除文件
func (s *fileService) Cleanup(ctx context.Context, uid int64) error {
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
	return s.fileRepo.DeletePhysicalByTime(ctx, cutoffTime, uid)
}

// CleanupByTime cleans up expired soft-deleted files for all users by cutoff time
// CleanupByTime 按截止时间清理所有用户的过期软删除文件
func (s *fileService) CleanupByTime(ctx context.Context, cutoffTime int64) error {
	return s.fileRepo.DeletePhysicalByTimeAll(ctx, cutoffTime)
}

// GetContent retrieves raw content of note or attachment file
// GetContent 获取笔记或附件文件的原始内容
// Return value description:
// 返回值说明:
//   - []byte: Raw file data // 文件原始数据
//   - string: MIME type (Content-Type) // MIME 类型 (Content-Type)
//   - int64: mtime (Last-Modified) // mtime (Last-Modified)
//   - string: etag (Content-Hash) // etag (Content-Hash)
//   - error: Error on failure // 出错时返回错误
func (s *fileService) GetContent(ctx context.Context, uid int64, params *dto.FileGetRequest) (io.ReadCloser, string, int64, string, error) {
	// 1. Get vault ID
	// 1. 获取仓库 ID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, "", 0, "", err
	}

	// 2. Confirm path hash
	// 2. 确认路径哈希
	pathHash := params.PathHash
	if pathHash == "" {
		pathHash = util.EncodeHash32(params.Path)
	}

	// 4. Attempt to get from File table (attachment/binary file)
	// 4. 尝试从 File 表获取 (附件/二进制文件)
	if s.fileRepo != nil {
		file, err := s.fileRepo.GetByPathHash(ctx, pathHash, vaultID, uid)
		if err == nil && file != nil {
			// Identify file MIME type
			// 识别文件 MIME 类型
			ext := filepath.Ext(params.Path)
			contentType := mime.TypeByExtension(ext)
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			// Use file's content hash as ETag from DB if available
			// 使用 DB 中的 ContentHash 作为 ETag
			etag := file.ContentHash
			if etag == "" {
				etag = file.PathHash
			}

			// Open file for streaming
			// 打开文件用于流式传输
			f, err := os.Open(file.SavePath)
			if err != nil {
				return nil, "", 0, "", code.ErrorFileReadFailed.WithDetails(err.Error())
			}

			return f, contentType, file.Mtime, etag, nil
		}
	}

	return nil, "", 0, "", code.ErrorNoteNotFound
}

// GetContentInfo retrieves file metadata and path for zero-copy download
// GetContentInfo 获取文件的元数据 and 路径，用于零拷贝下载
func (s *fileService) GetContentInfo(ctx context.Context, uid int64, params *dto.FileGetRequest) (string, string, int64, string, string, error) {
	// 1. Get vault ID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return "", "", 0, "", "", err
	}

	// 2. Confirm path hash
	pathHash := params.PathHash
	if pathHash == "" {
		pathHash = util.EncodeHash32(params.Path)
	}

	// 3. Attempt to get from File table
	if s.fileRepo != nil {
		file, err := s.fileRepo.GetByPathHash(ctx, pathHash, vaultID, uid)
		if err == nil && file != nil {
			// Check IsRecycle support
			if params.IsRecycle {
				if file.Action != domain.FileActionDelete {
					return "", "", 0, "", "", code.ErrorFileNotFound
				}
			} else {
				if file.Action == domain.FileActionDelete {
					return "", "", 0, "", "", code.ErrorFileNotFound
				}
			}
			// Identify file MIME type
			ext := filepath.Ext(file.Path)
			contentType := mime.TypeByExtension(ext)
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			// Use file's content hash as ETag
			etag := file.ContentHash
			if etag == "" {
				etag = file.PathHash
			}

			return file.SavePath, contentType, file.Mtime, etag, filepath.Base(file.Path), nil
		}
	}

	return "", "", 0, "", "", code.ErrorNoteNotFound
}

// ResolveEmbedLinks resolves local file links in note content
// ResolveEmbedLinks 解析笔记内容中的本地文件链接
func (s *fileService) ResolveEmbedLinks(ctx context.Context, uid int64, vaultName string, notePath string, content string) (map[string]string, error) {
	// Use VaultService.MustGetID to retrieve VaultID
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, vaultName)
	if err != nil {
		return nil, err
	}

	rawRefs := extractSharedNoteFileRefs(content)
	resultMap := make(map[string]string, len(rawRefs))
	for _, rawRef := range rawRefs {
		file, err := s.resolveNoteFileReference(ctx, uid, vaultID, notePath, rawRef)
		if err != nil {
			return nil, err
		}
		if file != nil {
			resultMap[rawRef] = file.Path
		}
	}

	return resultMap, nil
}

func (s *fileService) resolveNoteFileReference(ctx context.Context, uid int64, vaultID int64, notePath string, rawRef string) (*domain.File, error) {
	ref := strings.TrimSpace(rawRef)
	if !isLocalSharePath(ref) {
		return nil, nil
	}

	for _, candidate := range buildSharePathCandidates(notePath, ref) {
		file, err := s.fileRepo.GetByPath(ctx, candidate, vaultID, uid)
		if err == nil && file != nil && file.Action != domain.FileActionDelete {
			return file, nil
		}
	}

	normalizedRef := normalizeShareVaultPath(ref)
	if normalizedRef != "" && !strings.Contains(normalizedRef, "/") {
		file, err := s.fileRepo.GetByPathLike(ctx, normalizedRef, vaultID, uid)
		if err == nil && file != nil && file.Action != domain.FileActionDelete {
			return file, nil
		}
	}

	return nil, nil
}

// Sync syncs files (alias for ListByLastTime, used for WebSocket sync)
// Sync 同步文件（ListByLastTime 的别名，用于 WebSocket 同步）
func (s *fileService) Sync(ctx context.Context, uid int64, params *dto.FileSyncRequest) ([]*dto.FileDTO, error) {
	return s.ListByLastTime(ctx, uid, params)
}

// UploadCheck checks file upload (alias for UpdateCheck, used for WebSocket upload check)
// UploadCheck 检查文件上传（UpdateCheck 的别名，用于 WebSocket 上传检查）
func (s *fileService) UploadCheck(ctx context.Context, uid int64, params *dto.FileUpdateCheckRequest) (string, *dto.FileDTO, error) {
	return s.UpdateCheck(ctx, uid, params)
}

// UploadComplete completes file upload (alias for UpdateOrCreate, used for WebSocket upload completion)
// UploadComplete 完成文件上传（UpdateOrCreate 的别名，用于 WebSocket 上传完成）
func (s *fileService) UploadComplete(ctx context.Context, uid int64, params *dto.FileUpdateRequest) (bool, *dto.FileDTO, error) {
	return s.UpdateOrCreate(ctx, uid, params, true)
}

// Rename renames a file
// Rename 重命名文件
func (s *fileService) Rename(ctx context.Context, uid int64, params *dto.FileRenameRequest) (*dto.FileDTO, *dto.FileDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, nil, err
	}

	newPath := strings.Trim(params.Path, "/")
	newPathHash := params.PathHash
	if newPathHash == "" {
		newPathHash = util.EncodeHash32(newPath)
	}

	oldPath := strings.Trim(params.OldPath, "/")
	oldPathHash := params.OldPathHash
	if oldPathHash == "" {
		oldPathHash = util.EncodeHash32(oldPath)
	}

	key := fmt.Sprintf("rename_%d_%d_%s_%s", uid, vaultID, oldPathHash, newPathHash)
	type result struct {
		oldFile *dto.FileDTO
		newFile *dto.FileDTO
	}

	val, err, _ := s.sf.Do(key, func() (any, error) {
		// 1. 判断目标路径是否存在有效文件
		existFile, _ := s.fileRepo.GetByPathHash(ctx, newPathHash, vaultID, uid)
		if existFile != nil && existFile.Action != domain.FileActionDelete {
			return nil, code.ErrorFileExist
		}

		// 2. 获取旧文件
		f, err := s.fileRepo.GetByPathHash(ctx, oldPathHash, vaultID, uid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, code.ErrorFileNotFound
			}
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}

		// 3. 标记旧文件删除
		f.Action = domain.FileActionDelete
		f.Rename = 1
		f.UpdatedTimestamp = timex.Now().UnixMilli()
		oldFile, err := s.fileRepo.Update(ctx, f, uid)
		if err != nil {
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}

		// 4. 新建或复用文件记录
		var newFileCreated *domain.File
		if existFile != nil {
			// 复用已删除的记录
			existFile.Action = domain.FileActionCreate
			existFile.Path = newPath
			existFile.PathHash = newPathHash
			newPathDir := ""
			if idx := strings.LastIndex(newPath, "/"); idx >= 0 {
				newPathDir = newPath[:idx]
			}
			existFile.FID, _ = s.folderService.EnsurePathFID(ctx, uid, vaultID, newPathDir)
			existFile.ContentHash = f.ContentHash
			existFile.SavePath = f.SavePath
			existFile.Size = f.Size
			existFile.Mtime = timex.Now().UnixMilli()
			existFile.UpdatedTimestamp = timex.Now().UnixMilli()
			newFileCreated, err = s.fileRepo.Update(ctx, existFile, uid)
		} else {
			// 创建新记录
			newFile := &domain.File{
				VaultID:          vaultID,
				Action:           domain.FileActionCreate,
				Path:             newPath,
				PathHash:         newPathHash,
				FID:              f.FID,
				Ctime:            f.Ctime,
				Mtime:            timex.Now().UnixMilli(),
				UpdatedTimestamp: timex.Now().UnixMilli(),
				ContentHash:      f.ContentHash,
				SavePath:         f.SavePath,
				Size:             f.Size,
			}
			newPathDir := ""
			if idx := strings.LastIndex(newPath, "/"); idx >= 0 {
				newPathDir = newPath[:idx]
			}
			// 确保 FID 正确
			newFile.FID, _ = s.folderService.EnsurePathFID(ctx, uid, vaultID, newPathDir)
			newFileCreated, err = s.fileRepo.Create(ctx, newFile, uid)
		}

		if err != nil {
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}

		// 修正目录FID
		go s.folderService.SyncResourceFID(context.Background(), uid, vaultID, nil, []int64{newFileCreated.ID})

		if s.backupService != nil {
			go s.backupService.NotifyUpdated(uid)
		}
		if s.gitSyncService != nil {
			go s.gitSyncService.NotifyUpdated(uid, vaultID)
		}

		return &result{oldFile: s.domainToDTO(oldFile), newFile: s.domainToDTO(newFileCreated)}, nil
	})

	if err != nil {
		return nil, nil, err
	}

	res := val.(*result)
	return res.oldFile, res.newFile, nil
}

// WithClient sets client info, returns new FileService instance
// WithClient 设置客户端信息，返回新的 FileService 实例
func (s *fileService) WithClient(name, version string) FileService {
	return &fileService{
		fileRepo:       s.fileRepo,
		noteRepo:       s.noteRepo,
		vaultService:   s.vaultService,
		folderService:  s.folderService,
		sf:             s.sf,
		clientName:     name,
		clientVer:      version,
		config:         s.config,
		backupService:  s.backupService,
		gitSyncService: s.gitSyncService,
		countTimers:    s.countTimers, // Share the same timer map // 共享同一个计时器 map
	}
}

// RecycleClear 清理回收站
func (s *fileService) RecycleClear(ctx context.Context, uid int64, params *dto.FileRecycleClearRequest) error {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return err
	}

	if params.Path != "" && params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	err = s.fileRepo.RecycleClear(ctx, params.Path, params.PathHash, vaultID, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	go s.CountSizeSum(context.Background(), vaultID, uid)
	return nil
}

// CleanDuplicateFiles 清理重复的文件记录
func (s *fileService) CleanDuplicateFiles(ctx context.Context, uid int64, vaultID int64) error {
	// 获取所有文件（包含已删除，以便全局去重）
	files, err := s.fileRepo.ListByUpdatedTimestamp(ctx, 0, vaultID, uid)
	if err != nil {
		return err
	}

	// 按 PathHash 分组
	grouped := make(map[string][]*domain.File)
	for _, f := range files {
		grouped[f.PathHash] = append(grouped[f.PathHash], f)
	}

	for pathHash, list := range grouped {
		if len(list) <= 1 {
			continue
		}

		//保留规则：
		// 1. 优先保留 Action != delete 的记录
		// 2. 如果有多个活跃记录，保留 UpdatedTimestamp 最大（最新）的一条
		// 3. 如果时间戳一致，保留 ID 最大的记录

		var bestFile *domain.File
		for _, f := range list {
			if bestFile == nil {
				bestFile = f
				continue
			}

			// 比较逻辑
			isBetter := false
			if f.Action != domain.FileActionDelete && bestFile.Action == domain.FileActionDelete {
				isBetter = true
			} else if f.Action == bestFile.Action {
				if f.UpdatedTimestamp > bestFile.UpdatedTimestamp {
					isBetter = true
				} else if f.UpdatedTimestamp == bestFile.UpdatedTimestamp && f.ID > bestFile.ID {
					isBetter = true
				}
			}

			if isBetter {
				bestFile = f
			}
		}

		// 删除非 bestFile 的所有记录
		for _, f := range list {
			if f.ID != bestFile.ID {
				// 清除 singleflight 缓存，防止残留
				s.sf.Forget(fmt.Sprintf("update_or_create_%d_%d_%s", uid, vaultID, pathHash))
				s.sf.Forget(fmt.Sprintf("rename_%d_%d_%s", uid, vaultID, pathHash))
				
				_ = s.fileRepo.Delete(ctx, f.ID, uid)
			}
		}
	}

	return nil
}

// Verify fileService implements FileService interface
// 确保 fileService 实现了 FileService interface
var _ FileService = (*fileService)(nil)
