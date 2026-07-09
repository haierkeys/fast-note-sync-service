package websocket_router

import (
	"sync"
	"testing"
)

// TestSyncBatchEntry_MarkBatchReceived_DedupsRetransmit verifies the P1 idempotency fix:
// the same BatchIndex arriving twice (client retransmit after a lost ack) must be reported
// as a duplicate on the second call, so callers skip append/count and only resend the ack.
func TestSyncBatchEntry_MarkBatchReceived_DedupsRetransmit(t *testing.T) {
	e := &syncBatchEntry{}

	if e.markBatchReceived(0) {
		t.Fatal("first receipt of batchIndex 0 must not be reported as duplicate")
	}
	if !e.markBatchReceived(0) {
		t.Fatal("second receipt of batchIndex 0 (retransmit) must be reported as duplicate")
	}
	if e.markBatchReceived(1) {
		t.Fatal("first receipt of a different batchIndex must not be reported as duplicate")
	}
	if !e.markBatchReceived(1) {
		t.Fatal("retransmit of batchIndex 1 must be reported as duplicate")
	}
}

// TestSyncBatchEntry_MarkBatchReceived_ConcurrentSafe exercises markBatchReceived under
// the same mutex-holding discipline the call sites use, verifying it does not race or panic.
func TestSyncBatchEntry_MarkBatchReceived_ConcurrentSafe(t *testing.T) {
	e := &syncBatchEntry{}
	var wg sync.WaitGroup
	dupCount := make([]int, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// Each goroutine sends the same batchIndex twice, mimicking a retransmit race.
			for j := 0; j < 2; j++ {
				e.mu.Lock()
				if e.markBatchReceived(idx) {
					dupCount[idx]++
				}
				e.mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	for idx, dups := range dupCount {
		if dups != 1 {
			t.Fatalf("batchIndex %d: expected exactly 1 duplicate detection out of 2 sends, got %d", idx, dups)
		}
	}
}

// TestSyncBatchGetOrCreate_InitializesReceivedIndexes verifies entries created via
// syncBatchGetOrCreate have a non-nil ReceivedIndexes map ready to use.
func TestSyncBatchGetOrCreate_InitializesReceivedIndexes(t *testing.T) {
	ctx := "test-ctx-" + t.Name()
	defer syncBatchDelete(ctx, "note")

	entry := syncBatchGetOrCreate(ctx, "note", 3)
	if entry.ReceivedIndexes == nil {
		t.Fatal("expected ReceivedIndexes to be initialized")
	}
	if entry.markBatchReceived(0) {
		t.Fatal("first receipt should not be a duplicate")
	}
	if !entry.markBatchReceived(0) {
		t.Fatal("second receipt should be a duplicate")
	}
}
