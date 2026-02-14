package domain

import "time"

// Storage 存储配置领域模型
type Storage struct {
	ID              int64
	UID             int64
	Type            string
	Endpoint        string
	Region          string
	AccountID       string
	BucketName      string
	AccessKeyID     string
	AccessKeySecret string
	CustomPath      string
	AccessURLPrefix string
	User            string
	Password        string
	IsDeleted       bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
