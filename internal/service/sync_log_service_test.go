package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/domain"
	domainmocks "github.com/haierkeys/fast-note-sync-service/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// TestSyncLogService_Log_FlushesOnBatchSize verifies that once syncLogBatchMaxSize entries
// for the same uid have been queued, they are flushed via a single CreateBatch call rather
// than one goroutine + one DB write per Log() call.
func TestSyncLogService_Log_FlushesOnBatchSize(t *testing.T) {
	repo := new(domainmocks.MockSyncLogRepository)

	var mu sync.Mutex
	var batches [][]*domain.SyncLog
	done := make(chan struct{}, 1)
	repo.On("CreateBatch", mock.Anything, mock.Anything, int64(1)).
		Run(func(args mock.Arguments) {
			logs := args.Get(1).([]*domain.SyncLog)
			mu.Lock()
			batches = append(batches, logs)
			mu.Unlock()
			select {
			case done <- struct{}{}:
			default:
			}
		}).
		Return(nil)

	svc := NewSyncLogService(repo, zap.NewNop())
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = svc.Shutdown(ctx)
	}()

	for i := 0; i < syncLogBatchMaxSize; i++ {
		svc.Log(1, 1, domain.SyncLogTypeNote, domain.SyncLogActionModify, "content", "a.md", "hash-a", "web", "web", "1.0", 10)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("expected CreateBatch to be called once the batch size threshold was reached")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(batches) != 1 {
		t.Fatalf("expected exactly 1 batch flush, got %d", len(batches))
	}
	assert.Len(t, batches[0], syncLogBatchMaxSize)
}

// TestSyncLogService_Log_FlushesOnTimer verifies that a small number of queued entries
// (below the batch-size threshold) still get flushed after the flush interval elapses.
func TestSyncLogService_Log_FlushesOnTimer(t *testing.T) {
	repo := new(domainmocks.MockSyncLogRepository)

	done := make(chan []*domain.SyncLog, 1)
	repo.On("CreateBatch", mock.Anything, mock.Anything, int64(2)).
		Run(func(args mock.Arguments) {
			done <- args.Get(1).([]*domain.SyncLog)
		}).
		Return(nil)

	svc := NewSyncLogService(repo, zap.NewNop())
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = svc.Shutdown(ctx)
	}()

	svc.Log(2, 1, domain.SyncLogTypeFile, domain.SyncLogActionCreate, "", "b.md", "hash-b", "web", "web", "1.0", 5)

	select {
	case logs := <-done:
		assert.Len(t, logs, 1)
	case <-time.After(2 * time.Second):
		t.Fatal("expected CreateBatch to be called after the flush interval elapsed")
	}
}

// TestSyncLogService_Shutdown_FlushesBufferedEntries verifies that Shutdown flushes any
// entries still buffered (queued but not yet auto-flushed) instead of dropping them.
func TestSyncLogService_Shutdown_FlushesBufferedEntries(t *testing.T) {
	repo := new(domainmocks.MockSyncLogRepository)

	var flushed []*domain.SyncLog
	repo.On("CreateBatch", mock.Anything, mock.Anything, int64(3)).
		Run(func(args mock.Arguments) {
			flushed = args.Get(1).([]*domain.SyncLog)
		}).
		Return(nil)

	svc := NewSyncLogService(repo, zap.NewNop())

	svc.Log(3, 1, domain.SyncLogTypeNote, domain.SyncLogActionDelete, "", "c.md", "hash-c", "web", "web", "1.0", 1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := svc.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}

	assert.Len(t, flushed, 1, "Shutdown must flush buffered entries instead of dropping them")
}

// TestSyncLogService_Log_DropsWhenChannelFull verifies that once the bounded channel is
// full, further Log() calls do not block and the entry is dropped (with a warning, not
// asserted here) instead of growing unbounded.
func TestSyncLogService_Log_DropsWhenChannelFull(t *testing.T) {
	repo := new(domainmocks.MockSyncLogRepository)
	repo.On("CreateBatch", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	s := &syncLogService{
		repo:   repo,
		logger: zap.NewNop(),
		ch:     make(chan syncLogQueueItem, 1), // tiny buffer, no worker draining it // 极小缓冲，且无 worker 消费
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		// First fills the buffer, remaining calls must not block even though nothing drains it.
		for i := 0; i < 10; i++ {
			s.Log(1, 1, domain.SyncLogTypeNote, domain.SyncLogActionModify, "", "x.md", "hx", "web", "web", "1.0", 1)
		}
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Log() blocked instead of dropping entries once the channel was full")
	}
}
