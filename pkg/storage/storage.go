package storage

import (
	"io"
	"time"

	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/storage/aliyun_oss"
	"github.com/haierkeys/fast-note-sync-service/pkg/storage/aws_s3"
	"github.com/haierkeys/fast-note-sync-service/pkg/storage/cloudflare_r2"
	"github.com/haierkeys/fast-note-sync-service/pkg/storage/local_fs"
	"github.com/haierkeys/fast-note-sync-service/pkg/storage/minio"
	"github.com/haierkeys/fast-note-sync-service/pkg/storage/webdav"
)

type Type = string
type CloudType = Type

const OSS CloudType = "oss"
const R2 CloudType = "r2"
const S3 CloudType = "s3"
const LOCAL Type = "localfs"
const MinIO CloudType = "minio"
const WebDAV CloudType = "webdav"

var StorageTypeMap = map[Type]bool{
	OSS:    true,
	R2:     true,
	S3:     true,
	LOCAL:  true,
	MinIO:  true,
	WebDAV: true,
}

var CloudStorageTypeMap = map[Type]bool{
	OSS:   true,
	R2:    true,
	S3:    true,
	MinIO: true,
}

// Config Unified storage configuration
type Config struct {
	Type Type `yaml:"type"`

	// Common settings
	CustomPath string `yaml:"custom-path"`

	// Cloud Storage (S3/OSS/MinIO/R2)
	Endpoint        string `yaml:"endpoint"`
	Region          string `yaml:"region"`
	BucketName      string `yaml:"bucket-name"`
	AccessKeyID     string `yaml:"access-key-id"`
	AccessKeySecret string `yaml:"access-key-secret"`
	AccountID       string `yaml:"account-id"` // Cloudflare R2 specific

	// WebDAV
	User     string `yaml:"user"`
	Password string `yaml:"password"`

	// Local FS
	SavePath string `yaml:"save-path"`
}

type Storager interface {
	SendFile(pathKey string, file io.Reader, cType string, modTime time.Time) (string, error)
	SendContent(pathKey string, content []byte, modTime time.Time) (string, error)
	Delete(pathKey string) error
}

func NewClient(config *Config) (Storager, error) {
	if config == nil {
		return nil, code.ErrorInvalidStorageType
	}

	cType := config.Type
	if cType == LOCAL {
		cfg := &local_fs.Config{
			CustomPath: config.CustomPath,
			SavePath:   config.SavePath,
		}
		return local_fs.NewClient(cfg)
	} else if cType == OSS {
		cfg := &aliyun_oss.Config{
			Endpoint:        config.Endpoint,
			BucketName:      config.BucketName,
			AccessKeyID:     config.AccessKeyID,
			AccessKeySecret: config.AccessKeySecret,
			CustomPath:      config.CustomPath,
		}
		return aliyun_oss.NewClient(cfg)
	} else if cType == R2 {
		cfg := &cloudflare_r2.Config{
			AccountID:       config.AccountID,
			BucketName:      config.BucketName,
			AccessKeyID:     config.AccessKeyID,
			AccessKeySecret: config.AccessKeySecret,
			CustomPath:      config.CustomPath,
		}
		return cloudflare_r2.NewClient(cfg)
	} else if cType == S3 {
		cfg := &aws_s3.Config{
			Region:          config.Region,
			BucketName:      config.BucketName,
			AccessKeyID:     config.AccessKeyID,
			AccessKeySecret: config.AccessKeySecret,
			CustomPath:      config.CustomPath,
		}
		return aws_s3.NewClient(cfg)
	} else if cType == MinIO {
		cfg := &minio.Config{
			Endpoint:        config.Endpoint,
			Region:          config.Region,
			BucketName:      config.BucketName,
			AccessKeyID:     config.AccessKeyID,
			AccessKeySecret: config.AccessKeySecret,
			CustomPath:      config.CustomPath,
		}
		return minio.NewClient(cfg)
	} else if cType == WebDAV {
		cfg := &webdav.Config{
			Endpoint:   config.Endpoint,
			User:       config.User,
			Password:   config.Password,
			CustomPath: config.CustomPath,
		}
		return webdav.NewClient(cfg)
	}
	return nil, code.ErrorInvalidStorageType
}
