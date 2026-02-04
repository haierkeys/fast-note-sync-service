package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/internal/query"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

type folderRepository struct {
	*Dao
}

func NewFolderRepository(d *Dao) domain.FolderRepository {
	return &folderRepository{Dao: d}
}

func (r *folderRepository) GetKey(uid int64) string {
	return "user_folder_" + fmt.Sprintf("%d", uid)
}

// folder returns the query with auto-migration
func (r *folderRepository) folder(uid int64) *query.Query {
	return r.Dao.UseQueryWithOnceFunc(func(g *gorm.DB) {
		model.AutoMigrate(g, "Folder")
	}, r.GetKey(uid)+"#folder", r.GetKey(uid))
}

func (r *folderRepository) GetByID(ctx context.Context, id, uid int64) (*domain.Folder, error) {
	f := r.folder(uid).Folder
	m, err := f.WithContext(ctx).Where(f.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return r.modelToDomain(m), nil
}

func (r *folderRepository) GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*domain.Folder, error) {
	f := r.folder(uid).Folder
	m, err := f.WithContext(ctx).Where(f.VaultID.Eq(vaultID), f.PathHash.Eq(pathHash)).First()
	if err != nil {
		return nil, err
	}
	return r.modelToDomain(m), nil
}

func (r *folderRepository) GetAllByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) ([]*domain.Folder, error) {
	f := r.folder(uid).Folder
	ms, err := f.WithContext(ctx).Where(f.VaultID.Eq(vaultID), f.PathHash.Eq(pathHash), f.Action.Neq("delete")).Find()
	if err != nil {
		return nil, err
	}
	var res []*domain.Folder
	for _, m := range ms {
		res = append(res, r.modelToDomain(m))
	}
	return res, nil
}

func (r *folderRepository) GetByFID(ctx context.Context, fid int64, vaultID, uid int64) ([]*domain.Folder, error) {
	var ms []*model.Folder
	f := r.folder(uid).Folder
	ms, err := f.WithContext(ctx).Where(f.VaultID.Eq(vaultID), f.FID.Eq(fid), f.Action.Neq("delete")).Find()
	if err != nil {
		return nil, err
	}
	var res []*domain.Folder
	for _, m := range ms {
		res = append(res, r.modelToDomain(m))
	}
	return res, nil
}

func (r *folderRepository) Create(ctx context.Context, folder *domain.Folder, uid int64) (*domain.Folder, error) {
	var result *domain.Folder
	err := r.Dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		m := r.domainToModel(folder)
		m.CreatedAt = timex.Now()
		m.UpdatedAt = timex.Now()
		m.UpdatedTimestamp = time.Now().UnixMilli()
		f := r.folder(uid).Folder
		err := f.WithContext(ctx).Create(m)
		if err != nil {
			return err
		}
		result = r.modelToDomain(m)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *folderRepository) Update(ctx context.Context, folder *domain.Folder, uid int64) (*domain.Folder, error) {
	var result *domain.Folder
	err := r.Dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		m := r.domainToModel(folder)
		m.UpdatedAt = timex.Now()
		m.UpdatedTimestamp = time.Now().UnixMilli()
		f := r.folder(uid).Folder
		_, err := f.WithContext(ctx).Where(f.ID.Eq(m.ID)).Updates(m)
		if err != nil {
			return err
		}
		result = r.modelToDomain(m)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *folderRepository) Delete(ctx context.Context, id, uid int64) error {
	return r.Dao.ExecuteWrite(ctx, uid, r, func(db *gorm.DB) error {
		f := r.folder(uid).Folder
		_, err := f.WithContext(ctx).Where(f.ID.Eq(id)).Delete()
		return err
	})
}

func (r *folderRepository) ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.Folder, error) {
	var ms []*model.Folder
	f := r.folder(uid).Folder
	ms, err := f.WithContext(ctx).Where(f.VaultID.Eq(vaultID), f.UpdatedTimestamp.Gt(timestamp)).Find()
	if err != nil {
		return nil, err
	}
	var res []*domain.Folder
	for _, m := range ms {
		res = append(res, r.modelToDomain(m))
	}
	return res, nil
}

func (r *folderRepository) ListByPathPrefix(ctx context.Context, pathPrefix string, vaultID, uid int64) ([]*domain.Folder, error) {
	var ms []*model.Folder
	f := r.folder(uid).Folder
	// 使用 LIKE 'prefix/%' 来查找所有子目录
	pattern := pathPrefix + "/%"
	ms, err := f.WithContext(ctx).Where(f.VaultID.Eq(vaultID), f.Path.Like(pattern), f.Action.Neq("delete")).Find()
	if err != nil {
		return nil, err
	}
	var res []*domain.Folder
	for _, m := range ms {
		res = append(res, r.modelToDomain(m))
	}
	return res, nil
}

func (r *folderRepository) modelToDomain(m *model.Folder) *domain.Folder {
	if m == nil {
		return nil
	}
	return &domain.Folder{
		ID:               m.ID,
		VaultID:          m.VaultID,
		Action:           domain.FolderAction(m.Action),
		Path:             m.Path,
		PathHash:         m.PathHash,
		Level:            m.Level,
		FID:              m.FID,
		Ctime:            m.Ctime,
		Mtime:            m.Mtime,
		UpdatedTimestamp: m.UpdatedTimestamp,
		CreatedAt:        time.Time(m.CreatedAt),
		UpdatedAt:        time.Time(m.UpdatedAt),
	}
}

func (r *folderRepository) domainToModel(d *domain.Folder) *model.Folder {
	if d == nil {
		return nil
	}
	return &model.Folder{
		ID:               d.ID,
		VaultID:          d.VaultID,
		Action:           string(d.Action),
		Path:             d.Path,
		PathHash:         d.PathHash,
		Level:            d.Level,
		FID:              d.FID,
		Ctime:            d.Ctime,
		Mtime:            d.Mtime,
		UpdatedTimestamp: d.UpdatedTimestamp,
		CreatedAt:        timex.Time(d.CreatedAt),
		UpdatedAt:        timex.Time(d.UpdatedAt),
	}
}
