package cloudflare_r2

import (
	"context"
	"fmt"

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
	AccountId       string `yaml:"account-id"`
	BucketName      string `yaml:"bucket-name"`
	AccessKeyID     string `yaml:"access-key-id"`
	AccessKeySecret string `yaml:"access-key-secret"`
	CustomPath      string `yaml:"custom-path"`
}

type R2 struct {
	S3Client  *s3.Client
	S3Manager *manager.Uploader
	Config    *Config
	logger    *zap.Logger
}

// Option configuration option function type
// Option 配置选项函数类型
type Option func(*R2)

// WithLogger sets the logger
// WithLogger 设置日志器
func WithLogger(logger *zap.Logger) Option {
	return func(r *R2) {
		r.logger = logger
	}
}

var clients = make(map[string]*R2)

// NewClient creates an R2 storage instance
// NewClient 创建 R2 存储实例
// opts is optional parameters for configuring logger and other options
// opts 可选参数用于配置日志器等选项
func NewClient(cf map[string]any, opts ...Option) (*R2, error) {

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
		AccountId:       cf["AccountId"].(string),
		BucketName:      cf["BucketName"].(string),
		AccessKeyID:     cf["AccessKeyID"].(string),
		AccessKeySecret: cf["AccessKeySecret"].(string),
		CustomPath:      cf["CustomPath"].(string),
	}

	var accountId = conf.AccountId
	var accessKeyId = conf.AccessKeyID
	var accessKeySecret = conf.AccessKeySecret

	if clients[accessKeyId] != nil {
		// Apply options to existing client
		// 应用选项到已存在的客户端
		for _, opt := range opts {
			opt(clients[accessKeyId])
		}
		return clients[accessKeyId], nil
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {

		return nil, errors.Wrap(err, "cloudflare_r2")
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
	})

	clients[accessKeyId] = &R2{
		S3Client: client,
		Config:   conf,
		logger:   zap.NewNop(), // Default Nop logger
		// 默认空日志器
	}
	// Apply options
	// 应用选项
	for _, opt := range opts {
		opt(clients[accessKeyId])
	}
	return clients[accessKeyId], nil
}
