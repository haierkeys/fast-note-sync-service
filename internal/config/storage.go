package config

// StorageConfig Storage configuration
// StorageConfig 存储配置
type StorageConfig struct {
	LocalFS      StorageLocalFSConfig `yaml:"local-fs"`
	AliyunOSS    StorageBaseConfig    `yaml:"aliyun-oss"`
	AwsS3        StorageBaseConfig    `yaml:"aws-s3"`
	CloudflareR2 StorageBaseConfig    `yaml:"cloudflare-r2"`
	MinIO        StorageBaseConfig    `yaml:"minio"`
	WebDAV       StorageBaseConfig    `yaml:"webdav"`
}

// StorageLocalFSConfig Local file system storage configuration
// StorageLocalFSConfig 本地文件系统存储配置
type StorageLocalFSConfig struct {
	IsEnabled      bool   `yaml:"is-enable" default:"false"`       // Default false as per user requirement, but logically might need enabled
	HttpfsIsEnable bool   `yaml:"httpfs-is-enable" default:"true"` // Default true
	SavePath       string `yaml:"save-path" default:"storage/uploads"`
}

// StorageBaseConfig Base configuration for cloud storages
// StorageBaseConfig 云存储基础配置
type StorageBaseConfig struct {
	IsEnabled bool `yaml:"is-enable" default:"true"` // Default enabled
}
