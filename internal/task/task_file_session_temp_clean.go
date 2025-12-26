package task

import (
	"context"
	"os"
	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
	"go.uber.org/zap"
)

// FileSessionTempCleanTask 临时文件清理任务
type FileSessionTempCleanTask struct {
	firstRun bool
}

// Name 任务名称
func (t *FileSessionTempCleanTask) Name() string {
	return "FileSessionTempClean"
}

// LoopInterval 执行间隔 (0 表示不进行周期性执行)
func (t *FileSessionTempCleanTask) LoopInterval() time.Duration {
	return 0
}

// IsStartupRun 是否立即执行一次
func (t *FileSessionTempCleanTask) IsStartupRun() bool {
	return true
}

// Run 执行清理任务
func (t *FileSessionTempCleanTask) Run(ctx context.Context) error {
	t.firstRun = false

	tempDir := global.Config.App.TempPath
	if tempDir == "" {
		tempDir = "storage/temp"
	}

	var err error

	// 检查目录是否存在
	if _, err = os.Stat(tempDir); os.IsNotExist(err) {
		global.Logger.Error("task log",
			zap.String("task", t.Name()),
			zap.String("type", "startupRun"),
			zap.String("path", tempDir),
			zap.String("reason", "temp directory does not exist"),
			zap.String("msg", "failed"))
		return err
	}

	// 删除整个目录
	if err = os.RemoveAll(tempDir); err != nil {
		global.Logger.Error("task log",
			zap.String("task", t.Name()),
			zap.String("type", "startupRun"),
			zap.String("path", tempDir),
			zap.String("msg", "failed"),
			zap.Error(err))
		return err
	}

	// 重新创建目录
	if err = os.MkdirAll(tempDir, 0754); err != nil {
		global.Logger.Error("task log",
			zap.String("task", t.Name()),
			zap.String("type", "startupRun"),
			zap.String("path", tempDir),
			zap.String("msg", "failed"),
			zap.Error(err))
		return err
	}

	global.Logger.Info("task log",
		zap.String("task", t.Name()),
		zap.String("type", "startupRun"),
		zap.String("msg", "success"))

	return nil
}

// NewFileSessionTempCleanTask 创建临时文件清理任务
func NewFileSessionTempCleanTask() (Task, error) {
	return &FileSessionTempCleanTask{
		firstRun: true,
	}, nil
}

func init() {
	Register(NewFileSessionTempCleanTask)
}
