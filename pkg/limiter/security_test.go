package limiter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestEncodedLoginSharesRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := NewMethodLimiter().AddBuckets(BucketRule{Key: "/api/user/login", FillInterval: time.Hour, Capacity: 1, Quantum: 1})
	r := gin.New()
	r.POST("/api/user/login", func(c *gin.Context) {
		bucket, ok := l.GetBucket(l.Key(c))
		if !ok || bucket.TakeAvailable(1) == 0 {
			if !ok {
				t.Error("routed login bypassed its rate-limit bucket")
			}
			c.Status(http.StatusTooManyRequests)
			return
		}
		c.Status(http.StatusUnauthorized)
	})
	for _, path := range []string{"/api/user/login", "/api/user/l%6fgin", "/api/user/log%69n?x=1"} {
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, path, nil))
	}
}

func TestClientLoginBucketsDoNotLockOtherClients(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := NewMethodLimiter().AddBuckets(BucketRule{Key: "/api/user/login", PerClient: true, FillInterval: time.Hour, Capacity: 1, Quantum: 1})
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.POST("/api/user/login", func(c *gin.Context) {
		bucket, ok := l.GetBucket(l.Key(c))
		if !ok {
			t.Fatal("login rate-limit bucket missing")
		}
		if bucket.TakeAvailable(1) == 0 {
			c.Status(http.StatusTooManyRequests)
			return
		}
		c.Status(http.StatusUnauthorized)
	})
	for _, test := range []struct {
		ip, path string
		status   int
	}{
		{"192.0.2.1:1000", "/api/user/login", http.StatusUnauthorized},
		{"192.0.2.1:1000", "/api/user/l%6fgin", http.StatusTooManyRequests},
		{"192.0.2.2:1000", "/api/user/login", http.StatusUnauthorized},
		{"192.0.2.2:1000", "/api/user/log%69n?x=1", http.StatusTooManyRequests},
	} {
		request := httptest.NewRequest(http.MethodPost, test.path, nil)
		request.RemoteAddr = test.ip
		request.Header.Set("X-Forwarded-For", "198.51.100.99")
		response := httptest.NewRecorder()
		r.ServeHTTP(response, request)
		if response.Code != test.status {
			t.Fatalf("%s %s: status %d, expected %d", test.ip, test.path, response.Code, test.status)
		}
	}
}

func TestClientBucketMemoryBoundAndConcurrency(t *testing.T) {
	l := NewMethodLimiter().AddBuckets(BucketRule{Key: "/api/user/login", PerClient: true, FillInterval: time.Second, Capacity: 5, Quantum: 1}).(*MethodLimiter)
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < maxClientBuckets; i++ {
				bucket, ok := l.GetBucket(fmt.Sprintf("/api/user/login\x00%d-%d", worker, i))
				if !ok {
					t.Error("bucket unexpectedly absent")
					return
				}
				bucket.TakeAvailable(1)
			}
		}(worker)
	}
	wg.Wait()
	if len(l.clients) != maxClientBuckets || l.order.Len() != maxClientBuckets {
		t.Fatal("client cache is not bounded")
	}
	for _, element := range l.clients {
		element.Value.(*clientBucket).lastUsed = time.Now().Add(-2 * clientBucketIdle)
	}
	l.GetBucket("/api/user/login\x00fresh")
	if len(l.clients) != 1 {
		t.Fatal("idle buckets were not removed")
	}
}
