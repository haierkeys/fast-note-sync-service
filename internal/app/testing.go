// Package app provides application container, encapsulates all dependencies and services
// Package app 提供应用容器，封装所有依赖和服务
package app

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NewTestApp creates a minimal App instance for unit testing.
// NewTestApp 创建用于单元测试的最小 App 实例。
// Only the Services field and a nop logger are initialized;
// 仅初始化 Services 字段和 nop logger；
// all other infrastructure fields remain zero/nil.
// 所有其他基础设施字段保持零值/nil。
func NewTestApp(svcs *Services, dbs ...*gorm.DB) *App {
	var db *gorm.DB
	if len(dbs) > 0 {
		db = dbs[0]
	}

	return &App{
		Infra: &Infra{
			logger: zap.NewNop(), // safe nop logger, prevents Logger() panic // 安全的 nop logger，防止 Logger() panic
			config: &AppConfig{},
			DB:     db,
		},
		Services: svcs,
	}
}
