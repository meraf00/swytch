package domain

import "time"

type File struct {
	ID             string
	ObjectName     string
	OriginalName   string
	OriginalFormat string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
