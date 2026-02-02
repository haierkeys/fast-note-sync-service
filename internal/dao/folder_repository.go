package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/model"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
)

type folderRepository struct {
	*Dao
}

func NewFolderRepository(d *Dao) domain.FolderRepository {
	return &folderRepository{Dao: d}
}

func (r *folderRepository) GetKey(uid int64) string {
	return "user_" + fmt.Sprintf("%d", uid)
}

func (r *folderRepository) GetByID(ctx context.Context, id, uid int64) (*domain.Folder, error) {
	f := r.UseQuery(r.GetKey(uid)).Folder
	m, err := f.WithContext(ctx).Where(f.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return r.modelToDomain(m), nil
}

func (r *folderRepository) GetByPathHash(ctx context.Context, pathHash string, vaultID, uid int64) (*domain.Folder, error) {
	f := r.UseQuery(r.GetKey(uid)).Folder
	m, err := f.WithContext(ctx).Where(f.VaultID.Eq(vaultID), f.PathHash.Eq(pathHash)).First()
	if err != nil {
		return nil, err
	}
	return r.modelToDomain(m), nil
}

func (r *folderRepository) GetByFID(ctx context.Context, fid int64, vaultID, uid int64) ([]*domain.Folder, error) {
	var ms []*model.Folder
	f := r.UseQuery(r.GetKey(uid)).Folder
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
	m := r.domainToModel(folder)
	m.CreatedAt = timex.Now()
	m.UpdatedAt = timex.Now()
	f := r.UseQuery(r.GetKey(uid)).Folder
	err := f.WithContext(ctx).Create(m)
	if err != nil {
		return nil, err
	}
	return r.modelToDomain(m), nil
}

func (r *folderRepository) Update(ctx context.Context, folder *domain.Folder, uid int64) (*domain.Folder, error) {
	m := r.domainToModel(folder)
	m.UpdatedAt = timex.Now()
	f := r.UseQuery(r.GetKey(uid)).Folder
	_, err := f.WithContext(ctx).Where(f.ID.Eq(m.ID)).Updates(m)
	if err != nil {
		return nil, err
	}
	return r.modelToDomain(m), nil
}

func (r *folderRepository) Delete(ctx context.Context, id, uid int64) error {
	f := r.UseQuery(r.GetKey(uid)).Folder
	_, err := f.WithContext(ctx).Where(f.ID.Eq(id)).Delete()
	return err
}

func (r *folderRepository) ListByUpdatedTimestamp(ctx context.Context, timestamp, vaultID, uid int64) ([]*domain.Folder, error) {
	var ms []*model.Folder
	f := r.UseQuery(r.GetKey(uid)).Folder
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
		UpdatedTimestamp: d.UpdatedTimestamp,
		CreatedAt:        timex.Time(d.CreatedAt),
		UpdatedAt:        timex.Time(d.UpdatedAt),
	}
}
