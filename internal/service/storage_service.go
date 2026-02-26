package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/haierkeys/fast-note-sync-service/internal/config"
	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	"github.com/haierkeys/fast-note-sync-service/internal/dto"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/storage"
	"github.com/haierkeys/fast-note-sync-service/pkg/timex"
	"gorm.io/gorm"
)

// StorageService defines the business service interface for Storage
// StorageService 定义存储业务服务接口
type StorageService interface {
	// CreateOrUpdate creates or updates a storage configuration
	// CreateOrUpdate 创建或更新存储配置
	CreateOrUpdate(ctx context.Context, uid int64, id int64, storageDTO *dto.StoragePostRequest) (*dto.StorageDTO, error)

	// Get retrieves storage configuration by ID
	// Get 获取存储配置
	Get(ctx context.Context, uid int64, id int64) (*dto.StorageDTO, error)

	// List retrieves storage configuration list for current user
	// List 获取当前用户的存储配置列表
	List(ctx context.Context, uid int64) ([]*dto.StorageDTO, error)

	// Delete deletes storage configuration
	// Delete 删除存储配置
	Delete(ctx context.Context, uid int64, id int64) error

	// GetEnabledTypes returns list of enabled storage types
	// GetEnabledTypes 获取启用的存储类型列表
	GetEnabledTypes() ([]string, error)

	// Validate 通过发送测试文件并删除来验证存储连通性
	Validate(ctx context.Context, req *dto.StoragePostRequest) error
}

type storageService struct {
	repo   domain.StorageRepository
	config *config.StorageConfig
}

// NewStorageService creates StorageService instance
// NewStorageService 创建 StorageService 实例
func NewStorageService(repo domain.StorageRepository, config *config.StorageConfig) StorageService {
	return &storageService{repo: repo, config: config}
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
		IsEnabled:       m.IsEnabled,
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
		IsEnabled:       d.IsEnabled,
		IsDeleted:       d.IsDeleted,
	}
}

func (s *storageService) postRequestToDomain(req *dto.StoragePostRequest) *domain.Storage {
	if req == nil {
		return nil
	}
	return &domain.Storage{
		ID:              req.ID,
		Type:            req.Type,
		Endpoint:        req.Endpoint,
		Region:          req.Region,
		AccountID:       req.AccountID,
		BucketName:      req.BucketName,
		AccessKeyID:     req.AccessKeyID,
		AccessKeySecret: req.AccessKeySecret,
		CustomPath:      req.CustomPath,
		AccessURLPrefix: req.AccessURLPrefix,
		User:            req.User,
		Password:        req.Password,
		IsEnabled:       req.IsEnabled == 1,
	}
}

func (s *storageService) CreateOrUpdate(ctx context.Context, uid int64, id int64, req *dto.StoragePostRequest) (*dto.StorageDTO, error) {
	// Validate storage type availability
	// 验证存储类型可用性
	typeName := req.Type
	if !s.isStorageTypeEnabled(typeName) {
		return nil, code.ErrorStorageTypeDisabled
	}

	storage := s.postRequestToDomain(req)
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

func (s *storageService) GetEnabledTypes() ([]string, error) {
	var types []string
	if s.config.LocalFS.IsEnabled {
		types = append(types, string(storage.LOCAL))
	}
	if s.config.AliyunOSS.IsEnabled {
		types = append(types, string(storage.OSS))
	}
	if s.config.AwsS3.IsEnabled {
		types = append(types, string(storage.S3))
	}
	if s.config.CloudflareR2.IsEnabled {
		types = append(types, string(storage.R2))
	}
	if s.config.MinIO.IsEnabled {
		types = append(types, string(storage.MinIO))
	}
	if s.config.WebDAV.IsEnabled {
		types = append(types, string(storage.WebDAV))
	}
	return types, nil
}

func (s *storageService) isStorageTypeEnabled(t string) bool {
	switch storage.Type(t) {
	case storage.LOCAL:
		return s.config.LocalFS.IsEnabled
	case storage.OSS:
		return s.config.AliyunOSS.IsEnabled
	case storage.S3:
		return s.config.AwsS3.IsEnabled
	case storage.R2:
		return s.config.CloudflareR2.IsEnabled
	case storage.MinIO:
		return s.config.MinIO.IsEnabled
	case storage.WebDAV:
		return s.config.WebDAV.IsEnabled
	default:
		return false
	}
}

func (s *storageService) Validate(ctx context.Context, req *dto.StoragePostRequest) error {
	if !s.isStorageTypeEnabled(req.Type) {
		return code.ErrorStorageTypeDisabled
	}

	sConfig := &storage.Config{
		Type:            req.Type,
		CustomPath:      req.CustomPath,
		Endpoint:        req.Endpoint,
		Region:          req.Region,
		BucketName:      req.BucketName,
		AccessKeyID:     req.AccessKeyID,
		AccessKeySecret: req.AccessKeySecret,
		AccountID:       req.AccountID,
		User:            req.User,
		Password:        req.Password,
		SavePath:        s.config.LocalFS.SavePath,
	}

	client, err := storage.NewClient(sConfig)
	if err != nil {
		return code.ErrorStorageValidateFailed.WithDetails(err.Error())
	}

	testFile := fmt.Sprintf(".fast-note-test-%s", uuid.New().String()[:8])
	if _, err := client.SendContent(testFile, []byte("ok"), time.Now()); err != nil {
		return code.ErrorStorageValidateFailed.WithDetails(err.Error())
	}

	if err := client.Delete(testFile); err != nil {
		return code.ErrorStorageValidateFailed.WithDetails(err.Error())
	}

	return nil
}

var _ StorageService = (*storageService)(nil)
