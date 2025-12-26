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
	Register(NewSettingCleanupTask)
}

// SettingCleanupTask 配置清理任务
type SettingCleanupTask struct {
	interval time.Duration
	firstRun bool
}

// NewSettingCleanupTask 创建配置清理任务
func NewSettingCleanupTask() (Task, error) {
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
	return &SettingCleanupTask{
		interval: 10 * time.Minute,
		firstRun: true,
	}, nil
}

// Name 返回任务名称
func (t *SettingCleanupTask) Name() string {
	return "DbSettingCleanup"
}

// Run 执行清理任务
func (t *SettingCleanupTask) Run(ctx context.Context) error {
	svc := service.NewBackground(ctx)
	err := svc.SettingCleanupAll()

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

// LoopInterval 返回执行间隔
func (t *SettingCleanupTask) LoopInterval() time.Duration {
	return t.interval
}

// IsStartupRun 是否立即执行一次
func (t *SettingCleanupTask) IsStartupRun() bool {
	return true
}
