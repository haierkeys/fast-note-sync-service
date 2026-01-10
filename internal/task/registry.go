package task

import (
	"sync"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
)

// TaskFactory 任务工厂函数类型,用于创建任务实例
type TaskFactory func() (Task, error)

// TaskFactoryWithApp 带 App Container 的任务工厂函数类型
type TaskFactoryWithApp func(appContainer *app.App) (Task, error)

// taskRegistry 全局任务注册表
var (
	taskRegistry        []TaskFactory
	taskRegistryWithApp []TaskFactoryWithApp
	registryMutex       sync.RWMutex
)

// Register 注册任务工厂函数
// 通常在各个任务文件的 init() 函数中调用
func Register(factory TaskFactory) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	taskRegistry = append(taskRegistry, factory)
}

// RegisterWithApp 注册带 App Container 的任务工厂函数
func RegisterWithApp(factory TaskFactoryWithApp) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	taskRegistryWithApp = append(taskRegistryWithApp, factory)
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

// GetFactoriesWithApp 获取所有已注册的带 App Container 的任务工厂
func GetFactoriesWithApp() []TaskFactoryWithApp {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	// 返回副本,避免外部修改
	factories := make([]TaskFactoryWithApp, len(taskRegistryWithApp))
	copy(factories, taskRegistryWithApp)
	return factories
}
