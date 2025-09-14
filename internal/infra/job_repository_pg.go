package infra

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/meraf00/swytch/core"
	"github.com/meraf00/swytch/core/db"
	sql "github.com/meraf00/swytch/core/db/sqlc"
	"github.com/meraf00/swytch/core/lib/hashids"
	"github.com/meraf00/swytch/internal/domain"
)

type JobRepositoryPG struct {
	db core.Database
	hs hashids.HashID
}

func NewJobRepositoryPG(db core.Database, hs hashids.HashID) *JobRepositoryPG {
	return &JobRepositoryPG{
		db: db,
		hs: hs,
	}
}

func (r *JobRepositoryPG) GetJobByID(ctx context.Context, jobID string) (*domain.Job, error) {
	var job *domain.Job

	jobIDInt, err := r.hs.DecodeID(jobID)
	if err != nil {
		return nil, err
	}

	j, err := r.db.Queries().GetJobByID(ctx, int32(jobIDInt))
	if err != nil {
		return nil, err
	}

	job = &domain.Job{
		ID:        jobID,
		CreatedAt: j.CreatedAt.Time,
		UpdatedAt: j.UpdatedAt.Time,
	}

	return job, nil
}

func (r *JobRepositoryPG) GetJobWithTasksAndFiles(ctx context.Context, jobID string) (*domain.Job, error) {
	jobIDInt, err := r.hs.DecodeID(jobID)
	if err != nil {
		return nil, err
	}

	job, err := r.GetJobByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	tasks, err := r.db.Queries().GetTasksByJobID(ctx, db.ToPGInt4(int32(jobIDInt)))
	if err != nil {
		return nil, err
	}

	job.Tasks = make([]domain.Task, len(tasks))

	for i, t := range tasks {
		taskID, err := r.hs.EncodeID(uint(t.ID))
		if err != nil {
			return nil, err
		}

		fileID, err := r.hs.EncodeID(uint(t.File.ID))
		if err != nil {
			return nil, err
		}

		job.Tasks[i] = domain.Task{
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
	}

	return job, nil
}

func (r *JobRepositoryPG) CreateJob(ctx context.Context, job *domain.Job) (*domain.Job, error) {
	var newJob *domain.Job

	err := r.db.WithTransaction(ctx, func(q *sql.Queries) error {
		j, err := q.CreateJob(ctx)
		if err != nil {
			return err
		}

		for i := range job.Tasks {
			t := &job.Tasks[i]

			var objectName pgtype.UUID
			err := objectName.Scan(t.File.ObjectName)
			if err != nil {
				return err
			}

			f, err := q.CreateFile(ctx, sql.CreateFileParams{
				ObjectName:     objectName,
				OriginalName:   t.File.OriginalName,
				OriginalFormat: t.File.OriginalFormat,
			})
			if err != nil {
				return err
			}

			t.File.ID, err = r.hs.EncodeID(uint(f.ID))
			if err != nil {
				return err
			}

			t.File.ObjectName = objectName.String()
			t.File.OriginalName = f.OriginalName
			t.File.OriginalFormat = f.OriginalFormat

			t.Status = domain.StatusPending

			task, err := q.CreateTask(ctx, sql.CreateTaskParams{
				JobID:        db.ToPGInt4(j.ID),
				FileID:       db.ToPGInt4(f.ID),
				TargetFormat: t.TargetFormat,
			})
			if err != nil {
				return err
			}

			taskID, err := r.hs.EncodeID(uint(task.ID))
			if err != nil {
				return err
			}
			t.ID = taskID
			t.TargetFormat = task.TargetFormat
		}

		jobID, err := r.hs.EncodeID(uint(j.ID))
		if err != nil {
			return err
		}
		newJob = &domain.Job{
			ID:        jobID,
			CreatedAt: j.CreatedAt.Time,
			UpdatedAt: j.UpdatedAt.Time,
			Tasks:     job.Tasks,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return newJob, nil
}
