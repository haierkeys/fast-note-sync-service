package api_router

import (
	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/internal/app"
	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

// WebGUIHandler WebGUI 配置 API 路由处理器
// 使用 App Container 注入依赖
type WebGUIHandler struct {
	*Handler
}

// NewWebGUIHandler 创建 WebGUIHandler 实例
func NewWebGUIHandler(a *app.App) *WebGUIHandler {
	return &WebGUIHandler{
		Handler: NewHandler(a),
	}
}

// webGUIConfig WebGUI 配置响应结构
type webGUIConfig struct {
	FontSet          string `json:"fontSet"`
	RegisterIsEnable bool   `json:"registerIsEnable"`
}

// Config 获取 WebGUI 配置
func (h *WebGUIHandler) Config(c *gin.Context) {
	response := pkgapp.NewResponse(c)
	cfg := h.App.Config()
	data := webGUIConfig{
		FontSet:          cfg.WebGUI.FontSet,
		RegisterIsEnable: cfg.User.RegisterIsEnable,
	}
	response.ToResponse(code.Success.WithData(data))
}
