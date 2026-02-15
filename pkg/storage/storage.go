package storage

import (
	"io"

	"github.com/haierkeys/fast-note-sync-service/global"
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

type Storager interface {
	SendFile(pathKey string, file io.Reader, cType string) (string, error)
	SendContent(pathKey string, content []byte) (string, error)
}

var Instance map[Type]Storager

func NewClient(cType Type, config map[string]any) (Storager, error) {

	if cType == LOCAL {
		return local_fs.NewClient(config)
	} else if cType == OSS {
		return aliyun_oss.NewClient(config)
	} else if cType == R2 {
		return cloudflare_r2.NewClient(config)
	} else if cType == S3 {
		return aws_s3.NewClient(config)
	} else if cType == MinIO {
		return minio.NewClient(config)
	} else if cType == WebDAV {
		return webdav.NewClient(config)
	}
	return nil, code.ErrorInvalidStorageType
}

func IsUserEnabled(cType Type) error {

	// 检查云存储类型是否有效
	if !StorageTypeMap[cType] {
		return code.ErrorInvalidCloudStorageType
	}

	if cType == LOCAL && !global.Config.LocalFS.IsUserEnabled {
		return code.ErrorUserLocalFSDisabled
	} else if cType == OSS && !global.Config.OSS.IsUserEnabled {
		return code.ErrorUserALIOSSDisabled
	} else if cType == R2 && !global.Config.CloudflueR2.IsUserEnabled {
		return code.ErrorUserCloudflueR2Disabled
	} else if cType == S3 && !global.Config.AWSS3.IsUserEnabled {
		return code.ErrorUserAWSS3Disabled
	} else if cType == MinIO && !global.Config.MinIO.IsUserEnabled {
		return code.ErrorUserMinIODisabled
	}
	return nil
}

func GetIsUserEnabledStorageTypes() []CloudType {

	var list []CloudType
	if global.Config.CloudflueR2.IsUserEnabled {
		list = append(list, R2)
	}
	if global.Config.OSS.IsUserEnabled {
		list = append(list, OSS)
	}
	if global.Config.AWSS3.IsUserEnabled {
		list = append(list, S3)
	}
	if global.Config.MinIO.IsUserEnabled {
		list = append(list, MinIO)
	}
	if global.Config.LocalFS.IsUserEnabled {
		list = append(list, LOCAL)
	}
	if global.Config.WebDAV.IsUserEnabled {
		list = append(list, WebDAV)
	}
	return list
}
