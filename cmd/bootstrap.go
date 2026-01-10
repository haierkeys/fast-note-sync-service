package cmd

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// bootstrapLogger 启动阶段日志器
// 用于在主日志器初始化之前记录启动过程中的日志
var bootstrapLogger *zap.Logger

func init() {
	// 创建控制台输出的 encoder 配置
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 创建控制台输出
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleWriter := zapcore.Lock(os.Stderr)

	// 根据 DEBUG 环境变量设置日志级别
	level := zapcore.InfoLevel
	if os.Getenv("DEBUG") != "" {
		level = zapcore.DebugLevel
	}

	core := zapcore.NewCore(consoleEncoder, consoleWriter, level)
	bootstrapLogger = zap.New(core, zap.AddCaller())
}

// BootstrapLogger 获取启动阶段日志器
func BootstrapLogger() *zap.Logger {
	return bootstrapLogger
}
