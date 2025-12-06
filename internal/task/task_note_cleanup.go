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
	Register(NewCleanupTask)
}

// CleanupTask 清理任务
type CleanupTask struct {
	interval time.Duration
}

// NewCleanupTask 创建清理任务
func NewCleanupTask() (Task, error) {
	retentionTimeStr := global.Config.App.DeleteNoteRetentionTime
	duration, err := util.ParseDuration(retentionTimeStr)
	if err != nil {
		return nil, err
	}

	if duration <= 0 {
		return nil, nil
	}

	// 每分钟执行一次检查
	return &CleanupTask{
		interval: 10 * time.Minute,
	}, nil
}

// Name 返回任务名称
func (t *CleanupTask) Name() string {
	return "CleanupTask"
}

// Run 执行清理任务
func (t *CleanupTask) Run(ctx context.Context) error {
	svc := service.NewBackground(ctx)
	return svc.NoteCleanupAll()
}

// Interval 返回执行间隔
func (t *CleanupTask) Interval() time.Duration {
	return t.interval
}

// RunImmediately 是否立即执行一次
func (t *CleanupTask) RunImmediately() bool {
	return true
}
