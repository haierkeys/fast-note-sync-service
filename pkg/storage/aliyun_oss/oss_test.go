package aliyun_oss

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	config := &Config{
		Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
		Region:          "cn-hangzhou",
		BucketName:      "test-bucket",
		AccessKeyID:     "test-key",
		AccessKeySecret: "test-secret",
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Client)

	// Test auto-extraction
	config2 := &Config{
		Endpoint:        "oss-cn-shanghai-internal.aliyuncs.com",
		BucketName:      "test-bucket",
		AccessKeyID:     "test-key-2",
		AccessKeySecret: "test-secret-2",
	}
	client2, err := NewClient(config2)
	assert.NoError(t, err)
	assert.NotNil(t, client2)

	// Since clients are cached by AccessKeyID, this should return the identical client for the same ID
	client3, err := NewClient(config)
	assert.NoError(t, err)
	assert.Equal(t, client, client3)
}
