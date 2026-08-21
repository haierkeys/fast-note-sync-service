package middleware

import (
	"net/http"

	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
)

// NoFound 404 handler
// NoFound 404 处理
func NoFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		codeObj := code.ErrorNotFoundAPI
		c.Set("status_code", http.StatusNotFound)
		c.AbortWithStatusJSON(http.StatusNotFound, app.Res{
			Code:    codeObj.Code(),
			Status:  codeObj.Status(),
			Message: codeObj.MsgIn(c.GetString("lang")),
			Data:    codeObj.Data(),
		})
	}
}
