package aws_s3

import (
	"context"

	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (p *S3) Delete(fileKey string) error {
	bucket := p.GetBucket("")
	fileKey = path.Join(p.Config.CustomPath, fileKey)

	_, err := p.S3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileKey),
	})
	return err
}
