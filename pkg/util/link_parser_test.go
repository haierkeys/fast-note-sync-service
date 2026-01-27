// Package util provides common utility functions
package util

import (
	"testing"
)

func TestParseWikiLinks(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []WikiLink
	}{
		// Basic wikilinks (IsEmbed=false)
		{
			name:    "simple wikilink",
			content: "Check out [[Note Name]] for more info",
			expected: []WikiLink{
				{Path: "Note Name", Alias: "", IsEmbed: false},
			},
		},
		{
			name:    "wikilink with alias",
			content: "See [[Note Name|Display Text]] here",
			expected: []WikiLink{
				{Path: "Note Name", Alias: "Display Text", IsEmbed: false},
			},
		},
		{
			name:    "wikilink with heading",
			content: "Jump to [[Note Name#Heading]] section",
			expected: []WikiLink{
				{Path: "Note Name#Heading", Alias: "", IsEmbed: false},
			},
		},
		{
			name:    "wikilink with block reference",
			content: "Reference [[Note Name#^block-id]] here",
			expected: []WikiLink{
				{Path: "Note Name#^block-id", Alias: "", IsEmbed: false},
			},
		},

		// Embeds (IsEmbed=true)
		{
			name:    "simple embed",
			content: "Embedded: ![[Note Name]]",
			expected: []WikiLink{
				{Path: "Note Name", Alias: "", IsEmbed: true},
			},
		},
		{
			name:    "embed with heading",
			content: "Section embed: ![[Note Name#Heading]]",
			expected: []WikiLink{
				{Path: "Note Name#Heading", Alias: "", IsEmbed: true},
			},
		},
		{
			name:    "image embed",
			content: "Image: ![[Image.png]]",
			expected: []WikiLink{
				{Path: "Image.png", Alias: "", IsEmbed: true},
			},
		},
		{
			name:    "audio embed",
			content: "Audio: ![[Audio.mp3]]",
			expected: []WikiLink{
				{Path: "Audio.mp3", Alias: "", IsEmbed: true},
			},
		},
		{
			name:    "video embed",
			content: "Video: ![[Video.mp4]]",
			expected: []WikiLink{
				{Path: "Video.mp4", Alias: "", IsEmbed: true},
			},
		},
		{
			name:    "pdf embed",
			content: "Document: ![[Document.pdf]]",
			expected: []WikiLink{
				{Path: "Document.pdf", Alias: "", IsEmbed: true},
			},
		},

		// Mixed content
		{
			name:    "link and embed in same content",
			content: "Link: [[Note1]] and embed: ![[Note2]]",
			expected: []WikiLink{
				{Path: "Note1", Alias: "", IsEmbed: false},
				{Path: "Note2", Alias: "", IsEmbed: true},
			},
		},
		{
			name:    "same path as link and embed",
			content: "Link: [[Note]] and embed: ![[Note]]",
			expected: []WikiLink{
				{Path: "Note", Alias: "", IsEmbed: false},
				{Path: "Note", Alias: "", IsEmbed: true},
			},
		},

		// Should NOT capture - markdown links
		{
			name:     "markdown external link",
			content:  "Check [Display Text](https://example.com) here",
			expected: nil,
		},
		{
			name:     "markdown relative link",
			content:  "See [Display Text](note.md) for details",
			expected: nil,
		},
		{
			name:     "auto-linked URL",
			content:  "Visit https://example.com for more",
			expected: nil,
		},

		// Edge cases
		{
			name:     "empty content",
			content:  "",
			expected: nil,
		},
		{
			name:     "no links",
			content:  "Just plain text without any links",
			expected: nil,
		},
		{
			name:    "multiple links",
			content: "[[Note1]] and [[Note2]] and [[Note3|Alias]]",
			expected: []WikiLink{
				{Path: "Note1", Alias: "", IsEmbed: false},
				{Path: "Note2", Alias: "", IsEmbed: false},
				{Path: "Note3", Alias: "Alias", IsEmbed: false},
			},
		},
		{
			name:    "duplicate links deduplicated",
			content: "[[Note]] appears twice [[Note]]",
			expected: []WikiLink{
				{Path: "Note", Alias: "", IsEmbed: false},
			},
		},
		{
			name:    "link with path separators",
			content: "[[folder/subfolder/note]]",
			expected: []WikiLink{
				{Path: "folder/subfolder/note", Alias: "", IsEmbed: false},
			},
		},
		{
			name:    "embed with alias",
			content: "![[Image.png|400]]",
			expected: []WikiLink{
				{Path: "Image.png", Alias: "400", IsEmbed: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseWikiLinks(tt.content)

			if len(result) != len(tt.expected) {
				t.Errorf("ParseWikiLinks(%q) returned %d links, want %d", tt.content, len(result), len(tt.expected))
				t.Errorf("Got: %+v", result)
				t.Errorf("Want: %+v", tt.expected)
				return
			}

			for i, link := range result {
				if link.Path != tt.expected[i].Path {
					t.Errorf("Link[%d].Path = %q, want %q", i, link.Path, tt.expected[i].Path)
				}
				if link.Alias != tt.expected[i].Alias {
					t.Errorf("Link[%d].Alias = %q, want %q", i, link.Alias, tt.expected[i].Alias)
				}
				if link.IsEmbed != tt.expected[i].IsEmbed {
					t.Errorf("Link[%d].IsEmbed = %v, want %v", i, link.IsEmbed, tt.expected[i].IsEmbed)
				}
			}
		})
	}
}

func TestParseWikiLinks_ComplexContent(t *testing.T) {
	content := `# My Note

This note has various links:
- A regular link: [[Another Note]]
- A link with alias: [[Note|Custom Text]]
- An embedded note: ![[Embedded Note]]
- An embedded image: ![[photo.jpg]]
- A heading link: [[Note#Section]]
- A block reference: [[Note#^abc123]]

Some markdown links that should NOT be captured:
- [External](https://example.com)
- [Relative](./other.md)

More wikilinks at the end: [[Final Note]]
`

	result := ParseWikiLinks(content)

	expectedPaths := map[string]bool{
		"Another Note":    true,
		"Note":            true,
		"Embedded Note":   true,
		"photo.jpg":       true,
		"Note#Section":    true,
		"Note#^abc123":    true,
		"Final Note":      true,
	}

	if len(result) != len(expectedPaths) {
		t.Errorf("Expected %d links, got %d", len(expectedPaths), len(result))
		for _, link := range result {
			t.Logf("Found: %+v", link)
		}
	}

	for _, link := range result {
		if !expectedPaths[link.Path] {
			t.Errorf("Unexpected link path: %q", link.Path)
		}
	}

	// Verify embeds are correctly identified
	for _, link := range result {
		switch link.Path {
		case "Embedded Note", "photo.jpg":
			if !link.IsEmbed {
				t.Errorf("Link %q should be an embed", link.Path)
			}
		default:
			if link.IsEmbed {
				t.Errorf("Link %q should NOT be an embed", link.Path)
			}
		}
	}
}
