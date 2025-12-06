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

	m.logger.Info("registering tasks from registry", zap.Int("factory_count", len(factories)))

	for _, factory := range factories {
		task, err := factory()
		if err != nil {
			m.logger.Warn("failed to create task", zap.Error(err))
			continue
		}

		if task != nil {
			m.scheduler.AddTask(task)
			m.logger.Info("task registered successfully", zap.String("task_name", task.Name()))
		}
	}

	return nil
}

// Start 启动所有已注册的任务
func (m *Manager) Start() {
	m.scheduler.Start()
}
