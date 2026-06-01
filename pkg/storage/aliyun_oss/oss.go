package aliyun_oss

import (
	"regexp"
	"strings"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

func cacheKey(conf *Config) string {
	return conf.AccessKeyID + ":" + conf.AccessKeySecret + ":" + conf.Endpoint + ":" + conf.Region
}

type Config struct {
	Endpoint        string `yaml:"endpoint"`
	Region          string `yaml:"region"`
	BucketName      string `yaml:"bucket-name"`
	AccessKeyID     string `yaml:"access-key-id"`
	AccessKeySecret string `yaml:"access-key-secret"`
	CustomPath      string `yaml:"custom-path"`
}

type OSS struct {
	Client *oss.Client
	Config *Config
}

var clients = make(map[string]*OSS)

func NewClient(conf *Config) (*OSS, error) {
	key := cacheKey(conf)
	if clients[key] != nil {
		return clients[key], nil
	}

	region := conf.Region
	if region == "" && conf.Endpoint != "" {
		region = extractRegionFromEndpoint(conf.Endpoint)
	}

	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(conf.AccessKeyID, conf.AccessKeySecret)).
		WithRegion(region)

	if conf.Endpoint != "" {
		cfg.WithEndpoint(conf.Endpoint)
	}

	ossClient := oss.NewClient(cfg)

	clients[key] = &OSS{
		Client: ossClient,
		Config: conf,
	}
	return clients[key], nil
}

func extractRegionFromEndpoint(endpoint string) string {
	// Pattern to match region in standard OSS endpoints
	// oss-cn-hangzhou.aliyuncs.com -> cn-hangzhou
	// oss-cn-hangzhou-internal.aliyuncs.com -> cn-hangzhou
	re := regexp.MustCompile(`oss-([a-z0-9-]+)(?:-internal)?\.aliyuncs\.com`)
	matches := re.FindStringSubmatch(endpoint)
	if len(matches) > 1 {
		return matches[1]
	}

	// Try to handle format like cn-hangzhou.oss.aliyuncs.com
	if strings.Contains(endpoint, ".oss.aliyuncs.com") {
		parts := strings.Split(endpoint, ".")
		if len(parts) > 0 {
			return parts[0]
		}
	}

	return ""
}
