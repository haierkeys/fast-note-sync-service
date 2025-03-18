package storage

import (
	"io"

	"github.com/haierkeys/obsidian-better-sync-service/global"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/code"
	"github.com/haierkeys/obsidian-better-sync-service/pkg/storage/local_fs"
)

type Type = string
type CloudType = Type

const LOCAL Type = "localfs"

var StorageTypeMap = map[Type]bool{
	LOCAL: true,
}

var CloudStorageTypeMap = map[Type]bool{}

type Storager interface {
	SendFile(pathKey string, file io.Reader, cType string) (string, error)
	SendContent(pathKey string, content []byte) (string, error)
}

var Instance map[Type]Storager

func NewClient(cType Type, config map[string]any) (Storager, error) {

	if cType == LOCAL {
		return local_fs.NewClient(config)
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
	}
	return nil
}

func GetIsUserEnabledStorageTypes() []CloudType {

	var list []CloudType
	if global.Config.LocalFS.IsUserEnabled {
		list = append(list, LOCAL)
	}
	return list
}
