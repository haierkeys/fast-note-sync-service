package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// StorageDTO 存储配置 DTO
type StorageDTO struct {
	ID              int64      `json:"id"`
	UID             int64      `json:"-"`
	Type            string     `json:"type"`
	Endpoint        string     `json:"endpoint"`
	Region          string     `json:"region"`
	AccountID       string     `json:"accountId"`
	BucketName      string     `json:"bucketName"`
	AccessKeyID     string     `json:"accessKeyId"`
	AccessKeySecret string     `json:"accessKeySecret"`
	CustomPath      string     `json:"customPath"`
	AccessURLPrefix string     `json:"accessUrlPrefix"`
	User            string     `json:"user"`
	Password        string     `json:"password"`
	IsEnabled       bool       `json:"isEnabled"`
	IsDeleted       bool       `json:"-"`
	CreatedAt       timex.Time `json:"createdAt"`
	UpdatedAt       timex.Time `json:"updatedAt"`
}

// StoragePostRequest 存储配置创建/更新请求
type StoragePostRequest struct {
	ID              int64  `form:"id"`                                                // ID
	Type            string `form:"type" binding:"required,gte=1"`                     // 类型
	Endpoint        string `form:"endpoint"`                                          // 端点 oss
	Region          string `form:"region"`                                            // 区域 s3
	AccountID       string `form:"accountId"`                                         // 账户ID r2
	BucketName      string `form:"bucketName"`                                        // 存储桶名称
	AccessKeyID     string `form:"accessKeyId"`                                       // 访问密钥ID
	AccessKeySecret string `form:"accessKeySecret"`                                   // 访问密钥秘密
	CustomPath      string `form:"customPath"`                                        // 自定义路径
	AccessURLPrefix string `form:"accessUrlPrefix"  binding:"required,min=2,max=100"` // 访问地址前缀
	User            string `form:"user"`                                              // 访问用户名
	Password        string `form:"password"`                                          // 密码
	IsEnabled       int64  `form:"isEnabled"`                                         // 是否启用
}

// StorageGetRequest 存储配置获取请求
type StorageGetRequest struct {
	ID int64 `json:"id" form:"id" binding:"required"`
}
