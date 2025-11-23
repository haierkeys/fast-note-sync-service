package app

import (
	"github.com/haierkeys/fast-note-sync-service/global"
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"

	"github.com/gin-gonic/gin"
)

func GetPage(c *gin.Context) int {

	var page int

	if s, exist := c.GetQuery("page"); exist {
		page = convert.StrTo(s).MustInt()
	} else if s := c.PostForm("page"); s != "" {
		page = convert.StrTo(s).MustInt()
	}

	if page <= 0 {
		return 1
	}

	return page
}

func GetPageSize(c *gin.Context) int {

	var pageSize int

	if s, exist := c.GetQuery("pageSize"); exist {
		pageSize = convert.StrTo(s).MustInt()
	} else if s := c.PostForm("pageSize"); s != "" {
		pageSize = convert.StrTo(s).MustInt()
	}

	if pageSize <= 0 {
		return global.Config.App.DefaultPageSize
	}
	if pageSize > global.Config.App.MaxPageSize {
		return global.Config.App.MaxPageSize
	}

	return pageSize
}

func GetPageOffset(page, pageSize int) int {
	result := 0
	if page > 0 {
		result = (page - 1) * pageSize
	}

	return result
}
