package api_router

import (
	"github.com/gin-gonic/gin"
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"
)

type WebGUI struct{}

func NewWebGUI() *WebGUI {
	return &WebGUI{}
}

func (w *WebGUI) Config(c *gin.Context) {
	response := app.NewResponse(c)
	response.ToResponse(code.Success.WithData(global.Config.WebGUI))
}
