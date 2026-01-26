// Package util provides common utility functions
package util

import "regexp"

// WikiLink represents a wiki-style link extracted from content
type WikiLink struct {
	Path  string // The target path
	Alias string // Optional alias from [[link|alias]]
}

// wikiLinkRegex matches [[wiki-links]] and [[link|alias]] patterns
// Group 1: path, Group 2: optional alias
var wikiLinkRegex = regexp.MustCompile(`\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`)

// ParseWikiLinks extracts [[wiki-links]] and [[link|alias]] from content
// Returns a slice of WikiLink with path and optional alias
func ParseWikiLinks(content string) []WikiLink {
	if content == "" {
		return nil
	}

	matches := wikiLinkRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	// Use a map to deduplicate by path
	seen := make(map[string]bool)
	var links []WikiLink

	for _, match := range matches {
		path := match[1]
		if seen[path] {
			continue
		}
		seen[path] = true

		link := WikiLink{
			Path: path,
		}
		if len(match) > 2 && match[2] != "" {
			link.Alias = match[2]
		}
		links = append(links, link)
	}

	return links
}
