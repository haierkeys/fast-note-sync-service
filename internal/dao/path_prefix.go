package dao

import "strings"

func isPathWithinPrefix(path, prefix string) bool {
	path = strings.Trim(path, "/")
	prefix = strings.Trim(prefix, "/")
	if path == "" || prefix == "" || path == prefix {
		return false
	}
	return strings.HasPrefix(path, prefix+"/")
}
