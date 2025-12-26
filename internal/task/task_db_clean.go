package task

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/service"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
)

// DbCleanTask 清理任务
type DbCleanTask struct {
}

// Name 返回任务名称
func (t *DbCleanTask) Name() string {
	return "DbCleanup"
}

// LoopInterval 返回执行间隔
func (t *DbCleanTask) LoopInterval() time.Duration {
	return 10 * time.Minute
}

// IsStartupRun 是否立即执行一次
func (t *DbCleanTask) IsStartupRun() bool {
	return true
}

// Run 执行清理任务
func (t *DbCleanTask) Run(ctx context.Context) error {
	svc := service.NewBackground(ctx)

	// 执行所有清理任务
	var errs []error

	if err := svc.NoteCleanupAll(); err != nil {
		errs = append(errs, err)
		global.Logger.Error("task log",
			zap.String("task", t.Name()),
			zap.String("sub_task", "NoteCleanup"),
			zap.String("msg", "failed"),
			zap.Error(err))
	}

	if err := svc.FileCleanupAll(); err != nil {
		errs = append(errs, err)
		global.Logger.Error("task log",
			zap.String("task", t.Name()),
			zap.String("sub_task", "FileCleanup"),
			zap.String("msg", "failed"),
			zap.Error(err))
	}

	if err := svc.SettingCleanupAll(); err != nil {
		errs = append(errs, err)
		global.Logger.Error("task log",
			zap.String("task", t.Name()),
			zap.String("sub_task", "SettingCleanup"),
			zap.String("msg", "failed"),
			zap.Error(err))
	}

	if len(errs) > 0 {
		return errs[0] // 返回第一个错误
	}

	global.Logger.Info("task log",
		zap.String("task", t.Name()),
		zap.String("type", "loopRun"),
		zap.String("msg", "success"))

	return nil
}

// NewDbNoteDbCleanTask 创建清理任务
func NewDbCleanTask() (Task, error) {
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
	return &DbCleanTask{}, nil
}

// init 自动注册清理任务
func init() {
	Register(NewDbCleanTask)
}
