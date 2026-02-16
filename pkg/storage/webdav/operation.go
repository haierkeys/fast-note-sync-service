// operation.go

package webdav

import (
	"io"
	"os"

	"github.com/haierkeys/fast-note-sync-service/pkg/errors"
	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"
)

// SendFile 将本地文件上传到 WebDAV 服务器。
func (w *WebDAV) SendFile(fileKey string, file io.Reader, itype string) (string, error) {

	fileKey = fileurl.PathSuffixCheckAdd(w.Config.CustomPath, "/") + fileKey

	err := w.Client.MkdirAll(w.Config.CustomPath, 0644)
	if err != nil {
		return "", errors.Wrap(err, "webdav")
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return "", errors.Wrap(err, "webdav")
	}

	err = w.Client.Write(fileKey, content, os.ModePerm)

	if err != nil {
		return "", errors.Wrap(err, "webdav")
	}

	return fileKey, nil
}

// SendContent 将二进制内容上传到 WebDAV 服务器。
func (w *WebDAV) SendContent(fileKey string, content []byte) (string, error) {

	fileKey = fileurl.PathSuffixCheckAdd(w.Config.CustomPath, "/") + fileKey

	err := w.Client.Write(fileKey, content, os.ModePerm)

	if err != nil {
		return "", errors.Wrap(err, "webdav")
	}

	return fileKey, nil
}
