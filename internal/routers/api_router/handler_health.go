// Package api_router 提供 HTTP API 路由处理器
package api_router

import (
	"time"

	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	*Handler
}

// NewHealthHandler 创建健康检查处理器实例
func NewHealthHandler(a *app.App) *HealthHandler {
	return &HealthHandler{Handler: NewHandler(a)}
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status   string  `json:"status"`   // "healthy" 或 "unhealthy"
	Version  string  `json:"version"`  // 服务版本号
	Uptime   float64 `json:"uptime"`   // 运行时间（秒）
	Database string  `json:"database"` // "connected" 或 "error"
}

// Check 健康检查接口
// @Summary 健康检查
// @Description 检查服务健康状态，包括数据库连接
// @Tags 系统
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /api/health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	response := HealthResponse{
		Status:   "healthy",
		Version:  h.App.Version().Version,
		Uptime:   time.Since(h.App.StartTime).Seconds(),
		Database: "connected",
	}

	// 检查数据库连接
	if err := h.App.DB.Raw("SELECT 1").Error; err != nil {
		response.Status = "unhealthy"
		response.Database = "error"
		pkgapp.NewResponse(c).ToResponse(code.Failed.WithData(response))
		return
	}

	pkgapp.NewResponse(c).ToResponse(code.Success.WithData(response))
}
