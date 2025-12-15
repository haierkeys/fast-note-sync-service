package task

import (
	"context"
	"os"
	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
	"go.uber.org/zap"
)

// TempFileCleanupTask 临时文件清理任务
type TempFileCleanupTask struct {
}

// Name 任务名称
func (t *TempFileCleanupTask) Name() string {
	return "temp_file_cleanup"
}

// Interval 执行间隔 (0 表示不进行周期性执行)
func (t *TempFileCleanupTask) Interval() time.Duration {
	return 0
}

// RunImmediately 是否立即执行一次
func (t *TempFileCleanupTask) RunImmediately() bool {
	return true
}

// Run 执行任务
func (t *TempFileCleanupTask) Run(ctx context.Context) error {
	tempDir := global.Config.App.TempPath
	if tempDir == "" {
		tempDir = "storage/temp"
	}

	global.Logger.Info("starting temp file cleanup", zap.String("path", tempDir))

	// 检查目录是否存在
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		global.Logger.Info("temp directory does not exist, skipping cleanup", zap.String("path", tempDir))
		return nil
	}

	// 删除整个目录
	if err := os.RemoveAll(tempDir); err != nil {
		global.Logger.Error("failed to remove temp directory", zap.String("path", tempDir), zap.Error(err))
		return err
	}

	// 重新创建目录
	if err := os.MkdirAll(tempDir, 0644); err != nil {
		global.Logger.Error("failed to recreate temp directory", zap.String("path", tempDir), zap.Error(err))
		return err
	}

	global.Logger.Info("temp file cleanup completed successfully")
	return nil
}

func init() {
	Register(func() (Task, error) {
		return &TempFileCleanupTask{}, nil
	})
}
