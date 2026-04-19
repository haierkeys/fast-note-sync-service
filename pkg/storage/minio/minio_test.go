package minio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	config := &Config{
		Endpoint:        "http://localhost:9000",
		Region:          "us-east-1",
		BucketName:      "test-bucket",
		AccessKeyID:     "minio-test-key",
		AccessKeySecret: "minio-test-secret",
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Check cache
	client2, err := NewClient(config)
	assert.NoError(t, err)
	assert.Equal(t, client, client2)
}
