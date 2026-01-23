// Package app 提供应用容器，封装所有依赖和服务
package app

// 版本信息变量，由构建时注入
var (
	Version   string = "1.10.5"
	GitTag    string = "2000.01.01.release"
	BuildTime string = "2000-01-01T00:00:00+0800"
)

// 应用名称常量
const (
	// Name 应用名称
	Name = "Fast Note Sync Service"
)
