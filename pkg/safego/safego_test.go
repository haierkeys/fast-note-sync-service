package safego

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// TestGo_RecoversPanicAndLogs verifies that a panic inside the goroutine is recovered
// (does not crash the test process) and logged.
//
// Note: f's own deferred statements unwind (and thus run) before Go's recover/log runs,
// so a WaitGroup released from inside f cannot be used to wait for the log to be written;
// poll instead.
// 注意：f 自身的 defer 会在 Go 的 recover/log 执行之前完成（属于 panic 展开的一部分），
// 所以不能用 f 内部释放的 WaitGroup 来等待日志写入完成，这里改为轮询等待。
func TestGo_RecoversPanicAndLogs(t *testing.T) {
	core, logs := observer.New(zap.ErrorLevel)
	logger := zap.New(core)

	Go(logger, func() {
		panic("boom")
	})

	assert.Eventually(t, func() bool {
		return logs.Len() == 1
	}, time.Second, time.Millisecond)

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}
	if entries[0].Message != "panic recovered in background goroutine" {
		t.Fatalf("unexpected log message: %q", entries[0].Message)
	}
}

// TestGo_RunsNormallyWithoutPanic verifies f actually runs when it doesn't panic.
func TestGo_RunsNormallyWithoutPanic(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	ran := false
	Go(zap.NewNop(), func() {
		defer wg.Done()
		ran = true
	})
	wg.Wait()
	if !ran {
		t.Fatal("expected f to run")
	}
}

// TestGo_NilLoggerFallsBack verifies passing a nil logger does not panic the caller.
func TestGo_NilLoggerFallsBack(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	Go(nil, func() {
		defer wg.Done()
		panic("boom")
	})
	wg.Wait()
}
