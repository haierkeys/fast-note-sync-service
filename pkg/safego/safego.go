// Package safego provides a panic-safe replacement for bare `go func() { ... }()` launches.
// Package safego 为裸 `go func() { ... }()` 启动方式提供带 panic 防护的替代方案。
package safego

import (
	"runtime/debug"

	"go.uber.org/zap"
)

// Go runs f in a new goroutine, recovering any panic so it cannot crash the process.
// The panic value and stack trace are logged via logger (falls back to zap.L() if nil).
// Go 在新 goroutine 中运行 f，recover 掉其中的 panic，避免其导致进程崩溃。
// panic 值与堆栈会通过 logger 记录（logger 为 nil 时退回使用 zap.L()）。
func Go(logger *zap.Logger, f func()) {
	if logger == nil {
		logger = zap.L()
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered in background goroutine",
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())),
				)
			}
		}()
		f()
	}()
}
