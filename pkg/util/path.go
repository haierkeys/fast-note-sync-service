// Package util provides common utility functions
// Package util 提供通用工具函数
package util

import (
	"io"
	"os"
	"strings"
)

// ApplyDefaultFolder applies default folder prefix
// ApplyDefaultFolder 应用默认文件夹前缀
// When path does not contain "/" and defaultFolder is not empty, add defaultFolder as prefix to path
// 当 path 不包含 "/" 且 defaultFolder 非空时，将 defaultFolder 作为前缀添加到 path
// Example: ApplyDefaultFolder("note.md", "inbox") => "inbox/note.md"
// 例如: ApplyDefaultFolder("note.md", "inbox") => "inbox/note.md"
//
//	ApplyDefaultFolder("folder/note.md", "inbox") => "folder/note.md" (unchanged)
//	ApplyDefaultFolder("folder/note.md", "inbox") => "folder/note.md" (不变)
//	ApplyDefaultFolder("note.md", "") => "note.md" (unchanged)
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
// GeneratePathVariations 生成路径的所有后缀变体，用于反向链接匹配。
// 给定 "projects/test-backlinks/folder-a/note.md"，返回：
// ["note", "folder-a/note", "test-backlinks/folder-a/note", "projects/test-backlinks/folder-a/note"]
// 这允许匹配类似 [[note]], [[folder-a/note]] 等链接。
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
// ValidatePath 检查路径是否安全（无目录遍历）。
// 如果路径有效则返回 true，如果包含 ".." 则返回 false。
func ValidatePath(path string) bool {
	return !strings.Contains(path, "..")
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// MoveFile moves a file from src to dst, supporting cross-device move.
// MoveFile 将文件从 src 移动到 dst，支持跨设备移动。
func MoveFile(src, dst string) error {
	// Try rename first
	// 先尝试重命名
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// Check if the error is "invalid cross-device link" (EXDEV)
	// On Windows, this might also show up differently or Rename might fail cross-vol
	// We fallback to Copy + Remove for any error to be safe, or we can check error string
	// On Linux, EXDEV error message contains "invalid cross-device link"
	if strings.Contains(err.Error(), "invalid cross-device link") ||
		strings.Contains(strings.ToLower(err.Error()), "cross-device") {
		// Copy file content
		if copyErr := CopyFile(src, dst); copyErr != nil {
			return copyErr
		}
		// Remove source file
		return os.Remove(src)
	}

	return err
}
