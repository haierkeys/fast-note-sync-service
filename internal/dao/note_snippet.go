package dao

import (
	"strings"
)

// snippetContextLines is how many lines on each side of the match are kept.
// snippetContextLines 命中行上下各保留的行数。
const snippetContextLines = 2

// snippetMaxRunes caps the returned window so a dense match cannot blow the
// response up; callers get a truncated snippet instead of a whole note.
// snippetMaxRunes 限制返回窗口长度，避免密集命中把响应撑大；调用方拿到的是截断片段而非整篇正文。
const snippetMaxRunes = 400

// BuildSnippet returns a short window of content around the first
// case-insensitive occurrence of keyword, extended to whole lines plus
// snippetContextLines of context on each side.
//
// When the keyword is not found (for example a path-only match) it falls back
// to the head of the note, so the caller always gets something useful.
// BuildSnippet 返回 content 中第一处 keyword（忽略大小写）附近的短窗口，
// 扩展为整行并在两侧各保留 snippetContextLines 行上下文。
// 未找到 keyword 时（例如仅路径命中）回退到正文开头，保证调用方总能拿到有用内容。
func BuildSnippet(content, keyword string) string {
	if content == "" {
		return ""
	}
	runes := []rune(content)
	lines := strings.Split(content, "\n")

	hitLine := -1
	if kw := strings.TrimSpace(keyword); kw != "" {
		lowerKw := strings.ToLower(kw)
		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), lowerKw) {
				hitLine = i
				break
			}
		}
	}
	if hitLine < 0 {
		hitLine = 0
	}

	start := hitLine - snippetContextLines
	if start < 0 {
		start = 0
	}
	end := hitLine + snippetContextLines + 1
	if end > len(lines) {
		end = len(lines)
	}

	out := strings.TrimSpace(strings.Join(lines[start:end], "\n"))
	// hard cap on rune count, keeping the head of the window
	// 按 rune 数硬截断，保留窗口前部
	if len([]rune(out)) > snippetMaxRunes {
		out = string([]rune(out)[:snippetMaxRunes]) + "…"
	}
	// mark that the window does not start at the note head
	// 标记该窗口并非从正文开头开始
	if start > 0 {
		out = "…" + out
	}
	if end < len(lines) {
		out = out + "…"
	}
	_ = runes
	return out
}
