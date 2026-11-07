package dao

import (
	"strings"
	"testing"
)

// TestExtractTags covers the tag grammar the index relies on. Every case here
// is a rule Obsidian users depend on, so a change in behaviour must be
// deliberate rather than incidental.
// TestExtractTags 覆盖索引依赖的标签语法。这里的每个用例都是 Obsidian 用户依赖的规则，
// 行为变化必须是刻意的而不是顺带的。
func TestExtractTags(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantTags   []string
		wantSource string
	}{
		{
			name:       "frontmatter list",
			content:    "---\ntags: [golang, sqlite]\n---\nbody",
			wantTags:   []string{"golang", "sqlite"},
			wantSource: TagSourceFrontmatter,
		},
		{
			name:       "frontmatter scalar tag field",
			content:    "---\ntag: project/xhs\n---\nbody",
			wantTags:   []string{"project/xhs"},
			wantSource: TagSourceFrontmatter,
		},
		{
			name:       "frontmatter block list",
			content:    "---\ntags:\n  - alpha\n  - beta\n---\nbody",
			wantTags:   []string{"alpha", "beta"},
			wantSource: TagSourceFrontmatter,
		},
		{
			name:       "inline hashtags",
			content:    "记一下 #golang 和 #爬虫 的笔记",
			wantTags:   []string{"golang", "爬虫"},
			wantSource: TagSourceInline,
		},
		{
			name:       "inline nested tag",
			content:    "see #project/xhs for details",
			wantTags:   []string{"project/xhs"},
			wantSource: TagSourceInline,
		},
		{
			name:       "both sources",
			content:    "---\ntags: [meta]\n---\nbody with #inline",
			wantTags:   []string{"meta", "inline"},
			wantSource: TagSourceBoth,
		},
		{
			name:       "code fence is ignored",
			content:    "text\n```\n#notatag\n```\nafter #real",
			wantTags:   []string{"real"},
			wantSource: TagSourceInline,
		},
		{
			name:       "inline code is ignored",
			content:    "use `#notatag` here but #real counts",
			wantTags:   []string{"real"},
			wantSource: TagSourceInline,
		},
		{
			name:       "heading is not a tag",
			content:    "## 标题不是标签\n正文 #real",
			wantTags:   []string{"real"},
			wantSource: TagSourceInline,
		},
		{
			name:       "url fragment is not a tag",
			content:    "https://example.com/page#section end",
			wantTags:   nil,
			wantSource: "",
		},
		{
			name:       "duplicate tags collapse case-insensitively",
			content:    "#Golang and #golang once",
			wantTags:   []string{"Golang"},
			wantSource: TagSourceInline,
		},
		{
			name:       "numeric-only is not a tag",
			content:    "issue #123 closed",
			wantTags:   nil,
			wantSource: "",
		},
		{
			name:       "empty content",
			content:    "",
			wantTags:   nil,
			wantSource: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags, source := ExtractTags(tt.content)
			if source != tt.wantSource {
				t.Fatalf("source = %q, want %q", source, tt.wantSource)
			}
			if len(tags) != len(tt.wantTags) {
				t.Fatalf("tags = %v, want %v", tags, tt.wantTags)
			}
			for i := range tags {
				if tags[i] != tt.wantTags[i] {
					t.Fatalf("tags = %v, want %v", tags, tt.wantTags)
				}
			}
		})
	}
}

// TestBuildSnippet pins the snippet contract: the matched line is present, the
// match is centred roughly in the middle, and the result never returns the
// whole body for a long note.
// TestBuildSnippet 固定片段契约：命中行必须出现、命中大致居中，长笔记不会返回全文。
func TestBuildSnippet(t *testing.T) {
	var lines []string
	for i := 0; i < 200; i++ {
		lines = append(lines, "filler line with nothing special")
	}
	lines[100] = "here is the KEYWORD we look for"
	content := strings.Join(lines, "\n")

	got := BuildSnippet(content, "keyword")
	if !strings.Contains(got, "KEYWORD") {
		t.Fatalf("snippet lost the match: %q", got)
	}
	if len(got) > 600 {
		t.Fatalf("snippet too large: %d chars", len(got))
	}
	if !strings.HasPrefix(got, "…") || !strings.HasSuffix(got, "…") {
		t.Fatalf("snippet should mark truncation on both ends: %q", got)
	}

	// A short note is returned whole, with no truncation markers.
	short := "one line\nwith keyword inside"
	if s := BuildSnippet(short, "keyword"); s != short {
		t.Fatalf("short note should be returned verbatim, got %q", s)
	}

	// No match -> first lines, still truncated for long content.
	if s := BuildSnippet(content, "absent-term"); s == "" {
		t.Fatal("no-match snippet should fall back to the head of the note")
	}
}
