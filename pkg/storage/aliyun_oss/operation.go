package aliyun_oss

import (
	"bytes"
	"io"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"
)

func (p *OSS) GetBucket(bucketName string) error {
	// Get bucket
	if len(bucketName) <= 0 {
		bucketName = p.Config.BucketName
	}
	var err error
	p.Bucket, err = p.Client.Bucket(bucketName)
	return err
}

func (p *OSS) SendFile(fileKey string, file io.Reader, itype string, modTime time.Time) (string, error) {
	if p.Bucket == nil {
		err := p.GetBucket("")
		if err != nil {
			return "", err
		}
	}
	fileKey = fileurl.PathSuffixCheckAdd(p.Config.CustomPath, "/") + fileKey

	var options []oss.Option
	if !modTime.IsZero() {
		options = append(options, oss.Meta("modification-time", modTime.Format(time.RFC3339)))
	}

	err := p.Bucket.PutObject(fileKey, file, options...)
	if err != nil {
		return "", err
	}
	return fileKey, nil
}

func (p *OSS) SendContent(fileKey string, content []byte, modTime time.Time) (string, error) {

	if p.Bucket == nil {
		err := p.GetBucket("")
		if err != nil {
			return "", err
		}
	}
	fileKey = fileurl.PathSuffixCheckAdd(p.Config.CustomPath, "/") + fileKey

	var options []oss.Option
	if !modTime.IsZero() {
		options = append(options, oss.Meta("modification-time", modTime.Format(time.RFC3339)))
	}

	err := p.Bucket.PutObject(fileKey, bytes.NewReader(content), options...)
	if err != nil {
		return "", err
	}
	return fileKey, nil
}
