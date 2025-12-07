package task

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
)

// init 自动注册清理任务
func init() {
	Register(NewFileCleanupTask)
}

// FileCleanupTask 文件清理任务
type FileCleanupTask struct {
	interval time.Duration
}

// NewFileCleanupTask 创建文件清理任务
func NewFileCleanupTask() (Task, error) {
	retentionTimeStr := global.Config.App.SoftDeleteRetentionTime
	if retentionTimeStr == "" {
		return nil, nil
	}
	duration, err := util.ParseDuration(retentionTimeStr)
	if err != nil {
		return nil, err
	}

	if duration <= 0 {
		return nil, nil
	}

	// 每10分钟执行一次检查
	return &FileCleanupTask{
		interval: 10 * time.Minute,
	}, nil
}

// Name 返回任务名称
func (t *FileCleanupTask) Name() string {
	return "FileCleanupTask"
}

// Run 执行清理任务
func (t *FileCleanupTask) Run(ctx context.Context) error {
	svc := service.NewBackground(ctx)
	return svc.FileCleanupAll()
}

// Interval 返回执行间隔
func (t *FileCleanupTask) Interval() time.Duration {
	return t.interval
}

// RunImmediately 是否立即执行一次
func (t *FileCleanupTask) RunImmediately() bool {
	return true
}
