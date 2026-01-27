// Package util provides common utility functions
package util

import "regexp"

// WikiLink represents a wiki-style link extracted from content
type WikiLink struct {
	Path    string // The target path
	Alias   string // Optional alias from [[link|alias]]
	IsEmbed bool   // True if this is an embed (![[...]]) rather than a link ([[...]])
}

// wikiLinkRegex matches [[wiki-links]], [[link|alias]], and ![[embeds]] patterns
// Group 1: optional "!" prefix (embed marker)
// Group 2: path
// Group 3: optional alias
var wikiLinkRegex = regexp.MustCompile(`(!?)\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`)

// ParseWikiLinks extracts [[wiki-links]], [[link|alias]], and ![[embeds]] from content
// Returns a slice of WikiLink with path, optional alias, and embed flag
func ParseWikiLinks(content string) []WikiLink {
	if content == "" {
		return nil
	}

	matches := wikiLinkRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	// Use a map to deduplicate by path+isEmbed combination
	type linkKey struct {
		path    string
		isEmbed bool
	}
	seen := make(map[linkKey]bool)
	var links []WikiLink

	for _, match := range matches {
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
