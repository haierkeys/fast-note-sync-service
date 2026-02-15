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
)

type Config struct {
	IsEnabled       bool   `yaml:"is-enable"`
	IsUserEnabled   bool   `yaml:"is-user-enable"`
	AccountID       string `yaml:"account-id"`
	BucketName      string `yaml:"bucket-name"`
	AccessKeyID     string `yaml:"access-key-id"`
	AccessKeySecret string `yaml:"access-key-secret"`
	CustomPath      string `yaml:"custom-path"`
}

type R2 struct {
	S3Client  *s3.Client
	S3Manager *manager.Uploader
	Config    *Config
}

var clients = make(map[string]*R2)

func NewClient(cf map[string]any) (*R2, error) {

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

	var IsUserEnabled bool
	switch t := cf["IsUserEnabled"].(type) {
	case int64:
		if t == 0 {
			IsUserEnabled = false
		} else {
			IsUserEnabled = true
		}
	case bool:
		IsUserEnabled = t
	}

	conf := &Config{
		IsEnabled:       IsEnabled,
		IsUserEnabled:   IsUserEnabled,
		AccountID:       cf["AccountID"].(string),
		BucketName:      cf["BucketName"].(string),
		AccessKeyID:     cf["AccessKeyID"].(string),
		AccessKeySecret: cf["AccessKeySecret"].(string),
		CustomPath:      cf["CustomPath"].(string),
	}

	var accountId = conf.AccountID
	var accessKeyId = conf.AccessKeyID
	var accessKeySecret = conf.AccessKeySecret

	if clients[accessKeyId] != nil {
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

	if err != nil {
		return nil, errors.Wrap(err, "cloudflare_r2")
	}
	clients[accessKeyId] = &R2{
		S3Client: client,
		Config:   conf,
	}
	return clients[accessKeyId], nil
}
