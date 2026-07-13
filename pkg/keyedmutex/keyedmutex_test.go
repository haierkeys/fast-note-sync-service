package keyedmutex

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestKeyedMutex_SerializesSameKey verifies that concurrent Lock calls for the same
// key never overlap, and that every caller actually runs its critical section (unlike
// singleflight, which would let a late caller share an earlier caller's result without
// running its own work).
func TestKeyedMutex_SerializesSameKey(t *testing.T) {
	km := New()
	const n = 50
	var running int32
	var maxConcurrent int32
	var executed int32

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unlock := km.Lock("same-key")
			defer unlock()

			cur := atomic.AddInt32(&running, 1)
			for {
				max := atomic.LoadInt32(&maxConcurrent)
				if cur <= max || atomic.CompareAndSwapInt32(&maxConcurrent, max, cur) {
					break
				}
			}
			atomic.AddInt32(&executed, 1)
			time.Sleep(time.Millisecond)
			atomic.AddInt32(&running, -1)
		}()
	}
	wg.Wait()

	if executed != n {
		t.Fatalf("expected every caller to execute, got %d/%d", executed, n)
	}
	if maxConcurrent != 1 {
		t.Fatalf("expected max concurrency 1 for same key, got %d", maxConcurrent)
	}
}

// TestKeyedMutex_DifferentKeysConcurrent verifies that different keys do not block each other.
func TestKeyedMutex_DifferentKeysConcurrent(t *testing.T) {
	km := New()
	const n = 10
	var wg sync.WaitGroup
	start := make(chan struct{})
	var running int32
	var maxConcurrent int32

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			key := string(rune('a' + idx))
			unlock := km.Lock(key)
			defer unlock()

			cur := atomic.AddInt32(&running, 1)
			for {
				max := atomic.LoadInt32(&maxConcurrent)
				if cur <= max || atomic.CompareAndSwapInt32(&maxConcurrent, max, cur) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&running, -1)
		}(i)
	}
	close(start)
	wg.Wait()

	if maxConcurrent <= 1 {
		t.Fatalf("expected different keys to run concurrently, max concurrency was %d", maxConcurrent)
	}
}

// TestKeyedMutex_TryLock verifies TryLock succeeds when free and fails when contended,
// and that the fallback-to-Lock path still eventually succeeds after the holder releases.
func TestKeyedMutex_TryLock(t *testing.T) {
	km := New()

	unlock1, ok := km.TryLock("k")
	if !ok {
		t.Fatal("expected first TryLock on a free key to succeed")
	}

	if _, ok := km.TryLock("k"); ok {
		t.Fatal("expected TryLock on a held key to fail")
	}

	done := make(chan struct{})
	go func() {
		unlock2 := km.Lock("k")
		defer unlock2()
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("blocking Lock should not have succeeded while key is held")
	case <-time.After(20 * time.Millisecond):
	}

	unlock1()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("blocking Lock did not succeed after holder released")
	}
}

// TestKeyedMutex_CleansUpEntries verifies map entries are removed once unlocked.
func TestKeyedMutex_CleansUpEntries(t *testing.T) {
	km := New()
	unlock := km.Lock("k")
	unlock()

	km.mu.Lock()
	n := len(km.locks)
	km.mu.Unlock()

	if n != 0 {
		t.Fatalf("expected locks map to be empty after unlock, got %d entries", n)
	}
}
