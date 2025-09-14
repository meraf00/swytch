package domain

import "time"

type Job struct {
	Id        int
	Tasks     []Task
	CreatedAt time.Time
	UpdatedAt time.Time
}
