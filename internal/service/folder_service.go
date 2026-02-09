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
	Delete(ctx context.Context, uid int64, params *dto.FolderDeleteRequest) (*dto.FolderDTO, error)
	Rename(ctx context.Context, uid int64, params *dto.FolderRenameRequest) (*dto.FolderDTO, *dto.FolderDTO, error)
	ListNotes(ctx context.Context, uid int64, params *dto.FolderContentRequest, pager *app.Pager) ([]*dto.NoteNoContentDTO, int, error)
	ListFiles(ctx context.Context, uid int64, params *dto.FolderContentRequest, pager *app.Pager) ([]*dto.FileDTO, int, error)
	EnsurePathFID(ctx context.Context, uid int64, vaultID int64, path string) (int64, error)
	SyncResourceFID(ctx context.Context, uid int64, vaultID int64, noteIDs []int64, fileIDs []int64) error
	GetTree(ctx context.Context, uid int64, params *dto.FolderTreeRequest) (*dto.FolderTreeResponse, error)
	CleanDuplicateFolders(ctx context.Context, uid int64, vaultID int64) error
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

	if params.Path != strings.Trim(params.Path, "/") && params.Path != "" {
		return nil, code.ErrorInvalidParams.WithDetails("path cannot be empty")
	}

	var fid int64 = 0
	if params.Path != "" {
		if params.PathHash == "" {
			params.PathHash = util.EncodeHash32(params.Path)
		}
		f, err := s.folderRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
		if err == nil {
			fid = f.ID
		}
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

	if params.Path != strings.Trim(params.Path, "/") && params.Path != "" {
		return nil, code.ErrorInvalidParams.WithDetails("path cannot be empty")
	}

	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	// 统一调用 EnsurePathFID
	fid, err := s.EnsurePathFID(ctx, uid, vaultID, params.Path)
	if err != nil {
		return nil, code.ErrorFolderModifyOrCreateFailed.WithDetails(err.Error())
	}

	f, err := s.folderRepo.GetByID(ctx, fid, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	return s.domainToDTO(f), nil
}

func (s *folderService) Delete(ctx context.Context, uid int64, params *dto.FolderDeleteRequest) (*dto.FolderDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	if params.Path != strings.Trim(params.Path, "/") && params.Path != "" {
		return nil, code.ErrorInvalidParams.WithDetails("path cannot be empty")
	}

	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	f, err := s.folderRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorFolderNotFound
		}
		return nil, code.ErrorFolderGetFailed.WithDetails(err.Error())
	}

	// 软删除
	f.Action = domain.FolderActionDelete
	f.UpdatedTimestamp = time.Now().UnixMilli()
	_, err = s.folderRepo.Update(ctx, f, uid)
	if err != nil {
		return nil, code.ErrorFolderDeleteFailed.WithDetails(err.Error())
	}

	return s.domainToDTO(f), nil
}

func (s *folderService) ListByUpdatedTimestamp(ctx context.Context, uid int64, vault string, lastTime int64) ([]*dto.FolderDTO, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, vault)
	if err != nil {
		return nil, err
	}

	folders, err := s.folderRepo.ListByUpdatedTimestamp(ctx, lastTime, vaultID, uid)
	if err != nil {
		return nil, code.ErrorFolderListFailed.WithDetails(err.Error())
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

	if params.Path != strings.Trim(params.Path, "/") && params.Path != "" {
		return nil, nil, code.ErrorInvalidParams.WithDetails("path cannot be empty")
	}

	if params.OldPath != strings.Trim(params.OldPath, "/") && params.OldPath != "" {
		return nil, nil, code.ErrorInvalidParams.WithDetails("oldPath cannot be empty")
	}

	if params.PathHash == "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	if params.OldPathHash == "" {
		params.OldPathHash = util.EncodeHash32(params.OldPath)
	}

	// 1. 获取旧文件夹
	oldFolder, err := s.folderRepo.GetByPathHash(ctx, params.OldPathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newFolder, err := s.UpdateOrCreate(ctx, uid, &dto.FolderCreateRequest{
				Path:     params.Path,
				PathHash: params.PathHash,
				Vault:    params.Vault,
			})
			if err != nil {
				return nil, nil, err
			}
			return nil, newFolder, nil
		}
		return nil, nil, code.ErrorFolderGetFailed.WithDetails(params.OldPathHash + "->" + params.PathHash + ":" + err.Error())
	}

	// 2. 判断目标目录是否存在且有效
	existFolder, _ := s.folderRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if existFolder != nil && existFolder.Action != domain.FolderActionDelete {
		newFolder, err := s.Delete(ctx, uid, &dto.FolderDeleteRequest{
			Path:     params.OldPath,
			PathHash: params.OldPathHash,
			Vault:    params.Vault,
		})
		if err != nil {
			return nil, nil, err
		}
		return s.domainToDTO(oldFolder), newFolder, nil
	}

	// 3. 标记旧文件夹删除
	oldFolder.Action = domain.FolderActionDelete
	oldFolder.UpdatedTimestamp = timex.Now().UnixMilli()
	oldFolder, err = s.folderRepo.Update(ctx, oldFolder, uid)
	if err != nil {
		return nil, nil, code.ErrorFolderRenameFailed.WithDetails(err.Error())
	}

	// 4. 新建或复用文件夹记录（统一调用 EnsurePathFID）
	fid, err := s.EnsurePathFID(ctx, uid, vaultID, params.Path)
	if err != nil {
		return nil, nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	newFolderCreated, err := s.folderRepo.GetByID(ctx, fid, uid)
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

	if params.Path != strings.Trim(params.Path, "/") && params.Path != "" {
		return nil, code.ErrorInvalidParams.WithDetails("path cannot be empty")
	}

	if params.Path != "" {
		params.PathHash = util.EncodeHash32(params.Path)
	}

	if params.PathHash == "" {
		return nil, code.ErrorInvalidParams.WithDetails("path or pathHash is required")
	}

	f, err := s.folderRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorFolderNotFound
		}
		return nil, code.ErrorFolderGetFailed.WithDetails(err.Error())
	}

	return s.domainToDTO(f), nil
}

func (s *folderService) ListNotes(ctx context.Context, uid int64, params *dto.FolderContentRequest, pager *app.Pager) ([]*dto.NoteNoContentDTO, int, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, 0, err
	}

	if params.Path != strings.Trim(params.Path, "/") && params.Path != "" {
		return nil, 0, code.ErrorInvalidParams.WithDetails("path cannot be empty")
	}

	var fid int64 = 0
	if params.Path != "" {
		if params.PathHash == "" {
			params.PathHash = util.EncodeHash32(params.Path)
		}
		f, err := s.folderRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
		if err == nil {
			fid = f.ID
		}
	}

	notes, err := s.noteRepo.ListByFID(ctx, fid, vaultID, uid, pager.Page, pager.PageSize, params.SortBy, params.SortOrder)
	if err != nil {
		return nil, 0, code.ErrorNoteListFailed.WithDetails(err.Error())
	}

	count, err := s.noteRepo.ListByFIDCount(ctx, fid, vaultID, uid)
	if err != nil {
		return nil, 0, code.ErrorNoteListFailed.WithDetails(err.Error())
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

	if params.Path != strings.Trim(params.Path, "/") && params.Path != "" {
		return nil, 0, code.ErrorInvalidParams.WithDetails("path cannot be empty")
	}

	var fid int64 = 0
	if params.Path != "" {
		if params.PathHash == "" {
			params.PathHash = util.EncodeHash32(params.Path)
		}
		f, err := s.folderRepo.GetByPathHash(ctx, params.PathHash, vaultID, uid)
		if err == nil {
			fid = f.ID
		}
	}

	files, err := s.fileRepo.ListByFID(ctx, fid, vaultID, uid, pager.Page, pager.PageSize, params.SortBy, params.SortOrder)
	if err != nil {
		return nil, 0, code.ErrorFileListFailed.WithDetails(err.Error())
	}

	count, err := s.fileRepo.ListByFIDCount(ctx, fid, vaultID, uid)
	if err != nil {
		return nil, 0, code.ErrorFileListFailed.WithDetails(err.Error())
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
//
// NOTE: Race condition — when multiple notes sync concurrently (even from a
// single device), each goroutine calls EnsurePathFID independently. The
// check-then-create below (GetByPathHash → Create) is not atomic, so two
// goroutines can both see "not found" and both insert a folder record for the
// same path. This produces duplicate rows in the folder table (confirmed via
// direct DB inspection — e.g. 3 rows for "projects" with IDs 4,5,6).
//
// The query-side methods (List, ListNotes, ListFiles, GetTree) handle this by
// resolving all folder IDs per path (GetAllByPathHash + FID IN queries).
//
// A proper fix would be to either:
//   - Use singleflight keyed by (vaultID, path) to coalesce concurrent creates
//   - Add a UNIQUE constraint on (vault_id, path_hash) and handle conflict
func (s *folderService) EnsurePathFID(ctx context.Context, uid int64, vaultID int64, path string) (int64, error) {
	path = strings.Trim(path, "/")
	if path == "" {
		return 0, nil
	}

	// 分解目录
	parts := strings.Split(path, "/")
	var currentFID int64 = 0

	for i := range parts {
		currentPath := strings.Join(parts[:i+1], "/")
		pathHash := util.EncodeHash32(currentPath)

		key := fmt.Sprintf("ensure_folder_%d_%d_%s", uid, vaultID, pathHash)
		val, err, _ := s.sf.Do(key, func() (any, error) {
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
					return f.ID, nil
				}
				return nil, err
			} else if f.Action == domain.FolderActionDelete {
				f.Action = domain.FolderActionCreate
				f.Ctime = timex.Now().UnixMilli()
				f.Mtime = timex.Now().UnixMilli()
				f, err = s.folderRepo.Update(ctx, f, uid)
				if err != nil {
					return nil, err
				}
				return f.ID, nil
			}
			return f.ID, nil
		})

		if err != nil {
			return 0, err
		}

		currentFID = val.(int64)
	}
	return currentFID, nil
}

// SyncResourceFID 同步 Vault 下资源的 FID（支持全量或部分同步）
func (s *folderService) SyncResourceFID(ctx context.Context, uid int64, vaultID int64, noteIDs []int64, fileIDs []int64) error {
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
			path := strings.Trim(n.Path, "/")
			if !strings.Contains(path, "/") {
				if n.FID != 0 {
					n.FID = 0
					_, _ = s.noteRepo.Update(ctx, n, uid)
				}
				continue
			}
			parentPath := path[:strings.LastIndex(path, "/")]
			fid, err := s.EnsurePathFID(ctx, uid, vaultID, parentPath)
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
			path := strings.Trim(f.Path, "/")
			if !strings.Contains(path, "/") {
				if f.FID != 0 {
					f.FID = 0
					_, _ = s.fileRepo.Update(ctx, f, uid)
				}
				continue
			}
			parentPath := f.Path[:strings.LastIndex(f.Path, "/")]
			fid, err := s.EnsurePathFID(ctx, uid, vaultID, parentPath)
			if err == nil && f.FID != fid {
				f.FID = fid
				_, _ = s.fileRepo.Update(ctx, f, uid)
			}
		}
	}
	return nil
}

// GetTree returns the complete folder tree structure for a vault
func (s *folderService) GetTree(ctx context.Context, uid int64, params *dto.FolderTreeRequest) (*dto.FolderTreeResponse, error) {
	vaultID, err := s.vaultService.MustGetID(ctx, uid, params.Vault)
	if err != nil {
		return nil, err
	}

	// Get all non-deleted folders
	folders, err := s.folderRepo.ListByUpdatedTimestamp(ctx, 0, vaultID, uid)
	if err != nil {
		return nil, code.ErrorFolderListFailed.WithDetails(err.Error())
	}

	// Deduplicate by path, collect all IDs per path for counting
	type folderInfo struct {
		path   string
		ids    []int64 // all DB IDs for this path (for counting notes/files)
		parent string  // parent path ("" for root)
	}
	infoByPath := make(map[string]*folderInfo)

	for _, f := range folders {
		if f.Action == domain.FolderActionDelete {
			continue
		}
		info, exists := infoByPath[f.Path]
		if !exists {
			parent := ""
			if idx := strings.LastIndex(f.Path, "/"); idx >= 0 {
				parent = f.Path[:idx]
			}
			info = &folderInfo{path: f.Path, parent: parent}
			infoByPath[f.Path] = info
		}
		info.ids = append(info.ids, f.ID)
	}

	// Count notes and files per folder (sum across all duplicate IDs)
	noteCountByPath := make(map[string]int)
	fileCountByPath := make(map[string]int)
	for path, info := range infoByPath {
		for _, id := range info.ids {
			nc, err := s.noteRepo.ListByFIDCount(ctx, id, vaultID, uid)
			if err == nil {
				noteCountByPath[path] += int(nc)
			}
			fc, err := s.fileRepo.ListByFIDCount(ctx, id, vaultID, uid)
			if err == nil {
				fileCountByPath[path] += int(fc)
			}
		}
	}

	// Root counts (FID = 0)
	rootNoteCount := 0
	rootFileCount := 0
	count, err := s.noteRepo.ListByFIDCount(ctx, 0, vaultID, uid)
	if err == nil {
		rootNoteCount = int(count)
	}
	count, err = s.fileRepo.ListByFIDCount(ctx, 0, vaultID, uid)
	if err == nil {
		rootFileCount = int(count)
	}

	// Build parent→children map by path
	childrenByParent := make(map[string][]string)
	for path, info := range infoByPath {
		childrenByParent[info.parent] = append(childrenByParent[info.parent], path)
	}

	// Build tree recursively
	var buildNode func(path string, currentDepth int) *dto.FolderTreeNode
	buildNode = func(path string, currentDepth int) *dto.FolderTreeNode {
		name := path
		if idx := strings.LastIndex(path, "/"); idx >= 0 {
			name = path[idx+1:]
		}

		node := &dto.FolderTreeNode{
			Path:      path,
			Name:      name,
			NoteCount: noteCountByPath[path],
			FileCount: fileCountByPath[path],
		}

		if params.Depth > 0 && currentDepth >= params.Depth {
			return node
		}

		for _, childPath := range childrenByParent[path] {
			node.Children = append(node.Children, buildNode(childPath, currentDepth+1))
		}

		return node
	}

	// Build root level folders (parent = "")
	var rootFolders []*dto.FolderTreeNode
	for _, path := range childrenByParent[""] {
		rootFolders = append(rootFolders, buildNode(path, 1))
	}

	return &dto.FolderTreeResponse{
		Folders:       rootFolders,
		RootNoteCount: rootNoteCount,
		RootFileCount: rootFileCount,
	}, nil
}

func (s *folderService) CleanDuplicateFolders(ctx context.Context, uid int64, vaultID int64) error {
	// 1. 获取所有文件夹记录（包含已删除的，以便按逻辑清理）
	folders, err := s.folderRepo.ListByUpdatedTimestamp(ctx, 0, vaultID, uid)
	if err != nil {
		return err
	}

	// 2. 按 PathHash 分组
	grouped := make(map[string][]*domain.Folder)
	for _, f := range folders {
		grouped[f.PathHash] = append(grouped[f.PathHash], f)
	}

	// 3. 遍历分组，识别重复项
	for pathHash, list := range grouped {
		if len(list) <= 1 {
			continue
		}

		// 检查这组重复记录中是否有被标记为删除的
		var hasDeleted bool
		for _, f := range list {
			if f.Action == domain.FolderActionDelete {
				hasDeleted = true
				break
			}
		}

		if hasDeleted {
			// 如果存在已删除记录，则删除所有未标记删除的记录（解决已删除但仍被 EnsurePathFID 误创出的活跃记录）
			for _, f := range list {
				if f.Action != domain.FolderActionDelete {
					s.sf.Forget(fmt.Sprintf("ensure_folder_%d_%d_%s", uid, vaultID, pathHash))
					_ = s.folderRepo.Delete(ctx, f.ID, uid)
				}
			}
		} else {
			// 如果全部都是活跃记录，保留 ID 最大的一条（假设是最后创建的）
			var maxID int64
			for _, f := range list {
				if f.ID > maxID {
					maxID = f.ID
				}
			}

			for _, f := range list {
				if f.ID != maxID {
					s.sf.Forget(fmt.Sprintf("ensure_folder_%d_%d_%s", uid, vaultID, pathHash))
					_ = s.folderRepo.Delete(ctx, f.ID, uid)
				}
			}
		}
	}

	return nil
}
