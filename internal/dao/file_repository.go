// Package dao implements the data access layer
// Package dao 实现数据访问层
package dao

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

// fileRepository implements domain.FileRepository interface
// fileRepository 实现 domain.FileRepository 接口
type fileRepository struct {
	dao             *Dao
	customPrefixKey string
}

// NewFileRepository creates FileRepository instance
// NewFileRepository 创建 FileRepository 实例
func NewFileRepository(dao *Dao) domain.FileRepository {
	return &fileRepository{dao: dao, customPrefixKey: "user_file_"}
}

func (r *fileRepository) GetKey(uid int64) string {
	return r.customPrefixKey + strconv.FormatInt(uid, 10)
}

func init() {
	RegisterModel(ModelConfig{
		Name: "File",
		RepoFactory: func(d *Dao) daoDBCustomKey {
			return NewFileRepository(d).(daoDBCustomKey)
		},
	})
}

// file gets the file query object
// file 获取文件查询对象
func (r *fileRepository) file(uid int64) *query.Query {
	return r.dao.QueryWithOnceInit(func(g *gorm.DB) {
		model.AutoMigrate(g, "File")
	}, r.GetKey(uid)+"#file", r.GetKey(uid))
}

// toDomain converts database model to domain model
// toDomain 将数据库模型转换为领域模型
func (r *fileRepository) toDomain(m *model.File, uid int64) *domain.File {
	if m == nil {
		return nil
	}
	file := &domain.File{
		ID:               m.ID,
		VaultID:          m.VaultID,
		Action:           domain.FileAction(m.Action),
		FID:              m.FID,
		Path:             m.Path,
		PathHash:         m.PathHash,
		ContentHash:      m.ContentHash,
		SavePath:         m.SavePath,
		Rename:           m.Rename,
		Size:             m.Size,
		Ctime:            m.Ctime,
		Mtime:            m.Mtime,
		UpdatedTimestamp: m.UpdatedTimestamp,
		CreatedAt:        time.Time(m.CreatedAt),
		UpdatedAt:        time.Time(m.UpdatedAt),
	}
	r.fillFilePath(uid, file)
	return file
}

// toModel converts domain model to database model
// toModel 将领域模型转换为数据库模型
func (r *fileRepository) toModel(file *domain.File) *model.File {
	if file == nil {
		return nil
	}
	return &model.File{
		ID:               file.ID,
		VaultID:          file.VaultID,
		Action:           string(file.Action),
		FID:              file.FID,
		Path:             file.Path,
		PathHash:         file.PathHash,
		ContentHash:      file.ContentHash,
		SavePath:         file.SavePath,
		Rename:           file.Rename,
		Size:             file.Size,
		Ctime:            file.Ctime,
		Mtime:            file.Mtime,
		UpdatedTimestamp: file.UpdatedTimestamp,
		CreatedAt:        timex.Time(file.CreatedAt),
		UpdatedAt:        timex.Time(file.UpdatedAt),
	}
}

// fillFilePath fills file SavePath and handles old file migration
// fillFilePath 填充文件的保存路径并处理旧文件迁移
func (r *fileRepository) fillFilePath(uid int64, f *domain.File) {
	if f == nil {
		return
	}
	folderPath := r.dao.GetFileFolderPath(uid, f.ID)
	standardPath := filepath.Join(folderPath, "file.dat")

	// Record original SavePath for migration check
	// 记录原始 SavePath 以便进行迁移检查
	oldSavePath := f.SavePath

	// Update to standard path
	// 更新为标准路径
	f.SavePath = standardPath

	// Migrate only if standard path doesn't exist, old path is provided, and old file exists on disk
	// 仅在标准路径不存在，且明确给出了旧路径，且旧文件确实存在磁盘上时才执行迁移
	if _, err := os.Stat(standardPath); os.IsNotExist(err) && oldSavePath != "" && oldSavePath != standardPath {
		if _, errOld := os.Stat(oldSavePath); errOld == nil {
			// 只有在确定要移动文件时才创建目录
			_ = os.MkdirAll(folderPath, 0755)
			_ = os.Rename(oldSavePath, standardPath)
		}
	}
}

// GetByID retrieves file by ID
// GetByID 根据 ID 获取文件
func (r *fileRepository) GetByID(ctx context.Context, id, uid int64) (*domain.File, error) {
	u := r.file(uid).File
	m, err := u.WithContext(ctx).Where(u.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m, uid), nil
}

// GetByPathHash retrieves file by path hash
// GetByPathHash 根据路径哈希获取文件
func (r *fileRepository) GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*domain.File, error) {
	u := r.file(uid).File
	m, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.PathHash.Eq(pathHash),
	).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m, uid), nil
}

// ListByPathHash retrieves file list by path hash (handling duplicate records)
// ListByPathHash 根据路径哈希获取文件列表（处理重复记录）
func (r *fileRepository) ListByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) ([]*domain.File, error) {
	u := r.file(uid).File
	mList, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.PathHash.Eq(pathHash),
	).Find()
	if err != nil {
		return nil, err
	}
	var list []*domain.File
	for _, m := range mList {
		list = append(list, r.toDomain(m, uid))
	}
	return list, nil
}

// GetByPath retrieves file by path
// GetByPath 根据路径获取文件
func (r *fileRepository) GetByPath(ctx context.Context, path string, vaultID, uid int64) (*domain.File, error) {
	u := r.file(uid).File
	m, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Path.Eq(path),
	).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m, uid), nil
}

// GetByPathLike retrieves file by path suffix
// GetByPathLike 根据路径后缀获取文件
func (r *fileRepository) GetByPathLike(ctx context.Context, path string, vaultID, uid int64) (*domain.File, error) {
	u := r.file(uid).File
	m, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Path.Like("%"+path),
		u.Action.Neq("delete"),
	).First()
	if err != nil {
		return nil, err
	}
	return r.toDomain(m, uid), nil
}

// Create creates a file
// Create 创建文件
func (r *fileRepository) Create(ctx context.Context, file *domain.File, uid int64) (*domain.File, error) {
	var result *domain.File
	var createErr error

	err := r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.file(uid).File
		m := r.toModel(file)

		m.UpdatedTimestamp = timex.Now().UnixMilli()
		m.CreatedAt = timex.Now()
		m.UpdatedAt = timex.Now()

		tempSavePath := m.SavePath
		m.SavePath = "" // 不在数据库中保存路径

		createErr = u.WithContext(ctx).Create(m)
		if createErr != nil {
			return createErr
		}

		// Move file to Vault directory, fixed naming as file.dat
		// 移动文件到 Vault 目录，固定命名为 file.dat
		if tempSavePath != "" {
			folderPath := r.dao.GetFileFolderPath(uid, m.ID)
			_ = os.MkdirAll(folderPath, 0755)
			finalPath := filepath.Join(folderPath, "file.dat")

			if err := os.Rename(tempSavePath, finalPath); err != nil {
				_ = os.Rename(tempSavePath, finalPath)
			}
		}

		result = r.toDomain(m, uid)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, createErr
}

// Update updates a file
// Update 更新文件
func (r *fileRepository) Update(ctx context.Context, file *domain.File, uid int64) (*domain.File, error) {
	var result *domain.File
	var updateErr error

	err := r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.file(uid).File
		m := r.toModel(file)

		m.UpdatedTimestamp = timex.Now().UnixMilli()
		m.UpdatedAt = timex.Now()

		tempSavePath := m.SavePath
		m.SavePath = "" // 不在数据库中更新路径

		// If a new temporary path is provided, move it to the fixed file.dat
		// 如果提供了新的临时路径，则移动到固定的 file.dat
		if tempSavePath != "" {
			folderPath := r.dao.GetFileFolderPath(uid, m.ID)
			_ = os.MkdirAll(folderPath, 0755)
			finalPath := filepath.Join(folderPath, "file.dat")
			_ = os.Rename(tempSavePath, finalPath)
		}

		updateErr = u.WithContext(ctx).Where(
			u.ID.Eq(m.ID),
		).Save(m)

		if updateErr != nil {
			return updateErr
		}
		result = r.toDomain(m, uid)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, updateErr
}

// UpdateMtime updates file modification time
// UpdateMtime 更新文件修改时间
func (r *fileRepository) UpdateMtime(ctx context.Context, mtime int64, id, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.file(uid).File

		_, err := u.WithContext(ctx).Where(
			u.ID.Eq(id),
		).UpdateSimple(
			u.Mtime.Value(mtime),
			u.UpdatedTimestamp.Value(timex.Now().UnixMilli()),
			u.UpdatedAt.Value(timex.Now()),
		)
		return err
	})
}

// UpdateActionMtime updates file action and modification time
// UpdateActionMtime 更新文件类型并修改时间
func (r *fileRepository) UpdateActionMtime(ctx context.Context, action domain.FileAction, mtime int64, id, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.file(uid).File

		_, err := u.WithContext(ctx).Where(
			u.ID.Eq(id),
		).UpdateSimple(
			u.Action.Value(string(action)),
			u.Mtime.Value(mtime),
			u.UpdatedTimestamp.Value(timex.Now().UnixMilli()),
			u.UpdatedAt.Value(timex.Now()),
		)
		return err
	})
}

// Delete physically deletes a file
// Delete 物理删除文件
func (r *fileRepository) Delete(ctx context.Context, id, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.file(uid).File
		_, err := u.WithContext(ctx).Where(u.ID.Eq(id)).Delete()
		if err != nil {
			return err
		}

		// Delete physical file
		// 删除物理文件
		folderPath := r.dao.GetFileFolderPath(uid, id)
		_ = r.dao.RemoveContentFolder(folderPath)

		return nil
	})
}

// DeletePhysicalByTime physically deletes files marked as deleted by time
// DeletePhysicalByTime 根据时间物理删除已标记删除的文件
func (r *fileRepository) DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.file(uid).File

		// Find records to be deleted to remove folders in the file system
		// 查找待删除的记录，以便删除文件系统中的文件夹
		mList, err := u.WithContext(ctx).Where(
			u.Action.Eq("delete"),
			u.UpdatedTimestamp.Lt(timestamp),
		).Find()

		if err == nil {
			for _, m := range mList {
				folderPath := r.dao.GetFileFolderPath(uid, m.ID)
				_ = r.dao.RemoveContentFolder(folderPath)
			}
		}

		_, err = u.WithContext(ctx).Where(
			u.Action.Eq("delete"),
			u.UpdatedTimestamp.Lt(timestamp),
		).Delete()
		return err
	})
}

// DeletePhysicalByTimeAll physically deletes files marked as deleted for all users by time
// DeletePhysicalByTimeAll 根据时间物理删除所有用户的已标记删除的文件
func (r *fileRepository) DeletePhysicalByTimeAll(ctx context.Context, timestamp int64) error {
	// Get all user UIDs
	// 获取所有用户 UID
	uids, err := r.dao.GetAllUserUIDs()
	if err != nil {
		return err
	}

	// Execute cleanup user by user
	// 逐用户执行清理
	for _, uid := range uids {
		if err := r.DeletePhysicalByTime(ctx, timestamp, uid); err != nil {
			// 记录错误但继续处理其他用户
			continue
		}
	}
	return nil
}

// List retrieves file list by page
// List 分页获取文件列表
func (r *fileRepository) List(ctx context.Context, vaultID int64, page, pageSize int, uid int64, keyword string, isRecycle bool, sortBy string, sortOrder string) ([]*domain.File, error) {
	u := r.file(uid).File
	q := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
	)

	if isRecycle {
		q = q.Where(u.Action.Eq(string(domain.FileActionDelete)), u.Rename.Eq(0))
	} else {
		q = q.Where(u.Action.Neq(string(domain.FileActionDelete)))
	}

	if keyword != "" {
		q = q.Where(u.Path.Like("%" + keyword + "%"))
	}

	// Sorting
	// 排序
	var sortField field.OrderExpr
	switch sortBy {
	case "ctime":
		sortField = u.Ctime
	case "path":
		sortField = u.Path
	case "mtime":
		fallthrough
	default:
		sortField = u.Mtime
	}

	var orderExpr field.Expr
	if strings.ToLower(sortOrder) == "asc" {
		orderExpr = sortField
	} else {
		orderExpr = sortField.Desc()
	}

	orderExprs := []field.Expr{orderExpr}
	if sortBy != "path" {
		orderExprs = append(orderExprs, u.Path)
	}

	modelList, err := q.Order(orderExprs...).
		Limit(pageSize).
		Offset(app.GetPageOffset(page, pageSize)).
		Find()

	if err != nil {
		return nil, err
	}

	var list []*domain.File
	for _, m := range modelList {
		list = append(list, r.toDomain(m, uid))
	}
	return list, nil
}

// ListCount retrieves file count
// ListCount 获取文件数量
func (r *fileRepository) ListCount(ctx context.Context, vaultID, uid int64, keyword string, isRecycle bool) (int64, error) {
	u := r.file(uid).File
	q := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
	)

	if isRecycle {
		q = q.Where(u.Action.Eq(string(domain.FileActionDelete)), u.Rename.Eq(0))
	} else {
		q = q.Where(u.Action.Neq(string(domain.FileActionDelete)))
	}

	if keyword != "" {
		q = q.Where(u.Path.Like("%" + keyword + "%"))
	}

	count, err := q.Count()
	if err != nil {
		return 0, err
	}

	return count, nil
}

// ListByUpdatedTimestamp retrieves file list by updated timestamp
// ListByUpdatedTimestamp 根据更新时间戳获取文件列表
func (r *fileRepository) ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.File, error) {
	return r.ListByUpdatedTimestampPage(ctx, timestamp, vaultID, uid, 0, 0)
}

// ListByUpdatedTimestampPage retrieves file list by updated timestamp by page
// ListByUpdatedTimestampPage 根据更新时间戳分页获取文件列表
func (r *fileRepository) ListByUpdatedTimestampPage(ctx context.Context, timestamp, vaultID, uid int64, offset, limit int) ([]*domain.File, error) {
	u := r.file(uid).File
	query := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.UpdatedTimestamp.Gt(timestamp),
	).Order(u.UpdatedTimestamp.Desc())

	var mList []*model.File
	var err error
	if limit > 0 {
		mList, _, err = query.FindByPage(offset, limit)
	} else {
		mList, err = query.Find()
	}

	if err != nil {
		return nil, err
	}

	var list []*domain.File
	for _, m := range mList {
		list = append(list, r.toDomain(m, uid))
	}
	return list, nil
}

// ListByMtime retrieves file list by modification timestamp
// ListByMtime 根据修改时间戳获取文件列表
func (r *fileRepository) ListByMtime(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.File, error) {
	u := r.file(uid).File
	mList, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Mtime.Gt(timestamp),
	).Order(u.UpdatedTimestamp.Desc()).
		Find()

	if err != nil {
		return nil, err
	}

	var list []*domain.File
	for _, m := range mList {
		list = append(list, r.toDomain(m, uid))
	}
	return list, nil
}

// CountSizeSum retrieves total file count and size sum
// CountSizeSum 获取文件数量和大小总和
func (r *fileRepository) CountSizeSum(ctx context.Context, vaultID, uid int64) (*domain.CountSizeResult, error) {
	u := r.file(uid).File

	result := &struct {
		Size  int64
		Count int64
	}{}

	err := u.WithContext(ctx).Select(u.Size.Sum().As("size"), u.Size.Count().As("count")).Where(
		u.VaultID.Eq(vaultID),
		u.Action.Neq("delete"),
		u.Rename.Eq(0),
	).Scan(result)

	if err != nil {
		return nil, err
	}

	return &domain.CountSizeResult{
		Count: result.Count,
		Size:  result.Size,
	}, nil
}

// ListByFID retrieves file list by folder ID
// ListByFID 根据文件夹ID获取文件列表
func (r *fileRepository) ListByFID(ctx context.Context, fid, vaultID, uid int64, page, pageSize int, sortBy, sortOrder string) ([]*domain.File, error) {
	u := r.file(uid).File
	q := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.FID.Eq(fid),
		u.Action.Neq(string(domain.FileActionDelete)),
	)

	// Build order clause
	// 构建排序语句
	orderClause := buildFileOrderClause(sortBy, sortOrder)

	var modelList []*model.File
	err := q.UnderlyingDB().
		Order(orderClause).
		Limit(pageSize).
		Offset(app.GetPageOffset(page, pageSize)).
		Find(&modelList).Error

	if err != nil {
		return nil, err
	}

	var list []*domain.File
	for _, m := range modelList {
		list = append(list, r.toDomain(m, uid))
	}
	return list, nil
}

// ListByFIDCount retrieves file count by folder ID
// ListByFIDCount 根据文件夹ID获取文件数量
func (r *fileRepository) ListByFIDCount(ctx context.Context, fid, vaultID, uid int64) (int64, error) {
	u := r.file(uid).File
	q := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.FID.Eq(fid),
		u.Action.Neq(string(domain.FileActionDelete)),
	)

	return q.Count()
}

func (r *fileRepository) ListByFIDs(ctx context.Context, fids []int64, vaultID, uid int64, page, pageSize int, sortBy, sortOrder string) ([]*domain.File, error) {
	u := r.file(uid).File
	q := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.FID.In(fids...),
		u.Action.Neq(string(domain.FileActionDelete)),
	)

	orderClause := buildFileOrderClause(sortBy, sortOrder)

	var modelList []*model.File
	err := q.UnderlyingDB().
		Order(orderClause).
		Limit(pageSize).
		Offset(app.GetPageOffset(page, pageSize)).
		Find(&modelList).Error

	if err != nil {
		return nil, err
	}

	var list []*domain.File
	for _, m := range modelList {
		list = append(list, r.toDomain(m, uid))
	}
	return list, nil
}

func (r *fileRepository) ListByFIDsCount(ctx context.Context, fids []int64, vaultID, uid int64) (int64, error) {
	u := r.file(uid).File
	q := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.FID.In(fids...),
		u.Action.Neq(string(domain.FileActionDelete)),
	)

	return q.Count()
}

// ListByIDs retrieves file list by ID list
// ListByIDs 根据ID列表获取文件列表
func (r *fileRepository) ListByIDs(ctx context.Context, ids []int64, uid int64) ([]*domain.File, error) {
	if len(ids) == 0 {
		return []*domain.File{}, nil
	}
	u := r.file(uid).File
	ms, err := u.WithContext(ctx).Where(u.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	var res []*domain.File
	for _, m := range ms {
		res = append(res, r.toDomain(m, uid))
	}
	return res, nil
}

// RecycleClear cleans up the recycle bin
// RecycleClear 清理回收站
func (r *fileRepository) RecycleClear(ctx context.Context, path, pathHash string, vaultID, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.file(uid).File
		q := u.WithContext(ctx).Where(u.VaultID.Eq(vaultID), u.Action.Eq(string(domain.FileActionDelete)), u.Rename.Eq(0))
		if pathHash != "" {
			q = q.Where(u.PathHash.Eq(pathHash))
		}
		_, err := q.UpdateSimple(
			u.Rename.Value(2),
			u.UpdatedTimestamp.Value(timex.Now().UnixMilli()),
			u.UpdatedAt.Value(timex.Now()),
		)
		return err
	})
}

// UpdateFID 仅更新文件的文件夹关联 ID，不更新 updated_timestamp
// 用于 SyncResourceFID 内部整理，避免污染增量同步时间戳
// Only updates the folder ID (FID) without touching updated_timestamp
// Used by SyncResourceFID to avoid polluting incremental sync timestamps
func (r *fileRepository) UpdateFID(ctx context.Context, id, fid, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.file(uid).File
		_, err := u.WithContext(ctx).Where(u.ID.Eq(id)).UpdateSimple(u.FID.Value(fid))
		return err
	})
}

// Ensure fileRepository implements domain.FileRepository interface
// 确保 fileRepository 实现了 domain.FileRepository 接口
var _ domain.FileRepository = (*fileRepository)(nil)

func (r *fileRepository) ListByPathPrefix(ctx context.Context, pathPrefix string, vaultID, uid int64) ([]*domain.File, error) {
	u := r.file(uid).File
	// Use LIKE 'prefix/%'
	// 使用 LIKE 'prefix/%'
	pattern := pathPrefix + "/%"
	ms, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Path.Like(pattern),
		u.Action.Neq(string(domain.FileActionDelete)),
	).Find()
	if err != nil {
		return nil, err
	}
	var res []*domain.File
	for _, m := range ms {
		res = append(res, r.toDomain(m, uid))
	}
	return res, nil
}

// buildFileOrderClause builds file order clause
// buildFileOrderClause 构建文件排序语句
func buildFileOrderClause(sortBy, sortOrder string) string {
	// 默认值
	if sortBy == "" {
		sortBy = "mtime"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// 验证排序方向
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Map sort field
	// 映射排序字段
	var field string
	switch sortBy {
	case "ctime":
		field = "ctime"
	case "path":
		field = "path"
	case "mtime":
		fallthrough
	default:
		field = "mtime"
	}

	return field + " " + sortOrder
}

// DeleteByVaultID physically deletes all files in a vault
// DeleteByVaultID 物理删除仓库下的所有文件
func (r *fileRepository) DeleteByVaultID(ctx context.Context, vaultID, uid int64) error {
	return r.dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		u := r.file(uid).File

		// 查找该仓库下的所有文件 ID
		files, err := u.WithContext(ctx).Where(u.VaultID.Eq(vaultID)).Select(u.ID).Find()
		if err != nil {
			return err
		}

		if len(files) == 0 {
			return nil
		}

		var ids []int64
		for _, f := range files {
			ids = append(ids, f.ID)
		}

		// 从数据库删除
		_, err = u.WithContext(ctx).Where(u.VaultID.Eq(vaultID)).Delete()
		if err != nil {
			return err
		}

		// 删除物理文件夹
		for _, id := range ids {
			folderPath := r.dao.GetFileFolderPath(uid, id)
			_ = r.dao.RemoveContentFolder(folderPath)
		}

		return nil
	})
}
