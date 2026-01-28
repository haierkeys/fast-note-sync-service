package minio

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/haierkeys/fast-note-sync-service/pkg/fileurl"
	"github.com/haierkeys/fast-note-sync-service/pkg/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// UploadByFile uploads file
// UploadByFile 上传文件
func (p *MinIO) PutFile(fileKey string, file io.Reader, itype string) (string, error) {

	ctx := context.Background()
	bucket := p.Config.BucketName

	fileKey = fileurl.PathSuffixCheckAdd(p.Config.CustomPath, "/") + fileKey

	//  k, _ := h.Open()

	_, err := p.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(fileKey),
		Body:        file,
		ContentType: aws.String(itype),
	})

	if err != nil {
		return "", errors.Wrap(err, "minio")
	}

	return fileurl.PathSuffixCheckAdd(p.Config.BucketName, "/") + fileKey, nil
}

func (p *MinIO) PutContent(fileKey string, content []byte) (string, error) {

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
			p.logger.Warn("Bucket does not exist",
				zap.String(logger.FieldBucket, bucket),
				zap.Error(err),
			)
			err = noBucket
		}
	} else {
		err := s3.NewObjectExistsWaiter(p.S3Client).Wait(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(fileKey),
		}, time.Minute)
		if err != nil {
			p.logger.Warn("Failed attempt to wait for object to exist",
				zap.String(logger.FieldFileKey, fileKey),
				zap.String(logger.FieldBucket, bucket),
				zap.Error(err),
			)
		} else {
			_ = *output.Key
		}
	}

	return fileurl.PathSuffixCheckAdd(p.Config.BucketName, "/") + fileKey, errors.Wrap(err, "minio")
}

func (w *MinIO) DeleteFile(fileKey string) error {
	fileKey = fileurl.PathSuffixCheckAdd(w.Config.CustomPath, "/") + fileKey

	_, err := w.S3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(w.Config.BucketName),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return errors.Wrap(err, "minio: failed to delete file")
		// return errors.Wrap(err, "minio: 删除文件失败")
	}
	return nil
}
