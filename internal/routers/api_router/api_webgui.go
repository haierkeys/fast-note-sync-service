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

type webGUI struct {
	FontSet          string `json:"fontSet"`
	RegisterIsEnable bool   `json:"registerIsEnable"`
}

func (w *WebGUI) Config(c *gin.Context) {
	response := app.NewResponse(c)
	data := webGUI{
		FontSet:          global.Config.WebGUI.FontSet,
		RegisterIsEnable: global.Config.User.RegisterIsEnable,
	}
	response.ToResponse(code.Success.WithData(data))
}
