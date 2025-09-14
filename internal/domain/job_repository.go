package domain

import "context"

type JobRepository interface {
	GetJobByID(ctx context.Context, jobID string) (*Job, error)
	GetJobWithTasksAndFiles(ctx context.Context, jobID string) (*Job, error)
	CreateJob(ctx context.Context, job *Job) (*Job, error)
}
