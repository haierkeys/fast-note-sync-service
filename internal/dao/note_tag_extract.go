package dao

import (
	"regexp"
	"strings"
	"unicode"
)

// Source values recorded on note_tag.source.
// note_tag.source 的取值。
const (
	TagSourceFrontmatter = "frontmatter"
	TagSourceInline      = "inline"
	TagSourceBoth        = "both"
)

var (
	// inlineTagRe matches #tag and #nested/tag. It must not be preceded by a
	// word character or another '#', which rules out headings ("# Title" has a
	// space after '#'), URL fragments ("x#y" is preceded by a word char) and
	// "##" heading markers.
	inlineTagRe = regexp.MustCompile(`(?:^|[\s(\[{,;])#([\p{L}\p{N}_][\p{L}\p{N}_/-]*)`)

	// fenceRe matches an opening or closing code fence.
	fenceRe = regexp.MustCompile("^\\s*(```|~~~)")

	// inlineCodeRe matches a single-backtick code span.
	inlineCodeRe = regexp.MustCompile("`[^`]*`")

	// frontmatterBlockRe captures the leading --- ... --- block.
	frontmatterBlockRe = regexp.MustCompile(`(?s)\A---\r?\n(.*?)\r?\n---`)

	// frontmatterTagLineRe matches "tags:" / "tag:" at the start of a line.
	frontmatterTagLineRe = regexp.MustCompile(`(?im)^(tags|tag)\s*:\s*(.*)$`)
)

// ExtractTags returns the tags found in a note body plus which source they came
// from (frontmatter, inline, or both).
// ExtractTags 返回笔记正文中的标签，以及标签来源（frontmatter、行内、两者都有）。
func ExtractTags(content string) ([]string, string) {
	if content == "" {
		return nil, ""
	}

	var fromFrontmatter []string
	body := content
	if m := frontmatterBlockRe.FindStringSubmatch(content); m != nil {
		fromFrontmatter = parseFrontmatterTags(m[1])
		body = content[len(m[0]):]
	}

	fromInline := scanInlineTags(body)

	switch {
	case len(fromFrontmatter) > 0 && len(fromInline) > 0:
		return dedupeTags(append(fromFrontmatter, fromInline...)), TagSourceBoth
	case len(fromFrontmatter) > 0:
		return dedupeTags(fromFrontmatter), TagSourceFrontmatter
	case len(fromInline) > 0:
		return dedupeTags(fromInline), TagSourceInline
	}
	return nil, ""
}

// parseFrontmatterTags understands both YAML list form
//
//	tags:
//	  - a
//	  - b
//
// and inline form
//
//	tags: [a, b]
//	tags: a, b
func parseFrontmatterTags(block string) []string {
	var out []string
	lines := strings.Split(block, "\n")
	inList := false
	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)

		if m := frontmatterTagLineRe.FindStringSubmatch(line); m != nil {
			value := strings.TrimSpace(m[2])
			inList = value == ""
			value = strings.TrimPrefix(value, "[")
			value = strings.TrimSuffix(value, "]")
			out = append(out, splitTagValue(value)...)
			continue
		}
		if inList {
			if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "-\t") {
				out = append(out, splitTagValue(strings.TrimSpace(trimmed[2:]))...)
				continue
			}
			if trimmed == "" {
				continue
			}
			inList = false
		}
	}
	return out
}

func splitTagValue(value string) []string {
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '"' || r == '\''
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(strings.TrimPrefix(f, "#"))
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

// scanInlineTags walks the body line by line so fenced code blocks can be
// skipped, then strips inline code spans before matching.
// scanInlineTags 逐行扫描正文以跳过代码块，再去掉行内代码后匹配标签。
func scanInlineTags(body string) []string {
	var out []string
	inFence := false
	var fenceMarker string

	for _, raw := range strings.Split(body, "\n") {
		line := strings.TrimRight(raw, "\r")
		if m := fenceRe.FindStringSubmatch(line); m != nil {
			if !inFence {
				inFence = true
				fenceMarker = m[1]
			} else if m[1] == fenceMarker {
				inFence = false
				fenceMarker = ""
			}
			continue
		}
		if inFence {
			continue
		}
		line = inlineCodeRe.ReplaceAllString(line, " ")
		for _, m := range inlineTagRe.FindAllStringSubmatch(line, -1) {
			if len(m) <= 1 || m[1] == "" {
				continue
			}
			// Obsidian does not treat a purely numeric body as a tag (#123 is
			// an issue reference), but a numeric-led nested tag such as
			// #2024/report is valid, so require a letter anywhere in the body.
			// Obsidian 不把纯数字当标签（#123 是 issue 编号），但 #2024/report
			// 这类数字开头的嵌套标签有效，因此要求标签体内至少有一个字母。
			if !hasTagLetter(m[1]) {
				continue
			}
			out = append(out, m[1])
		}
	}
	return out
}

func dedupeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(t), "#"))
		if t == "" {
			continue
		}
		key := strings.ToLower(t)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, t)
	}
	return out
}

// hasTagLetter reports whether s contains at least one letter, which is what
// separates a real tag from a numeric reference like #123.
// hasTagLetter 判断 s 中是否至少含一个字母，用于区分真实标签与 #123 这类数字引用。
func hasTagLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}
