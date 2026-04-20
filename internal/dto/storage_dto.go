package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// StorageDTO Storage configuration DTO
// StorageDTO 存储配置 DTO
type StorageDTO struct {
	ID              int64      `json:"id"`              // ID // ID
	UID             int64      `json:"-"`               // User UID // 用户 ID
	Type            string     `json:"type"`            // Storage type // 存储类型
	Endpoint        string     `json:"endpoint"`        // Endpoint // 访问端点
	Region          string     `json:"region"`          // Region // 区域
	AccountID       string     `json:"accountId"`       // Account ID // 账户 ID
	BucketName      string     `json:"bucketName"`      // Bucket name // 存储桶名称
	AccessKeyID     string     `json:"accessKeyId"`     // Access key ID // 访问密钥 ID
	AccessKeySecret string     `json:"accessKeySecret"` // Access key secret // 访问密钥秘密
	CustomPath      string     `json:"customPath"`      // Custom path // 自定义路径
	AccessURLPrefix string     `json:"accessUrlPrefix"` // Access URL prefix // 访问地址前缀
	User            string     `json:"user"`            // Username // 用户名
	Password        string     `json:"password"`        // Password // 密码
	IsEnabled       bool       `json:"isEnabled"`       // Is enabled // 是否启用
	IsDeleted       bool       `json:"-"`               // Is deleted // 是否已删除
	CreatedAt       timex.Time `json:"createdAt"`       // Created at // 创建时间
	UpdatedAt       timex.Time `json:"updatedAt"`       // Updated at // 更新时间
}

// StoragePostRequest Storage configuration create/update request
// StoragePostRequest 存储配置创建/更新请求
type StoragePostRequest struct {
	ID              int64  `form:"id" example:"1"`                                                              // ID // ID
	Type            string `form:"type" binding:"required,gte=1" example:"local-fs"`                            // Storage type // 类型
	Endpoint        string `form:"endpoint" example:"oss-cn-hangzhou.aliyuncs.com"`                             // Endpoint (OSS) // 端点 oss
	Region          string `form:"region" example:"us-east-1"`                                                  // Region (S3) // 区域 s3
	AccountID       string `form:"accountId" example:"123456789"`                                               // Account ID (R2) // 账户ID r2
	BucketName      string `form:"bucketName" example:"my-bucket"`                                              // Bucket name // 存储桶名称
	AccessKeyID     string `form:"accessKeyId" example:""`                                                      // Access key ID // 访问密钥ID
	AccessKeySecret string `form:"accessKeySecret" example:""`                                                  // Access key secret // 访问密钥秘密
	CustomPath      string `form:"customPath" example:"/backups"`                                               // Custom path // 自定义路径
	AccessURLPrefix string `form:"accessUrlPrefix"  binding:"required,min=2,max=100" example:"https://cdn.com"` // Access URL prefix // 访问地址前缀
	User            string `form:"user" example:"admin"`                                                        // Username // 访问用户名
	Password        string `form:"password" example:"secret_password"`                                          // Password // 密码
	IsEnabled       int64  `form:"isEnabled" example:"1"`                                                       // Is enabled // 是否启用
}

// StorageGetRequest Storage configuration retrieval request
// StorageGetRequest 存储配置获取请求
type StorageGetRequest struct {
	ID int64 `json:"id" form:"id" binding:"required" example:"1"` // ID // ID
}
