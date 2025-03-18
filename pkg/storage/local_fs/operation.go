package local_fs

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path"

	"github.com/haierkeys/obsidian-better-sync-service/pkg/fileurl"
)

func (p *LocalFS) CheckSave() error {

	savePath := p.getSavePath()

	if fileurl.IsExist(savePath) {
		if err := fileurl.CreatePath(savePath, os.ModePerm); err != nil {
			return errors.New("failed to create the save-fileurl directory")
		}
	}
	if fileurl.IsPermission(savePath) {
		return errors.New("no permission to upload the save fileurl directory")
	}
	p.IsCheckSave = true
	return nil
}

func (p *LocalFS) getSavePath() string {
	return fileurl.PathSuffixCheckAdd(p.Config.SavePath, "/")
}

// SendFile  上传文件
func (p *LocalFS) SendFile(fileKey string, file io.Reader, itype string) (string, error) {
	if !p.IsCheckSave {
		if err := p.CheckSave(); err != nil {
			return "", err
		}
	}

	dstFileKey := p.getSavePath() + fileKey

	err := os.MkdirAll(path.Dir(dstFileKey), os.ModePerm)
	if err != nil {
		return "", err
	}

	out, err := os.Create(dstFileKey)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// file.Seek(0, 0)
	_, err = io.Copy(out, file)
	if err != nil {
		return "", err
	} else {
		return dstFileKey, nil
	}
}

func (p *LocalFS) SendContent(fileKey string, content []byte) (string, error) {

	if !p.IsCheckSave {
		if err := p.CheckSave(); err != nil {
			return "", err
		}
	}

	dstFileKey := p.getSavePath() + fileKey

	out, err := os.Create(dstFileKey)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, bytes.NewReader(content))
	if err != nil {
		return "", err
	} else {
		return dstFileKey, nil
	}
}
