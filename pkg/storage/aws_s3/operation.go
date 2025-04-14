package aws_s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/haierkeys/obsidian-better-sync-service/pkg/fileurl"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/pkg/errors"
)

// UploadByFile 上传文件
func (p *S3) PutFile(fileKey string, file io.Reader, itype string) (string, error) {

	bucket := p.Config.BucketName
	ctx := context.Background()

	fileKey = fileurl.PathSuffixCheckAdd(p.Config.CustomPath, "/") + fileKey

	//  k, _ := h.Open()

	_, err := p.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fileKey),
		Body:        file,
		ContentType: aws.String(itype),
	})

	if err != nil {
		return "", errors.Wrap(err, "aws_s3")
	}

	return fileKey, nil
}

func (p *S3) PutContent(fileKey string, content []byte) (string, error) {

	ctx := context.Background()
	bucket := p.Config.BucketName

	fileKey = fileurl.PathSuffixCheckAdd(p.Config.CustomPath, "/") + fileKey

	input := &s3.PutObjectInput{
		Bucket:            aws.String(bucket),
		Key:               aws.String(fileKey),
		Body:              bytes.NewReader(content),
		ChecksumAlgorithm: types.ChecksumAlgorithmSha256,
	}
	output, err := p.S3Manager.Upload(ctx, input)
	if err != nil {
		var noBucket *types.NoSuchBucket
		if errors.As(err, &noBucket) {
			fmt.Printf("Bucket %s does not exist.\n", bucket)
			err = noBucket
		}
	} else {
		err := s3.NewObjectExistsWaiter(p.S3Client).Wait(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(fileKey),
		}, time.Minute)
		if err != nil {
			fmt.Printf("Failed attempt to wait for object %s to exist in %s.\n", fileKey, bucket)
		} else {
			_ = *output.Key
		}
	}

	return fileKey, errors.Wrap(err, "aws_s3")
}

func (w *S3) DeleteFile(fileKey string) error {
	fileKey = fileurl.PathSuffixCheckAdd(w.Config.CustomPath, "/") + fileKey

	_, err := w.S3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(w.Config.BucketName),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}
	return nil
}
