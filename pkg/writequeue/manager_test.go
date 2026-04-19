package writequeue

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManager_ExecuteSequential(t *testing.T) {
	manager := New(&Config{QueueCapacity: 10, WriteTimeout: time.Second}, nil)
	defer manager.Shutdown(context.Background())

	var counter atomic.Int32
	var sequentialCheck []int32

	// We submit 3 write jobs for the SAME identifier
	// Since write queue is per-user, they MUST execute one after another in FIFO
	err1 := manager.Execute(context.Background(), "user-1", func() error {
		sequentialCheck = append(sequentialCheck, counter.Add(1))
		return nil
	})
	assert.NoError(t, err1)

	err2 := manager.Execute(context.Background(), "user-1", func() error {
		sequentialCheck = append(sequentialCheck, counter.Add(1))
		return nil
	})
	assert.NoError(t, err2)

	assert.Equal(t, []int32{1, 2}, sequentialCheck)
}

func TestManager_MultipleUsers(t *testing.T) {
	manager := New(&Config{QueueCapacity: 10, WriteTimeout: time.Second}, nil)
	defer manager.Shutdown(context.Background())

	ch := make(chan bool, 2)
	
	// They don't block each other
	go func() {
		manager.Execute(context.Background(), "user-1", func() error {
			time.Sleep(50 * time.Millisecond)
			return nil
		})
		ch <- true
	}()

	go func() {
		manager.Execute(context.Background(), "user-2", func() error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
		ch <- true
	}()

	<-ch
	<-ch
	// It should reach here cleanly
}

func TestManager_Shutdown(t *testing.T) {
	manager := New(&Config{QueueCapacity: 10, WriteTimeout: time.Second}, nil)
	
	manager.Shutdown(context.Background())
	assert.True(t, manager.IsClosed())

	err := manager.Execute(context.Background(), "user-1", func() error { return nil })
	assert.Equal(t, ErrWriteQueueClosed, err)
}
