package app

import (
	"context"
	"net/url"

	"github.com/meraf00/swytch/internal/domain"
)

type ConversionService struct {
	taskRepo    domain.TaskRepository
	jobRepo     domain.JobRepository
	fileService FileService
}

func NewConversionService(
	taskRepo domain.TaskRepository,
	jobRepo domain.JobRepository,
	fileService FileService,
) *ConversionService {
	return &ConversionService{
		taskRepo:    taskRepo,
		jobRepo:     jobRepo,
		fileService: fileService,
	}
}

func (t *ConversionService) GenerateTaskDownloadUrl(ctx context.Context, taskID string) (*url.URL, error) {
	task, err := t.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return t.fileService.GenerateDownloadUrl(ctx, task.ConvertedFileName)
}
