package webdav

import (
	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"
)

func (w *WebDAV) Delete(fileKey string) error {
	fileKey = fileurl.PathSuffixCheckAdd(w.Config.CustomPath, "/") + fileKey
	return w.Client.Remove(fileKey)
}
