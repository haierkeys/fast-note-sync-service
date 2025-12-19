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
	firstRun bool
}

// Name 任务名称
func (t *TempFileCleanupTask) Name() string {
	return "TempDirStartCleanupTask"
}

// Interval 执行间隔 (0 表示不进行周期性执行)
func (t *TempFileCleanupTask) Interval() time.Duration {
	return 0
}

// RunImmediately 是否立即执行一次
func (t *TempFileCleanupTask) RunImmediately() bool {
	return true
}

// Run 执行清理任务
func (t *TempFileCleanupTask) Run(ctx context.Context) error {
	status := "scheduled"
	if t.firstRun {
		status = "first-run"
		t.firstRun = false
	}

	tempDir := global.Config.App.UploadSavePath
	if tempDir == "" {
		tempDir = "storage/temp"
	}

	var err error

	global.Logger.Info("starting temp file cleanup ["+status+"]", zap.String("path", tempDir))

	// 检查目录是否存在
	if _, err = os.Stat(tempDir); os.IsNotExist(err) {
		global.Logger.Error(t.Name()+" failed ["+status+"]: temp directory does not exist, skipping cleanup", zap.String("path", tempDir))
		return err
	}

	// 删除整个目录
	if err = os.RemoveAll(tempDir); err != nil {
		global.Logger.Error(t.Name()+" failed ["+status+"]: failed to remove temp directory", zap.String("path", tempDir), zap.Error(err))
		return err
	}

	// 重新创建目录
	if err = os.MkdirAll(tempDir, 0644); err != nil {
		global.Logger.Error(t.Name()+" failed ["+status+"]: failed to recreate temp directory", zap.String("path", tempDir), zap.Error(err))
		return err
	}

	if err != nil {
		global.Logger.Error(t.Name()+" failed ["+status+"]: ", zap.Error(err))
	} else {
		global.Logger.Info(t.Name() + " completed successfully [" + status + "]")
	}

	return nil
}

// NewTempFileCleanupTask 创建临时文件清理任务
func NewTempFileCleanupTask() (Task, error) {
	return &TempFileCleanupTask{
		firstRun: true,
	}, nil
}

func init() {
	Register(NewTempFileCleanupTask)
}
