package workerpool

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPool_Submit(t *testing.T) {
	pool := New(&Config{MaxWorkers: 2, QueueSize: 10}, nil)
	defer pool.Shutdown(context.Background())

	var counter atomic.Int32

	// Submit multiple tasks
	for i := 0; i < 5; i++ {
		err := pool.Submit(context.Background(), func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond) // Simulate work
			counter.Add(1)
			return nil
		})
		assert.NoError(t, err)
	}

	assert.Equal(t, int32(5), counter.Load())
}

func TestPool_SubmitAsync(t *testing.T) {
	pool := New(&Config{MaxWorkers: 2, QueueSize: 10}, nil)
	defer pool.Shutdown(context.Background())

	var counter atomic.Int32

	for i := 0; i < 5; i++ {
		err := pool.SubmitAsync(context.Background(), func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond) // Simulate work
			counter.Add(1)
			return nil
		})
		assert.NoError(t, err)
	}

	// Because it's async, we shouldn't measure instantly, so wait for all to finish
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int32(5), counter.Load())
}

func TestPool_Shutdown(t *testing.T) {
	pool := New(&Config{MaxWorkers: 2, QueueSize: 10}, nil)
	
	err := pool.Shutdown(context.Background())
	assert.NoError(t, err)
	
	assert.True(t, pool.IsClosed())

	errSubmit := pool.Submit(context.Background(), func(ctx context.Context) error { return nil })
	assert.Equal(t, ErrWorkerPoolClosed, errSubmit)
}
