package middleware

import (
	"strings"

	"github.com/haierkeys/fast-note-sync-service/pkg/code"

	ut "github.com/go-playground/universal-translator"
	"github.com/gin-gonic/gin"
)

// LangWithTranslator 创建带翻译器的语言中间件（支持依赖注入）
func LangWithTranslator(uni *ut.UniversalTranslator) gin.HandlerFunc {

	return func(c *gin.Context) {

		var lang string

		if s, exist := c.GetQuery("lang"); exist {
			lang = s
		} else if s = c.GetHeader("lang"); len(s) != 0 {
			lang = s
		}

		lang = strings.ToLower(strings.ReplaceAll(lang, "-", "_"))

		trans, found := uni.GetTranslator(lang)

		if found {
			c.Set("trans", trans)
		} else {
			trans, _ := uni.GetTranslator("en")
			c.Set("trans", trans)
		}

		code.SetGlobalDefaultLang(lang)

		c.Next()
	}
}
