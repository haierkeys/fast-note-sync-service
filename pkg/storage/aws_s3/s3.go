package aws_s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Config struct {
	IsEnabled       bool   `yaml:"is-enable"`
	Region          string `yaml:"region"`
	BucketName      string `yaml:"bucket-name"`
	AccessKeyID     string `yaml:"access-key-id"`
	AccessKeySecret string `yaml:"access-key-secret"`
	CustomPath      string `yaml:"custom-path"`
}

type S3 struct {
	S3Client  *s3.Client
	S3Manager *manager.Uploader
	Config    *Config
	logger    *zap.Logger
}

// Option 配置选项函数类型
type Option func(*S3)

// WithLogger 设置日志器
func WithLogger(logger *zap.Logger) Option {
	return func(s *S3) {
		s.logger = logger
	}
}

var clients = make(map[string]*S3)

// NewClient 创建 S3 存储实例
// opts 可选参数用于配置日志器等选项
func NewClient(cf map[string]any, opts ...Option) (*S3, error) {
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
		Region:          cf["Region"].(string),
		BucketName:      cf["BucketName"].(string),
		AccessKeyID:     cf["AccessKeyID"].(string),
		AccessKeySecret: cf["AccessKeySecret"].(string),
		CustomPath:      cf["CustomPath"].(string),
	}

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
		return nil, errors.Wrap(err, "aws_s3")
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {})

	if err != nil {
		return nil, errors.Wrap(err, "aws_s3")
	}

	clients[accessKeyId] = &S3{
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
