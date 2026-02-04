package app

import (
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

func GetPage(page int) int {
	if page <= 0 {
		return 1
	}

	return page
}

// GetPageSize gets page size (using default configuration)
// GetPageSize 获取分页大小（使用默认配置）
func GetPageSize(pageSize int) int {

	if pageSize <= 0 {
		return DefaultPaginationConfig.DefaultPageSize
	}
	if pageSize > DefaultPaginationConfig.MaxPageSize {
		return DefaultPaginationConfig.MaxPageSize
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

// NewPager creates a new Pager instance from gin.Context
// NewPager 从 gin.Context 创建一个新的 Pager 实例
func NewPager(c *gin.Context, count ...int) *Pager {

	params := &PaginationRequest{}
	if valid, errs := BindAndValid(c, params); !valid {
		log(LogError, errs.Error())
	}

	var totalRows int
	if len(count) > 0 {
		totalRows = count[0]
	}

	return &Pager{
		Page:      GetPage(params.Page),
		PageSize:  GetPageSize(params.PageSize),
		TotalRows: totalRows,
	}
}
