package limiter

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMethodLimiter_Key(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		url      string
		expected string
	}{
		{"/api/v1/users", "/api/v1/users"},
		{"/api/v1/users?name=test&age=20", "/api/v1/users"},
		{"/", "/"},
		{"/?q=1", "/"},
	}

	limiter := NewMethodLimiter()

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest("GET", tt.url, nil)
			req.RequestURI = tt.url
			c.Request = req

			key := limiter.Key(c)
			assert.Equal(t, tt.expected, key)
		})
	}
}

func TestMethodLimiter_Buckets(t *testing.T) {
	limiter := NewMethodLimiter()

	// Initially empty
	bucket, ok := limiter.GetBucket("/api/test")
	assert.False(t, ok)
	assert.Nil(t, bucket)

	// Add bucket rule
	rule := BucketRule{
		Key:          "/api/test",
		FillInterval: 1 * time.Second,
		Capacity:     10,
		Quantum:      1,
	}

	limiter.AddBuckets(rule)

	// After addition
	bucket, ok = limiter.GetBucket("/api/test")
	assert.True(t, ok)
	assert.NotNil(t, bucket)
	assert.Equal(t, int64(10), bucket.Capacity())

	// Add another rule
	rule2 := BucketRule{
		Key:          "/api/other",
		FillInterval: 1 * time.Second,
		Capacity:     5,
		Quantum:      1,
	}
	limiter.AddBuckets(rule2)

	bucket2, ok := limiter.GetBucket("/api/other")
	assert.True(t, ok)
	assert.Equal(t, int64(5), bucket2.Capacity())
}
