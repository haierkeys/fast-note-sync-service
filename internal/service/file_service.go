// Package service 实现业务逻辑层
package service

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

// FileService 定义文件业务服务接口
type FileService interface {
	// Get 获取单条文件
	Get(ctx context.Context, uid int64, params *dto.FileGetRequest) (*dto.FileDTO, error)

	// UpdateCheck 检查文件是否需要更新
	UpdateCheck(ctx context.Context, uid int64, params *dto.FileUpdateCheckRequest) (string, *dto.FileDTO, error)

	// UploadCheck 检查文件上传（UpdateCheck 的别名，用于 WebSocket 上传检查）
	UploadCheck(ctx context.Context, uid int64, params *dto.FileUpdateCheckRequest) (string, *dto.FileDTO, error)

	// UpdateOrCreate 创建或修改文件
	UpdateOrCreate(ctx context.Context, uid int64, params *dto.FileUpdateRequest, mtimeCheck bool) (bool, *dto.FileDTO, error)

	// UploadComplete 完成文件上传（UpdateOrCreate 的别名，用于 WebSocket 上传完成）
	UploadComplete(ctx context.Context, uid int64, params *dto.FileUpdateRequest) (bool, *dto.FileDTO, error)

	// Delete 删除文件
	Delete(ctx context.Context, uid int64, params *dto.FileDeleteRequest) (*dto.FileDTO, error)

	// List 获取文件列表
	List(ctx context.Context, uid int64, params *dto.FileListRequest, pager *app.Pager) ([]*dto.FileDTO, int, error)

	// ListByLastTime 获取在 lastTime 之后更新的文件
	ListByLastTime(ctx context.Context, uid int64, params *dto.FileSyncRequest) ([]*dto.FileDTO, error)

	// Sync 同步文件（ListByLastTime 的别名，用于 WebSocket 同步）
	Sync(ctx context.Context, uid int64, params *dto.FileSyncRequest) ([]*dto.FileDTO, error)

	// CountSizeSum 统计 vault 中文件总数与总大小
	CountSizeSum(ctx context.Context, vaultID int64, uid int64) error

	// Cleanup 清理过期的软删除文件
	Cleanup(ctx context.Context, uid int64) error

	// CleanupByTime 按截止时间清理所有用户的过期软删除文件
	CleanupByTime(ctx context.Context, cutoffTime int64) error

	// ResolveEmbedLinks 解析内容中的嵌入链接
	ResolveEmbedLinks(ctx context.Context, uid int64, vaultName string, content string) (map[string]string, error)

	// GetContent 获取笔记或附件文件的原始内容
	GetContent(ctx context.Context, uid int64, params *dto.FileGetRequest) ([]byte, string, int64, string, error)

	// WithClient 设置客户端信息
	WithClient(name, version string) FileService
}

// fileService 实现 FileService 接口
type fileService struct {
	fileRepo     domain.FileRepository
	noteRepo     domain.NoteRepository
	vaultService VaultService
	sf           *singleflight.Group
	clientName   string
	clientVer    string
	config       *ServiceConfig
}

// NewFileService 创建 FileService 实例
func NewFileService(fileRepo domain.FileRepository, noteRepo domain.NoteRepository, vaultSvc VaultService, config *ServiceConfig) FileService {
	return &fileService{
		fileRepo:     fileRepo,
		noteRepo:     noteRepo,
		vaultService: vaultSvc,
		sf:           &singleflight.Group{},
		config:       config,
	}
}

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
		Size:             file.Size,
		Ctime:            file.Ctime,
		Mtime:            file.Mtime,
		UpdatedTimestamp: file.UpdatedTimestamp,
	}
}

// Get 获取单条文件
func (s *fileService) Get(ctx context.Context, uid int64, params *dto.FileGetRequest) (*dto.FileDTO, error) {
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

// UpdateCheck 检查文件是否需要更新
func (s *fileService) UpdateCheck(ctx context.Context, uid int64, params *dto.FileUpdateCheckRequest) (string, *dto.FileDTO, error) {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return "", nil, err
	}

	file, _ := s.fileRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if file != nil {
		fileDTO := s.domainToDTO(file)

		// 检查文件是否已删除
		if file.Action == domain.FileActionDelete {
			return "Create", nil, nil
		}

		// 检查内容是否一致
		if file.ContentHash == params.ContentHash {
			// 当用户 mtime 小于服务端 mtime 时，通知用户更新 mtime
			if params.Mtime < file.Mtime {
				return "UpdateMtime", fileDTO, nil
			} else if params.Mtime > file.Mtime {
				if err := s.fileRepo.UpdateMtime(ctx, params.Mtime, file.ID, uid); err != nil {
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

// UpdateOrCreate 创建或修改文件
func (s *fileService) UpdateOrCreate(ctx context.Context, uid int64, params *dto.FileUpdateRequest, mtimeCheck bool) (bool, *dto.FileDTO, error) {
	var isNew bool

	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return isNew, nil, err
	}

	file, _ := s.fileRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if file != nil {
		isNew = false
		// 检查内容是否一致,排除掉已被标记删除的文件
		if mtimeCheck && file.Action != domain.FileActionDelete && file.Mtime == params.Mtime && file.ContentHash == params.ContentHash {
			return isNew, s.domainToDTO(file), nil
		}
		// 检查内容是否一致但修改时间不同，则只更新修改时间
		if mtimeCheck && file.Mtime < params.Mtime && file.ContentHash == params.ContentHash {
			err := s.fileRepo.UpdateMtime(ctx, params.Mtime, file.ID, uid)
			if err != nil {
				return isNew, nil, code.ErrorDBQuery.WithDetails(err.Error())
			}
			file.Mtime = params.Mtime
			return isNew, s.domainToDTO(file), nil
		}

		// 设置 action
		var action domain.FileAction
		if file.Action == domain.FileActionDelete {
			action = domain.FileActionCreate
		} else {
			action = domain.FileActionModify
		}

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

		updated, err := s.fileRepo.Update(ctx, file, uid)
		if err != nil {
			return isNew, nil, code.ErrorDBQuery.WithDetails(err.Error())
		}

		go s.CountSizeSum(context.Background(), vaultID, uid)
		return isNew, s.domainToDTO(updated), nil
	}

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
		return isNew, nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	go s.CountSizeSum(context.Background(), vaultID, uid)
	return isNew, s.domainToDTO(created), nil
}

// Delete 删除文件
func (s *fileService) Delete(ctx context.Context, uid int64, params *dto.FileDeleteRequest) (*dto.FileDTO, error) {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	file, err := s.fileRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		return nil, err
	}

	// 更新为删除状态
	file.Action = domain.FileActionDelete
	file.ContentHash = ""
	file.SavePath = ""
	file.Size = 0
	file.Mtime = 0
	file.Ctime = 0

	updated, err := s.fileRepo.Update(ctx, file, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	go s.CountSizeSum(context.Background(), vaultID, uid)
	return s.domainToDTO(updated), nil
}

// List 获取文件列表
func (s *fileService) List(ctx context.Context, uid int64, params *dto.FileListRequest, pager *app.Pager) ([]*dto.FileDTO, int, error) {
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

// ListByLastTime 获取在 lastTime 之后更新的文件
func (s *fileService) ListByLastTime(ctx context.Context, uid int64, params *dto.FileSyncRequest) ([]*dto.FileDTO, error) {
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

// CountSizeSum 统计 vault 中文件总数与总大小
func (s *fileService) CountSizeSum(ctx context.Context, vaultID int64, uid int64) error {
	key := fmt.Sprintf("file_count_size_sum_%d_%d", uid, vaultID)
	_, err, _ := s.sf.Do(key, func() (any, error) {
		result, err := s.fileRepo.CountSizeSum(ctx, vaultID, uid)
		if err != nil {
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}
		return nil, s.vaultService.UpdateFileStats(ctx, result.Size, result.Count, vaultID, uid)
	})
	return err
}

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

// CleanupByTime 按截止时间清理所有用户的过期软删除文件
func (s *fileService) CleanupByTime(ctx context.Context, cutoffTime int64) error {
	return s.fileRepo.DeletePhysicalByTimeAll(ctx, cutoffTime)
}

// GetContent 获取笔记或附件文件的原始内容
// 返回值说明:
//   - []byte: 文件原始数据
//   - string: MIME 类型 (Content-Type)
//   - int64: mtime (Last-Modified)
//   - string: etag (Content-Hash)
//   - error: 出错时返回错误
func (s *fileService) GetContent(ctx context.Context, uid int64, params *dto.FileGetRequest) ([]byte, string, int64, string, error) {
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
	if s.noteRepo != nil {
		note, err := s.noteRepo.GetAllByPathHash(ctx, pathHash, vaultID, uid)
		if err == nil && note != nil {
			// 笔记内容固定识别为 markdown
			return []byte(note.Content), "text/markdown; charset=utf-8", note.Mtime, note.ContentHash, nil
		}
	}

	// 4. 尝试从 File 表获取 (附件/二进制文件)
	if s.fileRepo != nil {
		file, err := s.fileRepo.GetByPathHash(ctx, pathHash, vaultID, uid)
		if err == nil && file != nil && file.Action != domain.FileActionDelete {
			// 读取物理文件内容
			content, err := os.ReadFile(file.SavePath)
			if err != nil {
				return nil, "", 0, "", code.ErrorFileReadFailed.WithDetails(err.Error())
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

// ResolveEmbedLinks 解析内容中的嵌入链接
func (s *fileService) ResolveEmbedLinks(ctx context.Context, uid int64, vaultName string, content string) (map[string]string, error) {
	// 使用 VaultService.MustGetID 获取 VaultID
	vaultID, err := s.vaultService.MustGetID(ctx, uid, vaultName)
	if err != nil {
		return nil, err
	}

	// 正则匹配 ![[path|options]] 或 ![[path#anchor]]
	re := regexp.MustCompile(`!\[\[(.*?)\]\]`)
	matches := re.FindAllStringSubmatch(content, -1)

	resultMap := make(map[string]string)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		inner := match[1] // 标签内部内容，如 mm.jpg|100

		// 提取资源路径（移除尺寸 | 和锚点 # 之后的部分）
		resourcePath := inner
		if idx := strings.IndexAny(inner, "|#"); idx != -1 {
			resourcePath = inner[:idx]
		}
		resourcePath = strings.TrimSpace(resourcePath)

		if resourcePath == "" {
			continue
		}

		// 如果已经处理过这个路径，跳过
		if _, ok := resultMap[resourcePath]; ok {
			continue
		}

		// 搜索文件（LIKE 右匹配）
		file, err := s.fileRepo.GetByPathLike(ctx, resourcePath, vaultID, uid)
		if err == nil && file != nil {
			resultMap[resourcePath] = file.Path
		}
	}

	return resultMap, nil
}

// Sync 同步文件（ListByLastTime 的别名，用于 WebSocket 同步）
func (s *fileService) Sync(ctx context.Context, uid int64, params *dto.FileSyncRequest) ([]*dto.FileDTO, error) {
	return s.ListByLastTime(ctx, uid, params)
}

// UploadCheck 检查文件上传（UpdateCheck 的别名，用于 WebSocket 上传检查）
func (s *fileService) UploadCheck(ctx context.Context, uid int64, params *dto.FileUpdateCheckRequest) (string, *dto.FileDTO, error) {
	return s.UpdateCheck(ctx, uid, params)
}

// UploadComplete 完成文件上传（UpdateOrCreate 的别名，用于 WebSocket 上传完成）
func (s *fileService) UploadComplete(ctx context.Context, uid int64, params *dto.FileUpdateRequest) (bool, *dto.FileDTO, error) {
	return s.UpdateOrCreate(ctx, uid, params, true)
}

// WithClient 设置客户端信息，返回新的 FileService 实例
func (s *fileService) WithClient(name, version string) FileService {
	return &fileService{
		fileRepo:     s.fileRepo,
		noteRepo:     s.noteRepo,
		vaultService: s.vaultService,
		sf:           s.sf,
		clientName:   name,
		clientVer:    version,
		config:       s.config,
	}
}

// 确保 fileService 实现了 FileService 接口
var _ FileService = (*fileService)(nil)
