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
	return "DbSettingCleanupTask"
}

// Run 执行清理任务
func (t *SettingCleanupTask) Run(ctx context.Context) error {
	svc := service.NewBackground(ctx)

	status := "scheduled"
	if t.firstRun {
		status = "first-run"
		t.firstRun = false
	}

	err := svc.SettingCleanupAll()

	if err != nil {
		global.Logger.Error(t.Name()+" failed ["+status+"]: ", zap.Error(err))
	} else {
		global.Logger.Info(t.Name() + " completed successfully [" + status + "]")
	}

	return err
}

// Interval 返回执行间隔
func (t *SettingCleanupTask) Interval() time.Duration {
	return t.interval
}

// RunImmediately 是否立即执行一次
func (t *SettingCleanupTask) RunImmediately() bool {
	return true
}
