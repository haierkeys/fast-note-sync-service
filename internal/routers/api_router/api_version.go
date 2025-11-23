package api_router

import (
	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

// Version 版本信息 API 路由处理器
// 结构体名: Version
// 说明: 处理获取系统版本信息的 HTTP 请求。
type Version struct{}

// NewVersion 创建 Version 路由处理器实例
// 函数名: NewVersion
// 函数使用说明: 初始化并返回一个新的 Version 结构体实例。
// 返回值说明:
//   - *Version: 初始化后的 Version 实例
func NewVersion() *Version {
	return &Version{}
}

// ServerVersion 获取服务端版本信息
// 函数名: ServerVersion
// 函数使用说明: 处理获取服务端版本信息的 HTTP 请求。返回当前服务的版本号、Git 标签和构建时间。
// 参数说明:
//   - c *gin.Context: Gin 上下文
//
// 返回值说明:
//   - JSON: 包含版本信息的响应数据
func (v *Version) ServerVersion(c *gin.Context) {
	response := app.NewResponse(c)
	versionInfo := map[string]string{
		"version":   global.Version,
		"gitTag":    global.GitTag,
		"buildTime": global.BuildTime,
	}
	response.ToResponse(code.Success.WithData(versionInfo))
}
