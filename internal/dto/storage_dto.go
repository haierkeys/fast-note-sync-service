package dto

import "github.com/haierkeys/fast-note-sync-service/pkg/timex"

// StorageDTO 存储配置 DTO
type StorageDTO struct {
	ID              int64      `json:"id"`
	UID             int64      `json:"uid"`
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
	IsDeleted       bool       `json:"isDeleted"`
	CreatedAt       timex.Time `json:"createdAt"`
	UpdatedAt       timex.Time `json:"updatedAt"`
}

// StoragePostRequest 存储配置创建/更新请求
type StoragePostRequest struct {
	ID      int64       `json:"id" form:"id"`
	Storage *StorageDTO `json:"storage" binding:"required"`
}

// StorageGetRequest 存储配置获取请求
type StorageGetRequest struct {
	ID int64 `json:"id" form:"id" binding:"required"`
}
