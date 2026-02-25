package webdav

import (
	"path"
)

func (w *WebDAV) Delete(fileKey string) error {
	fileKey = path.Join("/", w.Config.CustomPath, fileKey)
	return w.Client.Remove(fileKey)
}
