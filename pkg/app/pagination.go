package app

import (
	"github.com/haierkeys/fast-note-sync-service/pkg/convert"

	"github.com/gin-gonic/gin"
)

// PaginationConfig pagination configuration // 分页配置
type PaginationConfig struct {
	DefaultPageSize int
	MaxPageSize     int
}

// DefaultPaginationConfig default pagination configuration // 默认分页配置
var DefaultPaginationConfig = PaginationConfig{
	DefaultPageSize: 10,
	MaxPageSize:     100,
}

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

// GetPageSizeWithConfig gets page size (using injected configuration)
// GetPageSizeWithConfig 获取分页大小（使用注入的配置）
func GetPageSizeWithConfig(c *gin.Context, cfg PaginationConfig) int {
	var pageSize int

	if s, exist := c.GetQuery("pageSize"); exist {
		pageSize = convert.StrTo(s).MustInt()
	} else if s := c.PostForm("pageSize"); s != "" {
		pageSize = convert.StrTo(s).MustInt()
	}

	if pageSize <= 0 {
		return cfg.DefaultPageSize
	}
	if pageSize > cfg.MaxPageSize {
		return cfg.MaxPageSize
	}

	return pageSize
}

// GetPageSize gets page size (using default configuration)
// GetPageSize 获取分页大小（使用默认配置）
func GetPageSize(c *gin.Context) int {
	return GetPageSizeWithConfig(c, DefaultPaginationConfig)
}

func GetPageOffset(page, pageSize int) int {
	result := 0
	if page > 0 {
		result = (page - 1) * pageSize
	}

	return result
}
