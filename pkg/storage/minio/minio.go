package minio

import (
	"context"

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
}

var clients = make(map[string]*MinIO)

func NewClient(cf map[string]any) (*MinIO, error) {
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
	}
	return clients[accessKeyId], nil
}
