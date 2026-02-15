// operation.go

package webdav

import (
	"io"
	"os"

	"github.com/haierkeys/fast-note-sync-service/pkg/errors"
	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"
)

// SendFile 将本地文件上传到 WebDAV 服务器�?
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

// SendContent 将二进制内容上传�?WebDAV 服务器�?
func (w *WebDAV) SendContent(fileKey string, content []byte) (string, error) {

	fileKey = fileurl.PathSuffixCheckAdd(w.Config.CustomPath, "/") + fileKey

	err := w.Client.Write(fileKey, content, os.ModePerm)

	if err != nil {
		return "", errors.Wrap(err, "webdav")
	}

	return fileKey, nil
}

// // DownloadFile �?WebDAV 服务器下载文件到本地�?
// func (w *WebDAV) DownloadFile(remotePath, localPath string) error {
// 	err := w.Client.DownloadFile(remotePath, localPath)
// 	if err != nil {
// 		return fmt.Errorf("下载文件失败: %v", err)
// 	}

// 	return nil
// }

// // DeleteFile �?WebDAV 服务器删除文件�?
// func (w *WebDAV) DeleteFile(remotePath string) error {
// 	err := w.Client.Remove(remotePath)
// 	if err != nil {
// 		return fmt.Errorf("删除文件失败: %v", err)
// 	}

// 	return nil
// }

// // MkDir �?WebDAV 服务器上创建目录�?
// func (w *WebDAV) MkDir(remotePath string) error {
// 	err := w.Client.Mkdir(remotePath)
// 	if err != nil {
// 		if !gowebdav.IsErrExist(err) {
// 			return fmt.Errorf("创建目录失败: %v", err)
// 		}
// 		// 如果目录已存在，忽略错误
// 		log.Printf("目录 %s 已存在，忽略错误", remotePath)
// 	}

// 	return nil
// }

// // ListFiles 列出 WebDAV 服务器上的文件和目录�?
// func (w *WebDAV) ListFiles(remotePath string) ([]string, error) {
// 	files, err := w.Client.ReadDir(remotePath)
// 	if err != nil {
// 		return nil, fmt.Errorf("列出文件失败: %v", err)
// 	}

// 	var fileNames []string
// 	for _, file := range files {
// 		fileNames = append(fileNames, file.Name())
// 	}

// 	return fileNames, nil
// }
