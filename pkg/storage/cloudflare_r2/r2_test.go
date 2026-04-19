package cloudflare_r2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	config := &Config{
		AccountID:       "cloudflare-account-id",
		BucketName:      "test-bucket",
		AccessKeyID:     "r2-test-key",
		AccessKeySecret: "r2-test-secret",
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.S3Client) // Still uses s3 aws SDK

	// Check cache
	client2, err := NewClient(config)
	assert.NoError(t, err)
	assert.Equal(t, client, client2)
}
