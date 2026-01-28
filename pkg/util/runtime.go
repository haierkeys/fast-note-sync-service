package util

import "runtime"

func WhoCalled() string {
	// Skip:
	// 0: runtime.Callers
	// 1: whoCalled
	// 2: caller of whoCalled
	// 3: caller of caller of whoCalled -> which is the caller of B that we want (if whoCalled is directly called in B)
	// 3: caller of caller of whoCalled -> 也就是我们想要的 B 的调用者（如果 whoCalled 在 B 内直接被调用）
	pc := make([]uintptr, 1)
	n := runtime.Callers(3, pc)
	if n == 0 {
		return "unknown"
	}
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.Function // Contains full path, for example "main.A" or "main.C"
	// 包含完整路径，例如 "main.A" 或 "main.C"
}
