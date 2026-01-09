// Package dao 实现数据访问层
package dao

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

// fileRepository 实现 domain.FileRepository 接口
type fileRepository struct {
	dao *Dao
}

// NewFileRepository 创建 FileRepository 实例
func NewFileRepository(dao *Dao) domain.FileRepository {
	return &fileRepository{dao: dao}
}

// file 获取文件查询对象
func (r *fileRepository) file(uid int64) *query.Query {
	key := "user_file_" + strconv.FormatInt(uid, 10)
	return r.dao.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "File")
	}, key+"#file", key)
}

// toDomain 将数据库模型转换为领域模型
func (r *fileRepository) toDomain(m *model.File, uid int64) *domain.File {
	if m == nil {
		return nil
	}
	file := &domain.File{
		ID:               m.ID,
		VaultID:          m.VaultID,
		Action:           domain.FileAction(m.Action),
		Path:             m.Path,
		PathHash:         m.PathHash,
		ContentHash:      m.ContentHash,
		SavePath:         m.SavePath,
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

// toModel 将领域模型转换为数据库模型
func (r *fileRepository) toModel(file *domain.File) *model.File {
	if file == nil {
		return nil
	}
	return &model.File{
		ID:               file.ID,
		VaultID:          file.VaultID,
		Action:           string(file.Action),
		Path:             file.Path,
		PathHash:         file.PathHash,
		ContentHash:      file.ContentHash,
		SavePath:         file.SavePath,
		Size:             file.Size,
		Ctime:            file.Ctime,
		Mtime:            file.Mtime,
		UpdatedTimestamp: file.UpdatedTimestamp,
		CreatedAt:        timex.Time(file.CreatedAt),
		UpdatedAt:        timex.Time(file.UpdatedAt),
	}
}

// fillFilePath 填充文件的保存路径并处理旧文件迁移
func (r *fileRepository) fillFilePath(uid int64, f *domain.File) {
	if f == nil {
		return
	}
	folderPath := r.dao.GetFileFolderPath(uid, f.ID)
	standardPath := filepath.Join(folderPath, "file.dat")

	// 如果标准路径不存在，检查是否有旧路径文件并迁移
	if _, err := os.Stat(standardPath); os.IsNotExist(err) && f.SavePath != "" && f.SavePath != standardPath {
		if _, errOld := os.Stat(f.SavePath); errOld == nil {
			_ = os.MkdirAll(folderPath, 0755)
			_ = os.Rename(f.SavePath, standardPath)
		} else {
			// 兜底方案：检查文件夹内是否有任何文件，如果有则重命名第一个
			if files, err := os.ReadDir(folderPath); err == nil && len(files) > 0 {
				oldName := files[0].Name()
				if oldName != "file.dat" {
					_ = os.Rename(filepath.Join(folderPath, oldName), standardPath)
				}
			}
		}
	}
	f.SavePath = standardPath
}

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

// Create 创建文件
func (r *fileRepository) Create(ctx context.Context, file *domain.File, uid int64) (*domain.File, error) {
	u := r.file(uid).File
	m := r.toModel(file)

	m.UpdatedTimestamp = timex.Now().UnixMilli()
	m.CreatedAt = timex.Now()
	m.UpdatedAt = timex.Now()

	tempSavePath := m.SavePath
	m.SavePath = "" // 不在数据库中保存路径

	err := u.WithContext(ctx).Create(m)
	if err != nil {
		return nil, err
	}

	// 移动文件到 Vault 目录，固定命名为 file.dat
	if tempSavePath != "" {
		folderPath := r.dao.GetFileFolderPath(uid, m.ID)
		_ = os.MkdirAll(folderPath, 0755)
		finalPath := filepath.Join(folderPath, "file.dat")

		if err := os.Rename(tempSavePath, finalPath); err != nil {
			_ = os.Rename(tempSavePath, finalPath)
		}
	}

	return r.toDomain(m, uid), nil
}

// Update 更新文件
func (r *fileRepository) Update(ctx context.Context, file *domain.File, uid int64) (*domain.File, error) {
	u := r.file(uid).File
	m := r.toModel(file)

	m.UpdatedTimestamp = timex.Now().UnixMilli()
	m.UpdatedAt = timex.Now()

	tempSavePath := m.SavePath
	m.SavePath = "" // 不在数据库中更新路径

	// 如果提供了新的临时路径，则移动到固定的 file.dat
	if tempSavePath != "" {
		folderPath := r.dao.GetFileFolderPath(uid, m.ID)
		_ = os.MkdirAll(folderPath, 0755)
		finalPath := filepath.Join(folderPath, "file.dat")
		_ = os.Rename(tempSavePath, finalPath)
	}

	err := u.WithContext(ctx).Where(
		u.ID.Eq(m.ID),
	).Save(m)

	if err != nil {
		return nil, err
	}
	return r.toDomain(m, uid), nil
}

// UpdateMtime 更新文件修改时间
func (r *fileRepository) UpdateMtime(ctx context.Context, mtime int64, id, uid int64) error {
	u := r.file(uid).File

	_, err := u.WithContext(ctx).Where(
		u.ID.Eq(id),
	).UpdateSimple(
		u.Mtime.Value(mtime),
		u.UpdatedTimestamp.Value(timex.Now().UnixMilli()),
		u.UpdatedAt.Value(timex.Now()),
	)
	return err
}

// Delete 物理删除文件
func (r *fileRepository) Delete(ctx context.Context, id, uid int64) error {
	u := r.file(uid).File
	_, err := u.WithContext(ctx).Where(u.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}

	// 删除物理文件
	folderPath := r.dao.GetFileFolderPath(uid, id)
	_ = r.dao.RemoveContentFolder(folderPath)

	return nil
}

// DeletePhysicalByTime 根据时间物理删除已标记删除的文件
func (r *fileRepository) DeletePhysicalByTime(ctx context.Context, timestamp, uid int64) error {
	u := r.file(uid).File

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
}

// List 分页获取文件列表
func (r *fileRepository) List(ctx context.Context, vaultID int64, page, pageSize int, uid int64) ([]*domain.File, error) {
	u := r.file(uid).File
	modelList, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Action.Neq("delete"),
	).Order(u.Path.Desc(), u.CreatedAt.Desc()).
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

// ListCount 获取文件数量
func (r *fileRepository) ListCount(ctx context.Context, vaultID, uid int64) (int64, error) {
	u := r.file(uid).File
	count, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.Action.Neq("delete"),
	).Order(u.CreatedAt).
		Count()

	if err != nil {
		return 0, err
	}

	return count, nil
}

// ListByUpdatedTimestamp 根据更新时间戳获取文件列表
func (r *fileRepository) ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.File, error) {
	u := r.file(uid).File
	mList, err := u.WithContext(ctx).Where(
		u.VaultID.Eq(vaultID),
		u.UpdatedTimestamp.Gt(timestamp),
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
	).Scan(result)

	if err != nil {
		return nil, err
	}

	return &domain.CountSizeResult{
		Count: result.Count,
		Size:  result.Size,
	}, nil
}

// 确保 fileRepository 实现了 domain.FileRepository 接口
var _ domain.FileRepository = (*fileRepository)(nil)
