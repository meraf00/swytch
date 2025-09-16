package domain

import "context"

type FileRepository interface {
	GetFileByID(ctx context.Context, fileID string) (*File, error)
	CreateFile(ctx context.Context, file *File) (*File, error)
}
