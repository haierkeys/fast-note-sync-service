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

	entry, created := syncBatchGetOrCreate(ctx, "note", 3)
	if !created {
		t.Fatal("expected created=true for a brand new context+type entry")
	}
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

// TestSyncBatchGetOrCreate_CreatedFlag_S3Regression covers the observability signal added in
// S3 (design §3.3 point 2): a second call for the same context+type must report created=false
// (existing entry reused), and after the entry is deleted (mimicking doSync's syncBatchDelete
// once a batch round completes), a later call for the same context+type must report
// created=true again — this is exactly the "late retransmit rebuilds an orphan entry" case the
// Debug log at the S3 call sites is meant to surface.
func TestSyncBatchGetOrCreate_CreatedFlag_S3Regression(t *testing.T) {
	ctx := "test-ctx-" + t.Name()
	defer syncBatchDelete(ctx, "note")

	_, created := syncBatchGetOrCreate(ctx, "note", 3)
	if !created {
		t.Fatal("first call: expected created=true")
	}

	_, created = syncBatchGetOrCreate(ctx, "note", 3)
	if created {
		t.Fatal("second call for the same context+type: expected created=false (entry reused)")
	}

	syncBatchDelete(ctx, "note")

	_, created = syncBatchGetOrCreate(ctx, "note", 3)
	if !created {
		t.Fatal("call after delete (simulating a late retransmit after doSync collected): expected created=true (orphan rebuild)")
	}
}
