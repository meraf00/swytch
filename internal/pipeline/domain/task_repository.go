package domain

import "context"

type TaskRepository interface {
	GetTaskByID(ctx context.Context, taskID string) (*Task, error)
}
