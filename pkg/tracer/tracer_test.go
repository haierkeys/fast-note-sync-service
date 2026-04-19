package tracer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewJaegerTracer(t *testing.T) {
	// Depending on Jaeger client config, using invalid host/port may immediately fail or just silently ignore udp
	tracer, closer, err := NewJaegerTracer("test-service", "127.0.0.1:0")

	assert.NoError(t, err)
	assert.NotNil(t, tracer)
	assert.NotNil(t, closer)

	// Clean up
	if closer != nil {
		closer.Close()
	}
}
