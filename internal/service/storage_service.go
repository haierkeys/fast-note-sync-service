package service

import (
	"context"
	"errors"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

// StorageService defines the business service interface for Storage
type StorageService interface {
	// CreateOrUpdate creates or updates a storage configuration
	CreateOrUpdate(ctx context.Context, uid int64, id int64, storageDTO *dto.StorageDTO) (*dto.StorageDTO, error)

	// Get retrieves storage configuration by ID
	Get(ctx context.Context, uid int64, id int64) (*dto.StorageDTO, error)

	// List retrieves storage configuration list for current user
	List(ctx context.Context, uid int64) ([]*dto.StorageDTO, error)

	// Delete deletes storage configuration
	Delete(ctx context.Context, uid int64, id int64) error
}

type storageService struct {
	repo domain.StorageRepository
}

// NewStorageService creates StorageService instance
func NewStorageService(repo domain.StorageRepository) StorageService {
	return &storageService{repo: repo}
}

func (s *storageService) domainToDTO(m *domain.Storage) *dto.StorageDTO {
	if m == nil {
		return nil
	}
	return &dto.StorageDTO{
		ID:              m.ID,
		UID:             m.UID,
		Type:            m.Type,
		Endpoint:        m.Endpoint,
		Region:          m.Region,
		AccountID:       m.AccountID,
		BucketName:      m.BucketName,
		AccessKeyID:     m.AccessKeyID,
		AccessKeySecret: m.AccessKeySecret,
		CustomPath:      m.CustomPath,
		AccessURLPrefix: m.AccessURLPrefix,
		User:            m.User,
		Password:        m.Password,
		IsDeleted:       m.IsDeleted,
		CreatedAt:       timex.Time(m.CreatedAt),
		UpdatedAt:       timex.Time(m.UpdatedAt),
	}
}

func (s *storageService) dtoToDomain(d *dto.StorageDTO) *domain.Storage {
	if d == nil {
		return nil
	}
	return &domain.Storage{
		ID:              d.ID,
		UID:             d.UID,
		Type:            d.Type,
		Endpoint:        d.Endpoint,
		Region:          d.Region,
		AccountID:       d.AccountID,
		BucketName:      d.BucketName,
		AccessKeyID:     d.AccessKeyID,
		AccessKeySecret: d.AccessKeySecret,
		CustomPath:      d.CustomPath,
		AccessURLPrefix: d.AccessURLPrefix,
		User:            d.User,
		Password:        d.Password,
		IsDeleted:       d.IsDeleted,
	}
}

func (s *storageService) CreateOrUpdate(ctx context.Context, uid int64, id int64, storageDTO *dto.StorageDTO) (*dto.StorageDTO, error) {
	storage := s.dtoToDomain(storageDTO)
	storage.UID = uid

	var result *domain.Storage
	var err error

	if id > 0 {
		storage.ID = id
		result, err = s.repo.Update(ctx, storage, uid)
	} else {
		result, err = s.repo.Create(ctx, storage, uid)
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorStorageNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	return s.domainToDTO(result), nil
}

func (s *storageService) Get(ctx context.Context, uid int64, id int64) (*dto.StorageDTO, error) {
	result, err := s.repo.GetByID(ctx, id, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.ErrorStorageNotFound
		}
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}
	return s.domainToDTO(result), nil
}

func (s *storageService) List(ctx context.Context, uid int64) ([]*dto.StorageDTO, error) {
	results, err := s.repo.List(ctx, uid)
	if err != nil {
		return nil, code.ErrorDBQuery.WithDetails(err.Error())
	}

	dtos := make([]*dto.StorageDTO, 0, len(results))
	for _, r := range results {
		dtos = append(dtos, s.domainToDTO(r))
	}
	return dtos, nil
}

func (s *storageService) Delete(ctx context.Context, uid int64, id int64) error {
	err := s.repo.Delete(ctx, id, uid)
	if err != nil {
		return code.ErrorDBQuery.WithDetails(err.Error())
	}
	return nil
}

var _ StorageService = (*storageService)(nil)
