package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	tmtypes "github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/pkg/errors"
)

func (p *MinIO) GetBucket(bucketName string) string {

	// Get bucket
	if len(bucketName) <= 0 {
		bucketName = p.Config.BucketName
	}

	return bucketName
}

// UploadByFile 上传文件
func (p *MinIO) SendFile(fileKey string, file io.Reader, itype string, modTime time.Time) (string, error) {

	ctx := context.Background()
	bucket := p.GetBucket("")

	fileKey = fileurl.PathSuffixCheckAdd(p.Config.CustomPath, "/") + fileKey

	//  k, _ := h.Open()

	input := &transfermanager.UploadObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fileKey),
		Body:        file,
		ContentType: aws.String(itype),
	}

	if !modTime.IsZero() {
		input.Metadata = map[string]string{
			"modification-time": modTime.Format(time.RFC3339),
		}
	}

	_, err := p.TransferManager.UploadObject(ctx, input)

	if err != nil {
		return "", errors.Wrap(err, "minio")
	}

	return fileurl.PathSuffixCheckAdd(p.Config.BucketName, "/") + fileKey, nil
}

func (p *MinIO) SendContent(fileKey string, content []byte, modTime time.Time) (string, error) {

	ctx := context.Background()
	bucket := p.GetBucket("")

	fileKey = fileurl.PathSuffixCheckAdd(p.Config.CustomPath, "/") + fileKey

	input := &transfermanager.UploadObjectInput{
		Bucket:            aws.String(bucket),
		Key:               aws.String(fileKey),
		Body:              bytes.NewReader(content),
		ChecksumAlgorithm: tmtypes.ChecksumAlgorithmSha256,
	}

	if !modTime.IsZero() {
		input.Metadata = map[string]string{
			"modification-time": modTime.Format(time.RFC3339),
		}
	}

	output, err := p.TransferManager.UploadObject(ctx, input)
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

	return fileurl.PathSuffixCheckAdd(p.Config.BucketName, "/") + fileKey, errors.Wrap(err, "minio")
}
