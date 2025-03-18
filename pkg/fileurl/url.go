package fileurl

import (
	"net/url"
	"strings"
)

// UrlEscape 转义文件路径
func UrlEscape(fileKey string) string {
	if strings.Contains(fileKey, "/") {
		i := strings.LastIndex(fileKey, "/")
		fileKey = fileKey[:i+1] + url.PathEscape(fileKey[i+1:])
	} else {
		fileKey = url.PathEscape(fileKey)
	}
	return fileKey
}
