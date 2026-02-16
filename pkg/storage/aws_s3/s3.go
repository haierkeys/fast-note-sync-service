package aws_s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
)

type Config struct {
	IsEnabled       bool   `yaml:"is-enable"`
	IsUserEnabled   bool   `yaml:"is-user-enable"`
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
}

var clients = make(map[string]*S3)

func NewClient(conf *Config) (*S3, error) {
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
		return nil, errors.Wrap(err, "aws_s3")
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {})

	if err != nil {
		return nil, errors.Wrap(err, "aws_s3")
	}

	clients[accessKeyId] = &S3{
		S3Client: client,
		Config:   conf,
	}
	return clients[accessKeyId], nil
}
