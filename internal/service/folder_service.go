package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

// FolderService 文件夹业务服务接口
type FolderService interface {
	Get(ctx context.Context, uid int64, params *dto.FolderGetRequest) (*dto.FolderDTO, error)
	List(ctx context.Context, uid int64, params *dto.FolderListRequest) ([]*dto.FolderDTO, error)
	ListByUpdatedTimestamp(ctx context.Context, uid int64, vault string, lastTime int64) ([]*dto.FolderDTO, error)
	UpdateOrCreate(ctx context.Context, uid int64, params *dto.FolderCreateRequest) (*dto.FolderDTO, error)
	Delete(ctx context.Context, uid int64, params *dto.FolderDeleteRequest) error
	Rename(ctx context.Context, uid int64, params *dto.FolderRenameRequest) (*dto.FolderDTO, *dto.FolderDTO, error)
	ListNotes(ctx context.Context, uid int64, params *dto.FolderContentRequest, pager *app.Pager) ([]*dto.NoteNoContentDTO, int, error)
	ListFiles(ctx context.Context, uid int64, params *dto.FolderContentRequest, pager *app.Pager) ([]*dto.FileDTO, int, error)
	EnsurePathFID(ctx context.Context, uid int64, vaultID int64, path string) (int64, error)
	SyncResourceFID(ctx context.Context, uid int64, vaultID int64, noteIDs []int64, fileIDs []int64) error
}

type folderService struct {
	folderRepo   domain.FolderRepository
	noteRepo     domain.NoteRepository
	fileRepo     domain.FileRepository
	vaultService VaultService
	sf           singleflight.Group
}

func NewFolderService(folderRepo domain.FolderRepository, noteRepo domain.NoteRepository, fileRepo domain.FileRepository, vaultSvc VaultService) FolderService {
	return &folderService{
		folderRepo:   folderRepo,
		noteRepo:     noteRepo,
		fileRepo:     fileRepo,
		vaultService: vaultSvc,
		sf:           singleflight.Group{},
	}
}

func (s *folderService) domainToDTO(f *domain.Folder) *dto.FolderDTO {
	if f == nil {
		return nil
	}
	return &dto.FolderDTO{
		ID:               f.ID,
		Action:           string(f.Action),
		Path:             f.Path,
		PathHash:         f.PathHash,
		Level:            f.Level,
		FID:              f.FID,
		Ctime:            f.Ctime,
		Mtime:            f.Mtime,
		UpdatedTimestamp: f.UpdatedTimestamp,
		UpdatedAt:        timex.Time(f.UpdatedAt),
		CreatedAt:        timex.Time(f.CreatedAt),
	}
}

func (s *folderService) List(ctx context.Context, uid int64, params *dto.FolderListRequest) ([]*dto.FolderDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	var fid int64 = 0
	if params.PathHash != "" {
		f, err := s.folderRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, code.ErrorFolderNotFound
			}
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}
		fid = f.ID
	} else if params.Path != "" {
		pathHash := util.EncodeHash32(params.Path)
		f, err := s.folderRepo.GetByPathHash(ctx, pathHash, vaultID, uid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, code.ErrorFolderNotFound
			}
			return nil, code.ErrorDBQuery.WithDetails(err.Error())
		}
		fid = f.ID
	}

	folders, err := s.folderRepo.GetByFID(ctx, fid, vaultID, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var res []*dto.FolderDTO
	for _, f := range folders {
		res = append(res, s.domainToDTO(f))
	}
	return res, nil
}

func (s *folderService) UpdateOrCreate(ctx context.Context, uid int64, params *dto.FolderCreateRequest) (*dto.FolderDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	path := strings.Trim(params.Path, "/")
	if path == "" {
		return nil, code.ErrorInvalidParams.WithDetails("path cannot be empty")
	}

	// 确保父级目录存在
	parts := strings.Split(path, "/")
	var currentFID int64 = 0
	var lastFolder *domain.Folder

	for i, _ := range parts {
		currentPath := strings.Join(parts[:i+1], "/")
		pathHash := util.EncodeHash32(currentPath)

		f, err := s.folderRepo.GetByPathHash(ctx, pathHash, vaultID, uid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 创建缺失的目录
				newFolder := &domain.Folder{
					VaultID:  vaultID,
					Action:   domain.FolderActionCreate,
					Path:     currentPath,
					PathHash: pathHash,
					Level:    int64(i + 1),
					FID:      currentFID,
					Ctime:    time.Now().UnixMilli(),
					Mtime:    time.Now().UnixMilli(),
				}
				f, err = s.folderRepo.Create(ctx, newFolder, uid)
				if err != nil {
					return nil, code.ErrorDBQuery.WithDetails(err.Error())
				}
			} else {
				return nil, code.ErrorDBQuery.WithDetails(err.Error())
			}
		} else if f.Action == domain.FolderActionDelete {
			// 如果已被删除，恢复它
			f.Action = domain.FolderActionCreate
			f.Ctime = time.Now().UnixMilli()
			f.Mtime = time.Now().UnixMilli()
			f, err = s.folderRepo.Update(ctx, f, uid)
			if err != nil {
				return nil, code.ErrorDBQuery.WithDetails(err.Error())
			}
		}
		currentFID = f.ID
		lastFolder = f
	}

	return s.domainToDTO(lastFolder), nil
}

func (s *folderService) Delete(ctx context.Context, uid int64, params *dto.FolderDeleteRequest) error {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return err
	}

	pathHash := params.PathHash
	if pathHash == "" {
		pathHash = util.EncodeHash32(params.Path)
	}

	f, err := s.folderRepo.GetByPathHash(ctx, pathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.ErrorFolderNotFound
		}
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 软删除
	f.Action = domain.FolderActionDelete
	f.UpdatedTimestamp = time.Now().UnixMilli()
	_, err = s.folderRepo.Update(ctx, f, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	return nil
}

func (s *folderService) ListByUpdatedTimestamp(ctx context.Context, uid int64, vault string, lastTime int64) ([]*dto.FolderDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, vault)
	if err != nil {
		return nil, err
	}

	folders, err := s.folderRepo.ListByUpdatedTimestamp(ctx, lastTime, vaultID, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var res []*dto.FolderDTO
	cache := make(map[string]bool)
	for _, f := range folders {
		if cache[f.PathHash] {
			continue
		}
		res = append(res, s.domainToDTO(f))
		cache[f.PathHash] = true
	}

	return res, nil
}

func (s *folderService) Rename(ctx context.Context, uid int64, params *dto.FolderRenameRequest) (*dto.FolderDTO, *dto.FolderDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, nil, err
	}

	newPath := strings.Trim(params.Path, "/")
	newPathHash := util.EncodeHash32(newPath)

	// 1. 判断目标目录是否存在
	existFolder, _ := s.folderRepo.GetByPathHash(ctx, newPathHash, vaultID, uid)
	if existFolder != nil && existFolder.Action != domain.FolderActionDelete {
		return nil, nil, code.ErrorFolderExist
	}

	oldPath := strings.Trim(params.OldPath, "/")
	oldPathHash := params.OldPathHash
	if oldPathHash == "" {
		oldPathHash = util.EncodeHash32(oldPath)
	}

	// 2. 获取旧文件夹
	f, err := s.folderRepo.GetByPathHash(ctx, oldPathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, code.ErrorFolderNotFound
		}
		return nil, nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 3. 标记旧文件夹删除
	f.Action = domain.FolderActionDelete
	f.UpdatedTimestamp = timex.Now().UnixMilli()
	oldFolder, err := s.folderRepo.Update(ctx, f, uid)
	if err != nil {
		return nil, nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	// 4. 新建或复用文件夹记录
	var newFolderCreated *domain.Folder
	if existFolder != nil {
		// 复用已删除的记录
		existFolder.Action = domain.FolderActionCreate
		existFolder.Path = newPath
		existFolder.PathHash = newPathHash
		existFolder.Level = f.Level
		existFolder.FID = f.FID
		existFolder.Mtime = timex.Now().UnixMilli()
		existFolder.Ctime = timex.Now().UnixMilli()
		existFolder.UpdatedTimestamp = timex.Now().UnixMilli()
		newFolderCreated, err = s.folderRepo.Update(ctx, existFolder, uid)
	} else {
		// 创建新记录
		newFolder := &domain.Folder{
			VaultID:          vaultID,
			Action:           domain.FolderActionCreate,
			Path:             newPath,
			PathHash:         newPathHash,
			Level:            f.Level,
			FID:              f.FID,
			Ctime:            timex.Now().UnixMilli(),
			Mtime:            timex.Now().UnixMilli(),
			UpdatedTimestamp: timex.Now().UnixMilli(),
		}
		newFolderCreated, err = s.folderRepo.Create(ctx, newFolder, uid)
	}

	if err != nil {
		return nil, nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	return s.domainToDTO(oldFolder), s.domainToDTO(newFolderCreated), nil
}

func (s *folderService) Get(ctx context.Context, uid int64, params *dto.FolderGetRequest) (*dto.FolderDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	pathHash := params.PathHash
	if pathHash == "" && params.Path != "" {
		pathHash = util.EncodeHash32(params.Path)
	}

	if pathHash == "" {
		return nil, code.ErrorInvalidParams.WithDetails("path or pathHash is required")
	}

	f, err := s.folderRepo.GetByPathHash(ctx, pathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorFolderNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	return s.domainToDTO(f), nil
}

func (s *folderService) ListNotes(ctx context.Context, uid int64, params *dto.FolderContentRequest, pager *app.Pager) ([]*dto.NoteNoContentDTO, int, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, 0, err
	}

	var fid int64 = 0
	if params.PathHash != "" {
		f, err := s.folderRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
		if err == nil {
			fid = f.ID
		}
	} else if params.Path != "" {
		pathHash := util.EncodeHash32(params.Path)
		f, err := s.folderRepo.GetByPathHash(ctx, pathHash, vaultID, uid)
		if err == nil {
			fid = f.ID
		}
	}

	notes, err := s.noteRepo.ListByFID(ctx, fid, vaultID, uid, pager.Page, pager.PageSize, params.SortBy, params.SortOrder)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}

	count, err := s.noteRepo.ListByFIDCount(ctx, fid, vaultID, uid)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var res []*dto.NoteNoContentDTO
	for _, n := range notes {
		res = append(res, &dto.NoteNoContentDTO{
			ID:               n.ID,
			Action:           string(n.Action),
			Path:             n.Path,
			PathHash:         n.PathHash,
			Version:          n.Version,
			Ctime:            n.Ctime,
			Mtime:            n.Mtime,
			UpdatedTimestamp: n.UpdatedTimestamp,
			UpdatedAt:        timex.Time(n.UpdatedAt),
			CreatedAt:        timex.Time(n.CreatedAt),
		})
	}
	return res, int(count), nil
}

func (s *folderService) ListFiles(ctx context.Context, uid int64, params *dto.FolderContentRequest, pager *app.Pager) ([]*dto.FileDTO, int, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, 0, err
	}

	var fid int64 = 0
	if params.PathHash != "" {
		f, err := s.folderRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
		if err == nil {
			fid = f.ID
		}
	} else if params.Path != "" {
		pathHash := util.EncodeHash32(params.Path)
		f, err := s.folderRepo.GetByPathHash(ctx, pathHash, vaultID, uid)
		if err == nil {
			fid = f.ID
		}
	}

	files, err := s.fileRepo.ListByFID(ctx, fid, vaultID, uid, pager.Page, pager.PageSize, params.SortBy, params.SortOrder)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}

	count, err := s.fileRepo.ListByFIDCount(ctx, fid, vaultID, uid)
	if err != nil {
		return nil, 0, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var res []*dto.FileDTO
	for _, f := range files {
		res = append(res, &dto.FileDTO{
			ID:               f.ID,
			Action:           string(f.Action),
			Path:             f.Path,
			PathHash:         f.PathHash,
			ContentHash:      f.ContentHash,
			SavePath:         f.SavePath,
			Size:             f.Size,
			Ctime:            f.Ctime,
			Mtime:            f.Mtime,
			UpdatedTimestamp: f.UpdatedTimestamp,
			UpdatedAt:        timex.Time(f.UpdatedAt),
			CreatedAt:        timex.Time(f.CreatedAt),
		})
	}
	return res, int(count), nil
}

// EnsurePathFID 确保资源的父目录存在并返回其 ID
func (s *folderService) EnsurePathFID(ctx context.Context, uid int64, vaultID int64, path string) (int64, error) {
	path = strings.Trim(path, "/")
	if path == "" || !strings.Contains(path, "/") {
		return 0, nil
	}

	lastSlash := strings.LastIndex(path, "/")
	parentPath := path[:lastSlash]

	// 补全目录逻辑
	parts := strings.Split(parentPath, "/")
	var currentFID int64 = 0

	for i := range parts {
		currentPath := strings.Join(parts[:i+1], "/")
		pathHash := util.EncodeHash32(currentPath)

		f, err := s.folderRepo.GetByPathHash(ctx, pathHash, vaultID, uid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newFolder := &domain.Folder{
					VaultID:  vaultID,
					Action:   domain.FolderActionCreate,
					Path:     currentPath,
					PathHash: pathHash,
					Level:    int64(i + 1),
					FID:      currentFID,
					Ctime:    timex.Now().UnixMilli(),
					Mtime:    timex.Now().UnixMilli(),
				}
				f, err = s.folderRepo.Create(ctx, newFolder, uid)
				if err != nil {
					return 0, err
				}
			} else {
				return 0, err
			}
		} else if f.Action == domain.FolderActionDelete {
			f.Action = domain.FolderActionCreate
			f.Ctime = timex.Now().UnixMilli()
			f.Mtime = timex.Now().UnixMilli()

			f, err = s.folderRepo.Update(ctx, f, uid)
			if err != nil {
				return 0, err
			}
		}
		currentFID = f.ID
	}
	return currentFID, nil
}

// SyncResourceFID 同步 Vault 下资源的 FID（支持全量或部分同步）
func (s *folderService) SyncResourceFID(ctx context.Context, uid int64, vaultID int64, noteIDs []int64, fileIDs []int64) error {
	key := fmt.Sprintf("sync_resource_fid_%d_%d", uid, vaultID)
	// 如果是部分同步，不使用 singleflight，或者使用不同的 key
	if len(noteIDs) > 0 || len(fileIDs) > 0 {
		key = fmt.Sprintf("sync_resource_fid_%d_%d_%v_%v", uid, vaultID, noteIDs, fileIDs)
	}

	_, err, _ := s.sf.Do(key, func() (any, error) {
		// 同步笔记
		var notes []*domain.Note
		var err error
		if len(noteIDs) > 0 {
			notes, err = s.noteRepo.ListByIDs(ctx, noteIDs, uid)
		} else if len(noteIDs) == 0 && len(fileIDs) == 0 {
			// 全量同步
			notes, err = s.noteRepo.ListByUpdatedTimestamp(ctx, 0, vaultID, uid)
		}

		if err == nil {
			for _, n := range notes {
				if n.Action == domain.NoteActionDelete {
					continue
				}
				fid, err := s.EnsurePathFID(ctx, uid, vaultID, n.Path)
				if err == nil && n.FID != fid {
					n.FID = fid
					_, _ = s.noteRepo.Update(ctx, n, uid)
				}
			}
		}

		// 同步文件
		var files []*domain.File
		if len(fileIDs) > 0 {
			files, err = s.fileRepo.ListByIDs(ctx, fileIDs, uid)
		} else if len(noteIDs) == 0 && len(fileIDs) == 0 {
			// 全量同步
			files, err = s.fileRepo.ListByUpdatedTimestamp(ctx, 0, vaultID, uid)
		}

		if err == nil {
			for _, f := range files {
				if f.Action == domain.FileActionDelete {
					continue
				}
				fid, err := s.EnsurePathFID(ctx, uid, vaultID, f.Path)
				if err == nil && f.FID != fid {
					f.FID = fid
					_, _ = s.fileRepo.Update(ctx, f, uid)
				}
			}
		}
		return nil, nil
	})
	return err
}
