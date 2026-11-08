package mcp_router

import (
	"fmt"

	pkgapp "github.com/haierkeys/fast-note-sync-service/pkg/app"
	"github.com/mark3labs/mcp-go/mcp"
)

// mcpListMaxPageSize is the largest page note_list/file_list return in one call.
// mcpListMaxPageSize 为 note_list/file_list 单次返回的最大条数。
const mcpListMaxPageSize = 100

// mcpListPageOption and mcpListPageSizeOption declare the paging arguments shared by list tools.
// mcpListPageOption 与 mcpListPageSizeOption 声明列表类工具共用的分页参数。
func mcpListPageOption() mcp.ToolOption {
	return mcp.WithNumber("page", mcp.Description("1-based page number. Default 1. Use hasMore and totalCount in the result to page through all items."))
}

func mcpListPageSizeOption() mcp.ToolOption {
	return mcp.WithNumber("pageSize", mcp.Description(fmt.Sprintf("Items per page, 1-%d. Default %d.", mcpListMaxPageSize, mcpListMaxPageSize)))
}

// mcpListPager builds a pager from tool arguments; missing or invalid values fall back to page 1
// and the maximum page size, matching the previous fixed behavior.
// mcpListPager 根据工具参数构造分页；缺省或非法值回退到第 1 页与最大页大小（与原固定行为一致）。
func mcpListPager(args map[string]interface{}) *pkgapp.Pager {
	page := int(getInt64Arg(args, "page"))
	pageSize := int(getInt64Arg(args, "pageSize"))
	if pageSize <= 0 {
		pageSize = mcpListMaxPageSize
	}
	return &pkgapp.Pager{
		Page:     pkgapp.GetPage(page),
		PageSize: pkgapp.GetPageSize(pageSize),
	}
}

// mcpListHasMore reports whether items remain after the returned page.
// mcpListHasMore 判断当前页之后是否还有条目。
func mcpListHasMore(pager *pkgapp.Pager, returned, total int) bool {
	return (pager.Page-1)*pager.PageSize+returned < total
}

// mcpListSummary is the first text line of a list result; it states the total and how to continue.
// mcpListSummary 为列表结果的首行文字，说明总数及如何继续翻页。
func mcpListSummary(kind string, pager *pkgapp.Pager, returned, total int) string {
	if returned == 0 {
		return fmt.Sprintf("Found 0 %s on page %d (%d total).\n", kind, pager.Page, total)
	}
	first := (pager.Page-1)*pager.PageSize + 1
	s := fmt.Sprintf("Found %d %s (items %d-%d of %d total, page %d):\n", returned, kind, first, first+returned-1, total, pager.Page)
	if mcpListHasMore(pager, returned, total) {
		s = fmt.Sprintf("%sMore results available: call again with page=%d.\n", s, pager.Page+1)
	}
	return s
}
