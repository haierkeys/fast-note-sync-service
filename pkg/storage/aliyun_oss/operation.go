package aliyun_oss

import (
	"bytes"
	"io"

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

func (p *OSS) SendFile(fileKey string, file io.Reader, itype string) (string, error) {
	if p.Bucket == nil {
		err := p.GetBucket("")
		if err != nil {
			return "", err
		}
	}
	fileKey = fileurl.PathSuffixCheckAdd(p.Config.CustomPath, "/") + fileKey
	err := p.Bucket.PutObject(fileKey, file)
	if err != nil {
		return "", err
	}
	return fileKey, nil
}

func (p *OSS) SendContent(fileKey string, content []byte) (string, error) {

	if p.Bucket == nil {
		err := p.GetBucket("")
		if err != nil {
			return "", err
		}
	}
	fileKey = fileurl.PathSuffixCheckAdd(p.Config.CustomPath, "/") + fileKey
	err := p.Bucket.PutObject(fileKey, bytes.NewReader(content))
	if err != nil {
		return "", err
	}
	return fileKey, nil
}
