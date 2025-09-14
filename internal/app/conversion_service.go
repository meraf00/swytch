package app

import (
	"context"
	"net/url"

	"github.com/meraf00/swytch/internal/domain"
)

type CreateJobParams struct {
	Files []struct {
		ObjectName     string
		OriginalName   string
		OriginalFormat string
		TargetFormats  []string
	}
}

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

func (cs *ConversionService) GenerateTaskDownloadUrl(ctx context.Context, taskID string) (*url.URL, error) {
	task, err := cs.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return cs.fileService.GenerateDownloadUrl(ctx, task.ConvertedFileName)
}

func (cs *ConversionService) CreateJob(ctx context.Context, job *CreateJobParams) (string, error) {
	var tasks []domain.Task

	for _, file := range job.Files {
		for _, format := range file.TargetFormats {
			task, err := domain.NewTask(domain.File{
				ObjectName:     file.ObjectName,
				OriginalName:   file.OriginalName,
				OriginalFormat: file.OriginalFormat,
			}, format)

			if err != nil {
				return "", err
			}
			tasks = append(tasks, domain.Task{
				File:         task.File,
				TargetFormat: task.TargetFormat,
			})
		}
	}

	newJob, err := cs.jobRepo.CreateJob(ctx, &domain.Job{
		Tasks: tasks,
	})

	if err != nil {
		return "", err
	}

	return newJob.ID, nil
}

func (cs *ConversionService) GetJob(ctx context.Context, jobID string) (*domain.Job, error) {
	return cs.jobRepo.GetJobByID(ctx, jobID)
}

func (cs *ConversionService) GetJobTasks(ctx context.Context, jobID string) ([]domain.Task, error) {
	job, err := cs.jobRepo.GetJobWithTasksAndFiles(ctx, jobID)
	if err != nil {
		return nil, err
	}

	return job.Tasks, nil
}
