package aws_s3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	config := &Config{
		Region:          "us-east-1",
		BucketName:      "test-bucket",
		AccessKeyID:     "aws-test-key",
		AccessKeySecret: "aws-test-secret",
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.S3Client)
	assert.NotNil(t, client.TransferManager)

	// Check cache
	client2, err := NewClient(config)
	assert.NoError(t, err)
	assert.Equal(t, client, client2)
}
