package api_router

import (
	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

type Version struct{}

// NewVersion 创建版本控制器实例
func NewVersion() *Version {
	return &Version{}
}

// ServerVersion 返回服务端版本信息
// @Summary 获取服务端版本信息
// @Description 返回服务端的版本号、Git标签和构建时间
// @Tags 系统信息
// @Accept json
// @Produce json
// @Success 200 {object} app.Response
// @Router /api/version [get]
func (v *Version) ServerVersion(c *gin.Context) {
	response := app.NewResponse(c)
	versionInfo := map[string]string{
		"version":   global.Version,
		"gitTag":    global.GitTag,
		"buildTime": global.BuildTime,
	}
	response.ToResponse(code.Success.WithData(versionInfo))
}
