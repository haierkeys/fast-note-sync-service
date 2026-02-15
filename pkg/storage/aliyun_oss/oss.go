package aliyun_oss

import (
	oss_sdk "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Config struct {
	IsEnabled       bool   `yaml:"is-enable"`
	IsUserEnabled   bool   `yaml:"is-user-enable"`
	Endpoint        string `yaml:"endpoint"`
	BucketName      string `yaml:"bucket-name"`
	AccessKeyID     string `yaml:"access-key-id"`
	AccessKeySecret string `yaml:"access-key-secret"`
	CustomPath      string `yaml:"custom-path"`
}

type OSS struct {
	Client *oss_sdk.Client
	Bucket *oss_sdk.Bucket
	Config *Config
}

var clients = make(map[string]*OSS)

func NewClient(cf map[string]any) (*OSS, error) {

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
		BucketName:      cf["BucketName"].(string),
		AccessKeyID:     cf["AccessKeyID"].(string),
		AccessKeySecret: cf["AccessKeySecret"].(string),
		CustomPath:      cf["CustomPath"].(string),
	}

	var id = conf.AccessKeyID
	var endpoint = conf.Endpoint
	var accessKeyId = conf.AccessKeyID
	var accessKeySecret = conf.AccessKeySecret

	var err error
	if clients[id] != nil {
		return clients[id], nil
	}
	// New client
	ossClient, err := oss_sdk.New(endpoint, accessKeyId, accessKeySecret)
	if err != nil {
		return nil, err
	}
	clients[id] = &OSS{
		Client: ossClient,
		Config: conf,
	}
	return clients[id], nil
}
