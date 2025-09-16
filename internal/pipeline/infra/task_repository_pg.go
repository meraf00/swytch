package infra

import (
	"context"

	"github.com/meraf00/swytch/core"
	"github.com/meraf00/swytch/core/lib/hashids"
	"github.com/meraf00/swytch/internal/pipeline/domain"
)

type TaskRepositoryPG struct {
	db core.Database
	hs hashids.HashID
}

func NewTaskRepositoryPG(db core.Database, hs hashids.HashID) *TaskRepositoryPG {
	return &TaskRepositoryPG{
		db: db,
		hs: hs,
	}
}

func (r *TaskRepositoryPG) GetTaskByID(ctx context.Context, taskID string) (*domain.Task, error) {
	var task *domain.Task

	taskIDInt, err := r.hs.DecodeID(taskID)
	if err != nil {
		return nil, err
	}

	t, err := r.db.Queries().GetTaskByID(ctx, int32(taskIDInt))
	if err != nil {
		return nil, err
	}

	fileID, err := r.hs.EncodeID(uint(t.File.ID))
	if err != nil {
		return nil, err
	}

	task = &domain.Task{
		ID:           taskID,
		TargetFormat: t.TargetFormat,
		Status:       domain.TaskStatus(t.Status.TaskStatus),
		File: domain.File{
			ID:             fileID,
			ObjectName:     t.File.ObjectName.String(),
			OriginalName:   t.File.OriginalName,
			OriginalFormat: t.File.OriginalFormat,
		},
		CreatedAt: t.CreatedAt.Time,
		UpdatedAt: t.UpdatedAt.Time,
	}

	return task, nil
}
