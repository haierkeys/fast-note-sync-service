package safe_close

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSafeClose_BasicFlow(t *testing.T) {
	sc := NewSafeClose()

	var counter int
	var mu sync.Mutex

	// Attach a mock service goroutine
	sc.Attach(func(done func(), closeSignal <-chan struct{}) {
		defer done()
		<-closeSignal // Block until closed
		mu.Lock()
		counter++
		mu.Unlock()
	})

	// Give it time to start
	time.Sleep(10 * time.Millisecond)

	expectedErr := errors.New("closing service")
	sc.SendCloseSignal(expectedErr)

	// Wait for services to finish
	err := sc.WaitClosed()

	assert.Equal(t, expectedErr, err)

	mu.Lock()
	assert.Equal(t, 1, counter, "Goroutine should have progressed after close signal")
	mu.Unlock()
}

func TestSafeClose_AttachAfterClose(t *testing.T) {
	sc := NewSafeClose()
	sc.SendCloseSignal(nil)

	err := sc.WaitClosed()
	assert.NoError(t, err)

	var run bool
	sc.Attach(func(done func(), closeSignal <-chan struct{}) {
		// Should not reach here
		defer done()
		run = true
	})

	time.Sleep(10 * time.Millisecond)
	assert.False(t, run, "Attached function should not run if SafeClose is already closed")
}

func TestSafeClose_MultipleSend(t *testing.T) {
	sc := NewSafeClose()
	
	firstErr := errors.New("first error")
	sc.SendCloseSignal(firstErr)

	// Second consecutive call should be noop and safely ignored
	sc.SendCloseSignal(errors.New("second error should be ignored"))

	err := sc.WaitClosed()
	assert.Equal(t, firstErr, err)
}
