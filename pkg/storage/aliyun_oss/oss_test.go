package aliyun_oss

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	config := &Config{
		Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
		BucketName:      "test-bucket",
		AccessKeyID:     "test-key",
		AccessKeySecret: "test-secret",
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Client)

	// Since clients are cached by AccessKeyID, this should return the identical client
	client2, err := NewClient(config)
	assert.NoError(t, err)
	assert.Equal(t, client, client2)
}
