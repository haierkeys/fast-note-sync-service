package task

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
)

// init 自动注册清理任务
func init() {
	Register(NewDbNoteCleanupTask)
}

// CleanupTask 清理任务
type CleanupTask struct {
	interval time.Duration
	firstRun bool
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
	return &CleanupTask{
		interval: 10 * time.Minute,
		firstRun: true,
	}, nil
}

// Name 返回任务名称
func (t *CleanupTask) Name() string {
	return "DbNoteCleanupTask"
}

// Run 执行清理任务
func (t *CleanupTask) Run(ctx context.Context) error {
	svc := service.NewBackground(ctx)

	status := "scheduled"
	if t.firstRun {
		status = "first-run"
		t.firstRun = false
	}

	err := svc.NoteCleanupAll()

	if err != nil {
		global.Logger.Error(t.Name()+" failed ["+status+"]: ", zap.Error(err))
	} else {
		global.Logger.Info(t.Name() + " completed successfully [" + status + "]")
	}

	return err
}

// LoopInterval 返回执行间隔
func (t *CleanupTask) LoopInterval() time.Duration {
	return t.interval
}

// IsStartupRun 是否立即执行一次
func (t *CleanupTask) IsStartupRun() bool {
	return true
}
