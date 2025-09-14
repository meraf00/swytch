package domain

type JobRepository interface {
	GetJobByID(jobID string) (*Job, error)
}
