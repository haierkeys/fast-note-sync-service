package storage

import (
	"io"

	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"github.com/haierkeys/fast-note-sync-service/pkg/storage/local_fs"
)

type Type = string
type CloudType = Type

const LOCAL Type = "localfs"

var StorageTypeMap = map[Type]bool{
	LOCAL: true,
}

var CloudStorageTypeMap = map[Type]bool{}

type Storager interface {
	PutFile(pathKey string, file io.Reader, cType string) (string, error)
	PutContent(pathKey string, content []byte) (string, error)
	DeleteFile(pathKey string) error
}

var Instance map[Type]Storager

func NewClient(cType Type, config map[string]any) (Storager, error) {

	if cType == LOCAL {
		return local_fs.NewClient(config)
	}
	return nil, code.ErrorInvalidStorageType
}

// IsUserEnabledWithConfig check if the storage type is enabled (using injected configuration)
// IsUserEnabledWithConfig 检查存储类型是否启用（使用注入的配置）
func IsUserEnabledWithConfig(cType Type, localFSEnabled bool) error {
	// Check if the cloud storage type is valid
	// 检查云存储类型是否有效
	if !StorageTypeMap[cType] {
		return code.ErrorInvalidCloudStorageType
	}

	if cType == LOCAL && !localFSEnabled {
		return code.ErrorUserLocalFSDisabled
	}
	return nil
}

// GetEnabledStorageTypesWithConfig get enabled storage types (using injected configuration)
// GetEnabledStorageTypesWithConfig 获取启用的存储类型（使用注入的配置）
func GetEnabledStorageTypesWithConfig(localFSEnabled bool) []CloudType {
	var list []CloudType
	if localFSEnabled {
		list = append(list, LOCAL)
	}
	return list
}
