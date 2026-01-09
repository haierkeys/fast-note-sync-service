package api_router

import (
	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

// VersionHandler 版本信息 API 路由处理器
// 使用 App Container 注入依赖
type VersionHandler struct {
	*Handler
}

// NewVersionHandler 创建 VersionHandler 实例
func NewVersionHandler(a *app.App) *VersionHandler {
	return &VersionHandler{
		Handler: NewHandler(a),
	}
}

// ServerVersion 获取服务端版本信息
func (h *VersionHandler) ServerVersion(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	versionInfo := map[string]string{
		"version":   global.Version,
		"gitTag":    global.GitTag,
		"buildTime": global.BuildTime,
	}
	response.ToResponse(code.Success.WithData(versionInfo))
}
