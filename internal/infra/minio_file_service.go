package infra

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/meraf00/swytch/core"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioFileService struct {
	accessKeyID     string
	secretAccessKey string
	bucketName      string
	urlTTL          time.Duration
	client          *minio.Client
}

func NewMinioFileService(s *core.StorageConfig) (*MinioFileService, error) {
	minioClient, err := minio.New(s.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.AccessKeyID, s.SecretAccessKey, ""),
		Secure: s.UseSSL,
	})

	if err != nil {
		return nil, err
	}

	return &MinioFileService{
		accessKeyID:     s.AccessKeyID,
		secretAccessKey: s.SecretAccessKey,
		bucketName:      s.BucketName,
		urlTTL:          s.PresignedUrlTTL,
		client:          minioClient,
	}, nil
}

func (m *MinioFileService) GenerateUploadUrl(ctx context.Context, objectName string) (*url.URL, error) {
	return m.client.PresignedPutObject(ctx, m.bucketName, objectName, m.urlTTL)
}

func (m *MinioFileService) GenerateDownloadUrl(ctx context.Context, objectName string) (*url.URL, error) {
	reqParams := make(url.Values)
	reqParams.Set(
		"response-content-disposition",
		fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, objectName, objectName),
	)

	return m.client.PresignedGetObject(ctx, m.bucketName, objectName, m.urlTTL, reqParams)
}

func (m *MinioFileService) UploadFile(ctx context.Context, objectName string, filePath string, contentType string) error {
	_, err := m.client.FPutObject(ctx, m.bucketName, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (m *MinioFileService) DownloadFile(ctx context.Context, objectName string, downloadPath string) error {
	return m.client.FGetObject(ctx, m.bucketName, objectName, downloadPath, minio.GetObjectOptions{})
}
