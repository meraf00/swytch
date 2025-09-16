package domain

import "time"

type Job struct {
	ID        string
	Tasks     []Task
	CreatedAt time.Time
	UpdatedAt time.Time
}
