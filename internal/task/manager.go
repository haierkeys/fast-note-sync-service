package task

import (
	"github.com/haierkeys/fast-note-sync-service/pkg/safe_close"
	"go.uber.org/zap"
)

// Manager 任务管理器,负责创建和管理所有任务
type Manager struct {
	scheduler *Scheduler
	logger    *zap.Logger
}

// NewManager 创建任务管理器
func NewManager(logger *zap.Logger, sc *safe_close.SafeClose) *Manager {
	return &Manager{
		scheduler: NewScheduler(logger, sc),
		logger:    logger,
	}
}

// RegisterTasks 注册所有任务
// 从全局注册表获取所有已注册的任务工厂,并创建任务实例
func (m *Manager) RegisterTasks() error {
	factories := GetFactories()

	if len(factories) == 0 {
		m.logger.Info("no task factories registered")
		return nil
	}

	m.logger.Info("tasks", zap.Int("count", len(factories)))

	for _, factory := range factories {
		task, err := factory()
		if err != nil {
			m.logger.Warn("failed to create task", zap.Error(err))
			continue
		}

		loopInterval := task.LoopInterval()
		isStartupRun := task.IsStartupRun()

		if isStartupRun && loopInterval <= 0 {
			m.logger.Info("task registered", zap.String("name", task.Name()), zap.Bool("once", true))
		} else if isStartupRun && loopInterval > 0 {
			m.logger.Info("task registered", zap.String("name", task.Name()), zap.Bool("once", true), zap.Duration("loop", loopInterval))
		} else if !isStartupRun && loopInterval > 0 {
			m.logger.Info("task registered", zap.String("name", task.Name()), zap.Bool("once", false), zap.Duration("loop", loopInterval))
		} else {
			m.logger.Info("task skipped", zap.String("name", task.Name()))
			continue
		}

		if task != nil {
			m.scheduler.AddTask(task)
		}
	}

	return nil
}

// Start 启动所有已注册的任务
func (m *Manager) Start() {
	m.scheduler.Start()
}
