package global

import (
	dumpx "github.com/gookit/goutil/dump"
	"go.uber.org/zap"
)

var Logger *zap.Logger

func Log() *zap.Logger {
	return Logger
}

func Dump(a any) {
	dumpx.P(a)
}
