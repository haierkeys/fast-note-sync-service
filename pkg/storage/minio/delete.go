package minio

import (
	"context"

	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (p *MinIO) Delete(fileKey string) error {
	fileKey = path.Join(p.Config.CustomPath, fileKey)
	ctx := context.Background()
	bucket := p.GetBucket("")
	_, err := p.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileKey),
	})
	return err
}
