package app

import (
	"context"
	"net/url"
)

type FileService interface {
	GenerateUploadUrl(ctx context.Context, objectName string) (*url.URL, error)
	GenerateDownloadUrl(ctx context.Context, objectName string) (*url.URL, error)
	UploadFile(ctx context.Context, objectName string, filePath string, contentType string) error
	DownloadFile(ctx context.Context, objectName string, downloadPath string) error
}
