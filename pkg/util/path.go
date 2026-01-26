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
