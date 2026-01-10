package minio

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Config struct {
	IsEnabled       bool   `yaml:"is-enable"`
	BucketName      string `yaml:"bucket-name"`
	Endpoint        string `yaml:"endpoint"`
	Region          string `yaml:"region"`
	AccessKeyID     string `yaml:"access-key-id"`
	AccessKeySecret string `yaml:"access-key-secret"`
	CustomPath      string `yaml:"custom-path"`
}

type MinIO struct {
	S3Client  *s3.Client
	S3Manager *manager.Uploader
	Config    *Config
	logger    *zap.Logger
}

// Option 配置选项函数类型
type Option func(*MinIO)

// WithLogger 设置日志器
func WithLogger(logger *zap.Logger) Option {
	return func(m *MinIO) {
		m.logger = logger
	}
}

var clients = make(map[string]*MinIO)

// NewClient 创建 MinIO 存储实例
// opts 可选参数用于配置日志器等选项
func NewClient(cf map[string]any, opts ...Option) (*MinIO, error) {
	// New client

	var IsEnabled bool
	switch t := cf["IsEnabled"].(type) {
	case int64:
		if t == 0 {
			IsEnabled = false
		} else {
			IsEnabled = true
		}
	case bool:
		IsEnabled = t
	}

	conf := &Config{
		IsEnabled:       IsEnabled,
		Endpoint:        cf["Endpoint"].(string),
		Region:          cf["Region"].(string),
		BucketName:      cf["BucketName"].(string),
		AccessKeyID:     cf["AccessKeyID"].(string),
		AccessKeySecret: cf["AccessKeySecret"].(string),
		CustomPath:      cf["CustomPath"].(string),
	}

	var endpoint = conf.Endpoint
	var region = conf.Region
	var accessKeyId = conf.AccessKeyID
	var accessKeySecret = conf.AccessKeySecret

	if clients[accessKeyId] != nil {
		// 应用选项到已存在的客户端
		for _, opt := range opts {
			opt(clients[accessKeyId])
		}
		return clients[accessKeyId], nil
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
		config.WithRegion(region),
	)

	if err != nil {
		return nil, errors.Wrap(err, "minio")
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(endpoint)
	})

	if err != nil {
		return nil, errors.Wrap(err, "minio")
	}

	clients[accessKeyId] = &MinIO{
		S3Client: client,
		Config:   conf,
		logger:   zap.NewNop(), // 默认空日志器
	}
	// 应用选项
	for _, opt := range opts {
		opt(clients[accessKeyId])
	}
	return clients[accessKeyId], nil
}
