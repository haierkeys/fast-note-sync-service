package task

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
)

// CleanupTask 清理任务
type CleanupTask struct {
}

// Name 返回任务名称
func (t *CleanupTask) Name() string {
	return "DbNoteCleanup"
}

// LoopInterval 返回执行间隔
func (t *CleanupTask) LoopInterval() time.Duration {
	return 10 * time.Minute
}

// IsStartupRun 是否立即执行一次
func (t *CleanupTask) IsStartupRun() bool {
	return true
}

// Run 执行清理任务
func (t *CleanupTask) Run(ctx context.Context) error {
	svc := service.NewBackground(ctx)
	err := svc.NoteCleanupAll()

	if err != nil {
		global.Logger.Error("task log",
			zap.String("task", t.Name()),
			zap.String("type", "loopRun"),
			zap.String("msg", "failed"),
			zap.Error(err))
	} else {
		global.Logger.Info("task log",
			zap.String("task", t.Name()),
			zap.String("type", "loopRun"),
			zap.String("msg", "success"))
	}

	return err
}

// NewDbNoteCleanupTask 创建清理任务
func NewDbNoteCleanupTask() (Task, error) {
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

	// 每分钟执行一次检查
	return &CleanupTask{}, nil
}

// init 自动注册清理任务
func init() {
	Register(NewDbNoteCleanupTask)
}
