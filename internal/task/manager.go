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
func (m *Manager) RegisterTasks() error {
	// 创建并添加清理任务
	cleanupTask, err := NewCleanupTask()
	if err != nil {
		m.logger.Warn("failed to create cleanup task", zap.Error(err))
		return err
	}

	if cleanupTask != nil {
		m.scheduler.AddTask(cleanupTask)
	} else {
		m.logger.Info("cleanup task is disabled (retention time not configured)")
	}

	// 未来可以在这里添加更多任务
	// otherTask := NewOtherTask()
	// m.scheduler.AddTask(otherTask)

	return nil
}

// Start 启动所有已注册的任务
func (m *Manager) Start() {
	m.scheduler.Start()
}
