package cmd

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// bootstrapLogger bootstrap stage logger
// bootstrapLogger 启动阶段日志器
// Used to record logs during the startup process before the main logger is initialized
// 用于在主日志器初始化之前记录启动过程中的日志
var bootstrapLogger *zap.Logger

func init() {
	// Create encoder configuration for console output
	// 创建控制台输出的 encoder 配置
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create console output
	// 创建控制台输出
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleWriter := zapcore.Lock(os.Stderr)

	// Set log level based on DEBUG environment variable
	// 根据 DEBUG 环境变量设置日志级别
	level := zapcore.InfoLevel
	if os.Getenv("DEBUG") != "" {
		level = zapcore.DebugLevel
	}

	core := zapcore.NewCore(consoleEncoder, consoleWriter, level)
	bootstrapLogger = zap.New(core, zap.AddCaller())
}

// BootstrapLogger gets the bootstrap stage logger
// BootstrapLogger 获取启动阶段日志器
func BootstrapLogger() *zap.Logger {
	return bootstrapLogger
}
