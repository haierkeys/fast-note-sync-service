// operation.go

package webdav

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gookit/goutil/dump"
	"github.com/haierkeys/fast-note-sync-service/pkg/errors"
)

func (w *WebDAV) setModifiedTime(path string, modTime time.Time) error {
	// Construct full URL
	endpoint := w.Config.Endpoint
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}
	// path usually starts with / from PathSuffixCheckAdd, but let's be safe
	url := endpoint + strings.TrimPrefix(path, "/")

	// XML body for PROPPATCH
	// We use a custom namespace to avoid conflicts with system properties
	xmlBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8" ?>
<d:propertyupdate xmlns:d="DAV:" xmlns:u="http://haierkeys.github.io/ns/">
<d:set><d:prop><u:modification-time>%s</u:modification-time></d:prop></d:set>
</d:propertyupdate>`, modTime.Format(time.RFC3339))

	req, err := http.NewRequest("PROPPATCH", url, strings.NewReader(xmlBody))
	if err != nil {
		return err
	}

	req.SetBasicAuth(w.Config.User, w.Config.Password)
	req.Header.Set("Content-Type", "application/xml")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	// For WebDAV, 207 Multi-Status is effectively a success if the property set was successful,
	// checking strictly < 300 catches 200, 201, 204, 207.
	// To be more robust we could parse XML response but for this helper it's usually enough.

	return nil
}

// SendFile 将本地文件上传到 WebDAV 服务器。
func (w *WebDAV) SendFile(fileKey string, file io.Reader, itype string, modTime time.Time) (string, error) {

	fileKey = path.Join("/", w.Config.CustomPath, fileKey)
	dump.P(fileKey)
	dir := path.Dir(fileKey)
	if dir != "/" && dir != "." && dir != "" {
		err := w.Client.MkdirAll(dir, 0644)
		if err != nil {
			return "", errors.Wrap(err, "webdav")
		}
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return "", errors.Wrap(err, "webdav")
	}

	err = w.Client.Write(fileKey, content, os.ModePerm)

	if err != nil {
		return "", errors.Wrap(err, "webdav")
	}

	if !modTime.IsZero() {
		_ = w.setModifiedTime(fileKey, modTime)
	}

	return fileKey, nil
}

// SendContent 将二进制内容上传到 WebDAV 服务器。
func (w *WebDAV) SendContent(fileKey string, content []byte, modTime time.Time) (string, error) {

	fileKey = path.Join("/", w.Config.CustomPath, fileKey)
	dir := path.Dir(fileKey)
	if dir != "/" && dir != "." && dir != "" {
		err := w.Client.MkdirAll(dir, 0644)
		if err != nil {
			return "", errors.Wrap(err, "webdav")
		}
	}

	err := w.Client.Write(fileKey, content, os.ModePerm)

	if err != nil {
		return "", errors.Wrap(err, "webdav")
	}

	if !modTime.IsZero() {
		_ = w.setModifiedTime(fileKey, modTime)
	}

	return fileKey, nil
}
