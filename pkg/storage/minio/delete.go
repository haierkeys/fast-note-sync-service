package minio

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"
)

func (p *MinIO) Delete(fileKey string) error {
	fileKey = fileurl.PathSuffixCheckAdd(p.Config.CustomPath, "/") + fileKey
	ctx := context.Background()
	bucket := p.GetBucket("")
	_, err := p.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileKey),
	})
	return err
}
