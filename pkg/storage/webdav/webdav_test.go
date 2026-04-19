package webdav

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	config := &Config{
		Endpoint:   "http://localhost:8080/webdav",
		User:       "user",
		Password:   "password",
		CustomPath: "/my-notes",
	}

	client, err := NewClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Client)

	// Since NewClient calls c.Connect() right away, it might not return an error immediately based on gowebdav behavior,
	// but it creates the instance properly.

	// Check cache using the composed key
	client2, err := NewClient(config)
	assert.NoError(t, err)
	assert.Equal(t, client, client2)
}
