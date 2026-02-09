package task

import (
	"context"
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/util"
	"go.uber.org/zap"
)

// DbCleanTask 清理任务
type DbCleanTask struct {
	app                 *app.App
	logger              *zap.Logger
	retentionDuration   time.Duration
	historyKeepVersions int
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
	// 计算截止时间
	cutoffTime := time.Now().Add(-t.retentionDuration).UnixMilli()

	var errs []error

	// 调用各 Service 的 CleanupByTime 方法
	if err := t.app.NoteService.CleanupByTime(ctx, cutoffTime); err != nil {
		errs = append(errs, err)
		t.logger.Error("cleanup failed",
			zap.String("task", t.Name()),
			zap.String("service", "NoteService"),
			zap.Error(err))
	} else {
		t.logger.Info("cleanup success",
			zap.String("task", t.Name()),
			zap.String("service", "NoteService"))
	}

	if err := t.app.FileService.CleanupByTime(ctx, cutoffTime); err != nil {
		errs = append(errs, err)
		t.logger.Error("cleanup failed",
			zap.String("task", t.Name()),
			zap.String("service", "FileService"),
			zap.Error(err))
	} else {
		t.logger.Info("cleanup success",
			zap.String("task", t.Name()),
			zap.String("service", "FileService"))
	}

	if err := t.app.SettingService.CleanupByTime(ctx, cutoffTime); err != nil {
		errs = append(errs, err)
		t.logger.Error("cleanup failed",
			zap.String("task", t.Name()),
			zap.String("service", "SettingService"),
			zap.Error(err))
	} else {
		t.logger.Info("cleanup success",
			zap.String("task", t.Name()),
			zap.String("service", "SettingService"))
	}

	// 清理 NoteHistory
	if err := t.app.NoteHistoryService.CleanupByTime(ctx, cutoffTime, t.historyKeepVersions); err != nil {
		errs = append(errs, err)
		t.logger.Error("cleanup failed",
			zap.String("task", t.Name()),
			zap.String("service", "NoteHistoryService"),
			zap.Error(err))
	} else {
		t.logger.Info("cleanup success",
			zap.String("task", t.Name()),
			zap.String("service", "NoteHistoryService"))
	}

	// 清理闲置数据库连接 (保持 1 小时闲置)
	t.app.Dao.CleanupConnections(time.Hour)

	if len(errs) > 0 {
		return errs[0] // 返回第一个错误
	}

	return nil
}

// NewDbCleanTask 创建清理任务
func NewDbCleanTask(appContainer *app.App) (Task, error) {
	retentionTimeStr := appContainer.Config().App.SoftDeleteRetentionTime
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

	// 获取历史记录保留版本数，默认 10
	historyKeepVersions := appContainer.Config().App.HistoryKeepVersions
	if historyKeepVersions <= 0 {
		historyKeepVersions = 10
	}

	return &DbCleanTask{
		app:                 appContainer,
		logger:              appContainer.Logger(),
		retentionDuration:   duration,
		historyKeepVersions: historyKeepVersions,
	}, nil
}

// init 自动注册清理任务
func init() {
	RegisterWithApp(func(appContainer *app.App) (Task, error) {
		return NewDbCleanTask(appContainer)
	})
}
