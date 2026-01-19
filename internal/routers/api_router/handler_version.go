package api_router

import (
	"github.com/gin-gonic/gin"
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
// @Summary 获取服务端版本信息
// @Description 获取服务端当前的软件版本、Git 标签和构建时间
// @Tags 系统
// @Produce json
// @Success 200 {object} pkgapp.Res{data=map[string]string} "成功"
// @Router /api/version [get]
func (h *VersionHandler) ServerVersion(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	versionInfo := h.App.Version()
	response.ToResponse(code.Success.WithData(map[string]string{
		"version":   versionInfo.Version,
		"gitTag":    versionInfo.GitTag,
		"buildTime": versionInfo.BuildTime,
	}))
}
