package middleware

import (
	"strings"

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

		lang = strings.ToLower(strings.ReplaceAll(lang, "-", "_"))

		trans, found := global.Ut.GetTranslator(lang)

		if found {
			c.Set("trans", trans)
		} else {
			trans, _ := global.Ut.GetTranslator("en")
			c.Set("trans", trans)
		}

		code.SetGlobalDefaultLang(lang)

		c.Next()
	}
}
