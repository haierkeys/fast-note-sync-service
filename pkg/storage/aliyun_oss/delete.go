package aliyun_oss

import (
	"context"
	"path"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
)

func (p *OSS) Delete(fileKey string) error {
	fileKey = path.Join(p.Config.CustomPath, fileKey)

	request := &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(p.Config.BucketName),
		Key:    oss.Ptr(fileKey),
	}

	_, err := p.Client.DeleteObject(context.Background(), request)
	return err
}
