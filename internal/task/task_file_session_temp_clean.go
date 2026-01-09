package task

import (
	"context"
	"os"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"go.uber.org/zap"
)

// FileSessionTempCleanTask 临时文件清理任务
type FileSessionTempCleanTask struct {
	firstRun bool
	app      *app.App
	logger   *zap.Logger
	tempPath string
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

	tempDir := t.tempPath
	if tempDir == "" {
		tempDir = "storage/temp"
	}

	var err error

	// 检查目录是否存在，不存在则创建并直接返回成功
	if _, err = os.Stat(tempDir); os.IsNotExist(err) {
		if err = os.MkdirAll(tempDir, 0754); err != nil {
			t.logger.Error("task log",
				zap.String("task", t.Name()),
				zap.String("path", tempDir),
				zap.Error(err))
			return err
		}
		return nil
	}

	// 删除整个目录
	if err = os.RemoveAll(tempDir); err != nil {
		t.logger.Error("task log",
			zap.String("task", t.Name()),
			zap.String("type", "startupRun"),
			zap.String("path", tempDir),
			zap.String("msg", "failed"),
			zap.Error(err))
		return err
	}

	// 重新创建目录
	if err = os.MkdirAll(tempDir, 0754); err != nil {
		t.logger.Error("task log",
			zap.String("task", t.Name()),
			zap.String("type", "startupRun"),
			zap.String("path", tempDir),
			zap.String("msg", "failed"),
			zap.Error(err))
		return err
	}

	t.logger.Info("task log",
		zap.String("task", t.Name()),
		zap.String("type", "startupRun"),
		zap.String("msg", "success"))

	return nil
}

// NewFileSessionTempCleanTask 创建临时文件清理任务
func NewFileSessionTempCleanTask(appContainer *app.App) (Task, error) {
	return &FileSessionTempCleanTask{
		firstRun: true,
		app:      appContainer,
		logger:   appContainer.Logger(),
		tempPath: appContainer.Config().App.TempPath,
	}, nil
}

func init() {
	RegisterWithApp(func(appContainer *app.App) (Task, error) {
		return NewFileSessionTempCleanTask(appContainer)
	})
}
