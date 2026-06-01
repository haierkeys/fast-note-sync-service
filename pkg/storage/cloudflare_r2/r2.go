package cloudflare_r2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
)

func validateAccountID(accountID string) error {
	if accountID == "" {
		return fmt.Errorf("Invalid R2 AccountID: AccountID is empty. The AccountID should be your Cloudflare Account ID (a hex string like abc123def456), not a custom domain. You can find your Account ID in the Cloudflare R2 dashboard settings.")
	}
	for _, c := range accountID {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return fmt.Errorf("Invalid R2 AccountID format '%s'. The AccountID should be your Cloudflare Account ID (a hex string like abc123def456), not a custom domain. You can find your Account ID in the Cloudflare R2 dashboard settings.", accountID)
		}
	}
	return nil
}

func cacheKey(conf *Config) string {
	return conf.AccessKeyID + ":" + conf.AccessKeySecret + ":" + conf.AccountID
}

type Config struct {
	AccountID       string `yaml:"account-id"`
	BucketName      string `yaml:"bucket-name"`
	AccessKeyID     string `yaml:"access-key-id"`
	AccessKeySecret string `yaml:"access-key-secret"`
	CustomPath      string `yaml:"custom-path"`
}

type R2 struct {
	S3Client        *s3.Client
	TransferManager *transfermanager.Client
	Config          *Config
}

var clients = make(map[string]*R2)

func NewClient(conf *Config) (*R2, error) {
	var accountId = conf.AccountID
	var accessKeyId = conf.AccessKeyID
	var accessKeySecret = conf.AccessKeySecret

	if err := validateAccountID(accountId); err != nil {
		return nil, err
	}

	key := cacheKey(conf)
	if clients[key] != nil {
		return clients[key], nil
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

	clients[key] = &R2{
		S3Client:        client,
		TransferManager: transfermanager.New(client),
		Config:          conf,
	}
	return clients[key], nil
}
