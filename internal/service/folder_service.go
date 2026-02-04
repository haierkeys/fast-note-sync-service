package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"gorm.io/gorm"
)

// FolderService 文件夹业务服务接口
type FolderService interface {
	Get(ctx context.Context, uid int64, vault string, pathHash string) (*dto.FolderDTO, error)
	List(ctx context.Context, uid int64, params *dto.FolderListRequest) ([]*dto.FolderDTO, error)
	ListByUpdatedTimestamp(ctx context.Context, uid int64, vault string, lastTime int64) ([]*dto.FolderDTO, error)
	UpdateOrCreate(ctx context.Context, uid int64, params *dto.FolderCreateRequest) (*dto.FolderDTO, error)
	Delete(ctx context.Context, uid int64, params *dto.FolderDeleteRequest) error
	Rename(ctx context.Context, uid int64, params *dto.FolderRenameRequest) error
	ListNotes(ctx context.Context, uid int64, params *dto.FolderContentRequest) ([]*dto.NoteDTO, error)
	ListFiles(ctx context.Context, uid int64, params *dto.FolderContentRequest) ([]*dto.FileDTO, error)
}

type folderService struct {
	folderRepo   domain.FolderRepository
	noteRepo     domain.NoteRepository
	fileRepo     domain.FileRepository
	vaultService VaultService
}

func NewFolderService(folderRepo domain.FolderRepository, noteRepo domain.NoteRepository, fileRepo domain.FileRepository, vaultSvc VaultService) FolderService {
	return &folderService{
		folderRepo:   folderRepo,
		noteRepo:     noteRepo,
		fileRepo:     fileRepo,
		vaultService: vaultSvc,
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
					VaultID:          vaultID,
					Action:           domain.FolderActionCreate,
					Path:             currentPath,
					PathHash:         pathHash,
					Level:            int64(i + 1),
					FID:              currentFID,
					UpdatedTimestamp: time.Now().UnixMilli(),
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
			f.UpdatedTimestamp = time.Now().UnixMilli()
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

func (s *folderService) Rename(ctx context.Context, uid int64, params *dto.FolderRenameRequest) error {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return err
	}

	oldPathHash := util.EncodeHash32(params.OldPath)
	f, err := s.folderRepo.GetByPathHash(ctx, oldPathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.ErrorFolderNotFound
		}
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	// Update folder itself
	f.Path = params.Path
	f.PathHash = util.EncodeHash32(params.Path)
	f.UpdatedTimestamp = timex.Now().UnixMilli()
	f.Action = domain.FolderActionModify

	_, err = s.folderRepo.Update(ctx, f, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}

	return nil
}

func (s *folderService) Get(ctx context.Context, uid int64, vault string, pathHash string) (*dto.FolderDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, vault)
	if err != nil {
		return nil, err
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

func (s *folderService) ListNotes(ctx context.Context, uid int64, params *dto.FolderContentRequest) ([]*dto.NoteDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
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

	notes, err := s.noteRepo.ListByFID(ctx, fid, vaultID, uid, params.Page, params.PageSize, params.SortBy, params.SortOrder)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	var res []*dto.NoteDTO
	for _, n := range notes {
		res = append(res, &dto.NoteDTO{
			ID:               n.ID,
			Action:           string(n.Action),
			Path:             n.Path,
			PathHash:         n.PathHash,
			Content:          n.Content,
			ContentHash:      n.ContentHash,
			Version:          n.Version,
			Ctime:            n.Ctime,
			Mtime:            n.Mtime,
			UpdatedTimestamp: n.UpdatedTimestamp,
			UpdatedAt:        timex.Time(n.UpdatedAt),
			CreatedAt:        timex.Time(n.CreatedAt),
		})
	}
	return res, nil
}

func (s *folderService) ListFiles(ctx context.Context, uid int64, params *dto.FolderContentRequest) ([]*dto.FileDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
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

	files, err := s.fileRepo.ListByFID(ctx, fid, vaultID, uid, params.Page, params.PageSize, params.SortBy, params.SortOrder)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
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
	return res, nil
}
