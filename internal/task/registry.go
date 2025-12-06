package task

import "sync"

// TaskFactory 任务工厂函数类型,用于创建任务实例
type TaskFactory func() (Task, error)

// taskRegistry 全局任务注册表
var (
	taskRegistry  []TaskFactory
	registryMutex sync.RWMutex
)

// Register 注册任务工厂函数
// 通常在各个任务文件的 init() 函数中调用
func Register(factory TaskFactory) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	taskRegistry = append(taskRegistry, factory)
}

// GetFactories 获取所有已注册的任务工厂
func GetFactories() []TaskFactory {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	// 返回副本,避免外部修改
	factories := make([]TaskFactory, len(taskRegistry))
	copy(factories, taskRegistry)
	return factories
}
