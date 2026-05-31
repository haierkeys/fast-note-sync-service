package config

// AttachmentStaticConfig Attachment static access configuration
// AttachmentStaticConfig 附件模拟静态访问配置
type AttachmentStaticConfig struct {
	IsEnable      bool                `yaml:"is-enable" default:"false"` // Whether to enable attachment static access // 是否启用附件模拟静态访问
	AllowedVaults map[string][]string `yaml:"allowed-vaults"`             // Allowed users and vault names // 允许访问的用户和库名白名单
	AllowedTypes  []string            `yaml:"allowed-types"`              // Allowed file extension types // 允许访问的文件后缀类型列表
}
