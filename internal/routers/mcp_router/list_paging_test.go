package mcp_router

import (
	"strings"
	"testing"
)

func TestMCPListPagerDefaultsAndBounds(t *testing.T) {
	cases := []struct {
		name       string
		args       map[string]interface{}
		page, size int
	}{
		{"defaults keep previous behavior", map[string]interface{}{}, 1, 100},
		{"explicit page and size", map[string]interface{}{"page": float64(3), "pageSize": float64(50)}, 3, 50},
		{"size capped", map[string]interface{}{"pageSize": float64(1000)}, 1, 100},
		{"invalid values fall back", map[string]interface{}{"page": float64(-2), "pageSize": float64(0)}, 1, 100},
	}
	for _, c := range cases {
		p := mcpListPager(c.args)
		if p.Page != c.page || p.PageSize != c.size {
			t.Errorf("%s: got page=%d size=%d, want %d/%d", c.name, p.Page, p.PageSize, c.page, c.size)
		}
	}
}

func TestMCPListHasMoreAndSummary(t *testing.T) {
	// 639 notes, 100 per page: pages 1-6 have more, page 7 (39 items) is the last.
	p := mcpListPager(map[string]interface{}{"page": float64(1)})
	if !mcpListHasMore(p, 100, 639) {
		t.Fatal("page 1 of 639 must report hasMore")
	}
	s := mcpListSummary("notes", p, 100, 639)
	if !strings.Contains(s, "of 639 total") || !strings.Contains(s, "page=2") {
		t.Fatalf("summary must state the total and the next page: %q", s)
	}
	last := mcpListPager(map[string]interface{}{"page": float64(7)})
	if mcpListHasMore(last, 39, 639) {
		t.Fatal("last page must not report hasMore")
	}
	if s := mcpListSummary("notes", last, 39, 639); !strings.Contains(s, "items 601-639 of 639") || strings.Contains(s, "More results") {
		t.Fatalf("unexpected last-page summary: %q", s)
	}
	// Exactly 100 items: a single full page is complete.
	if mcpListHasMore(p, 100, 100) {
		t.Fatal("a full single page must not report hasMore")
	}
}
