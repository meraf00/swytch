package domain

import "time"

type File struct {
	Id             int
	OriginalName   string
	OriginalFormat string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
