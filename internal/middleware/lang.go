package middleware

import (
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	"github.com/gin-gonic/gin"
)

func Lang() gin.HandlerFunc {

	return func(c *gin.Context) {

		var lang string

		if s, exist := c.GetQuery("lang"); exist {
			lang = s
		} else if s = c.GetHeader("lang"); len(s) != 0 {
			lang = s
		}

		trans, found := global.Ut.GetTranslator(lang)

		if found {
			c.Set("trans", trans)
		} else {
			trans, _ := global.Ut.GetTranslator("zh")
			c.Set("trans", trans)
		}

		code.SetGlobalDefaultLang(lang)

		c.Next()
	}
}
