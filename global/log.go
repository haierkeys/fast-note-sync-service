package global

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

func Log() *zap.Logger {
	return Logger
}
