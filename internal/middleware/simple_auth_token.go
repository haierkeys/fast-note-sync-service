/**
  @author: haierkeys
  @since: 2022/9/14
  @desc:
**/

package middleware

import (
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
)

func SimpleAuthToken() gin.HandlerFunc {
	return func(c *gin.Context) {

		if global.Config.Security.AuthToken == "" {
			c.Next()
		}

		response := app.NewResponse(c)

		var token string

		if s, exist := c.GetQuery("authorization"); exist {
			token = s
		} else if s, exist = c.GetQuery("Authorization"); exist {
			token = s
		} else if s = c.GetHeader("authorization"); len(s) != 0 {
			token = s
		} else if s = c.GetHeader("Authorization"); len(s) != 0 {
			token = s
		}

		if token != global.Config.Security.AuthToken {
			response.ToResponse(code.ErrorInvalidAuthToken)
			c.Abort()
		}
		c.Next()
	}
}
