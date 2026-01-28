// Package util provides common utility functions
// Package util 提供通用工具函数
package util

import "regexp"

// WikiLink represents a wiki-style link extracted from content // WikiLink 表示从内容中提取的维基风格链接
type WikiLink struct {
	Path    string // The target path // 目标路径
	Alias   string // Optional alias from [[link|alias]] // 可选别名
	IsEmbed bool   // True if this is an embed (![[...]]) rather than a link ([[...]]) // 是否为嵌入 (![[...]])
}

// wikiLinkRegex matches [[wiki-links]], [[link|alias]], and ![[embeds]] patterns
// Group 1: optional "!" prefix (embed marker) // 可选的 "!" 前缀（嵌入标记）
// Group 2: path // 路径
// Group 3: optional alias // 可选别名
// wikiLinkRegex 匹配 [[wiki-links]], [[link|alias]], 和 ![[embeds]] 模式
var wikiLinkRegex = regexp.MustCompile(`(!?)\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`)

// ParseWikiLinks extracts [[wiki-links]], [[link|alias]], and ![[embeds]] from content
// Returns a slice of WikiLink with path, optional alias, and embed flag
// ParseWikiLinks 从内容中提取 [[wiki-links]], [[link|alias]], 和 ![[embeds]]
// 返回包含路径、可选别名和嵌入标记的 WikiLink 切片
func ParseWikiLinks(content string) []WikiLink {
	if content == "" {
		return nil
	}

	matches := wikiLinkRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	// Use a map to deduplicate by path+isEmbed combination
	// 使用 map 按 path+isEmbed 组合进行去重
	type linkKey struct {
		path    string
		isEmbed bool
	}
	seen := make(map[linkKey]bool)
	var links []WikiLink

	for _, match := range matches {
		// Process match // 处理匹配项
		isEmbed := match[1] == "!"
		path := match[2]
		key := linkKey{path: path, isEmbed: isEmbed}
		if seen[key] {
			continue
		}
		seen[key] = true

		link := WikiLink{
			Path:    path,
			IsEmbed: isEmbed,
		}
		if len(match) > 3 && match[3] != "" {
			link.Alias = match[3]
		}
		links = append(links, link)
	}

	return links
}
