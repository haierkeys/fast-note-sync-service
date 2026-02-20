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
	ID              int64  `form:"id" example:"1"`                                                              // ID
	Type            string `form:"type" binding:"required,gte=1" example:"local-fs"`                            // 类型
	Endpoint        string `form:"endpoint" example:"oss-cn-hangzhou.aliyuncs.com"`                             // 端点 oss
	Region          string `form:"region" example:"us-east-1"`                                                  // 区域 s3
	AccountID       string `form:"accountId" example:"123456789"`                                               // 账户ID r2
	BucketName      string `form:"bucketName" example:"my-bucket"`                                              // 存储桶名称
	AccessKeyID     string `form:"accessKeyId" example:"AKIAIOSFODNN7EXAMPLE"`                                  // 访问密钥ID
	AccessKeySecret string `form:"accessKeySecret" example:"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"`          // 访问密钥秘密
	CustomPath      string `form:"customPath" example:"/backups"`                                               // 自定义路径
	AccessURLPrefix string `form:"accessUrlPrefix"  binding:"required,min=2,max=100" example:"https://cdn.com"` // 访问地址前缀
	User            string `form:"user" example:"admin"`                                                        // 访问用户名
	Password        string `form:"password" example:"secret_password"`                                          // 密码
	IsEnabled       int64  `form:"isEnabled" example:"1"`                                                       // 是否启用
}

// StorageGetRequest 存储配置获取请求
type StorageGetRequest struct {
	ID int64 `json:"id" form:"id" binding:"required" example:"1"`
}
