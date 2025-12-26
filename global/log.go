package global

import (
	"fmt"
	"runtime"

	dumpx "github.com/gookit/goutil/dump"
	"go.uber.org/zap"
)

var Logger *zap.Logger

func Log() *zap.Logger {
	return Logger
}

func Dump(a ...any) {
	_, file, line, ok := runtime.Caller(1)
	if ok {
		fmt.Printf("\033[32m%s:%d:\033[0m\n", file, line)
	}
	dumpx.P(a...)
}
