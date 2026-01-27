// Package util 提供通用工具函数
package util

import "strings"

// ApplyDefaultFolder 应用默认文件夹前缀
// 当 path 不包含 "/" 且 defaultFolder 非空时，将 defaultFolder 作为前缀添加到 path
// 例如: ApplyDefaultFolder("note.md", "inbox") => "inbox/note.md"
//
//	ApplyDefaultFolder("folder/note.md", "inbox") => "folder/note.md" (不变)
//	ApplyDefaultFolder("note.md", "") => "note.md" (不变)
func ApplyDefaultFolder(path, defaultFolder string) string {
	if defaultFolder == "" || strings.Contains(path, "/") {
		return path
	}
	return strings.TrimSuffix(defaultFolder, "/") + "/" + path
}

// GeneratePathVariations generates all suffix variations of a path for backlink matching.
// Given "projects/test-backlinks/folder-a/note.md", returns:
// ["note", "folder-a/note", "test-backlinks/folder-a/note", "projects/test-backlinks/folder-a/note"]
// This allows matching links like [[note]], [[folder-a/note]], etc.
func GeneratePathVariations(path string) []string {
	// Strip .md extension if present
	path = strings.TrimSuffix(path, ".md")

	if path == "" {
		return nil
	}

	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return nil
	}

	// Build progressively longer suffixes from right to left
	variations := make([]string, 0, len(parts))
	for i := len(parts) - 1; i >= 0; i-- {
		suffix := strings.Join(parts[i:], "/")
		variations = append(variations, suffix)
	}

	return variations
}

// ValidatePath checks if a path is safe (no directory traversal).
// Returns true if the path is valid, false if it contains "..".
func ValidatePath(path string) bool {
	return !strings.Contains(path, "..")
}
