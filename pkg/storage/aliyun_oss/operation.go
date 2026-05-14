package aliyun_oss

import (
	"bytes"
	"context"
	"io"
	"path"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

func (p *OSS) SendFile(fileKey string, file io.Reader, itype string, modTime time.Time) (string, error) {
	fileKey = path.Join(p.Config.CustomPath, fileKey)

	request := &oss.PutObjectRequest{
		Bucket: oss.Ptr(p.Config.BucketName),
		Key:    oss.Ptr(fileKey),
		Body:   file,
	}

	if !modTime.IsZero() {
		request.Metadata = map[string]string{
			"modification-time": modTime.Format(time.RFC3339),
		}
	}

	_, err := p.Client.PutObject(context.Background(), request)
	if err != nil {
		return "", err
	}
	return fileKey, nil
}

func (p *OSS) SendContent(fileKey string, content []byte, modTime time.Time) (string, error) {
	fileKey = path.Join(p.Config.CustomPath, fileKey)

	request := &oss.PutObjectRequest{
		Bucket: oss.Ptr(p.Config.BucketName),
		Key:    oss.Ptr(fileKey),
		Body:   bytes.NewReader(content),
	}

	if !modTime.IsZero() {
		request.Metadata = map[string]string{
			"modification-time": modTime.Format(time.RFC3339),
		}
	}

	_, err := p.Client.PutObject(context.Background(), request)
	if err != nil {
		return "", err
	}
	return fileKey, nil
}
