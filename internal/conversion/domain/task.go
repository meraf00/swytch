package domain

import (
	"errors"
	"slices"
	"time"
)

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusProcessing TaskStatus = "processing"
	StatusCompleted  TaskStatus = "completed"
	StatusFailed     TaskStatus = "failed"
)

type Task struct {
	Id                int
	File              File
	TargetFormat      string
	ConvertedFileName string
	Status            TaskStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

var AllowedConversions = map[string][]string{
	// source -> dest
	"pdf":  {"pdf", "docx", "epub"},
	"docx": {"pdf", "docx", "epub"},
	"epub": {"pdf", "docx", "epub"},

	"png":  {"png", "webp", "jpeg", "svg", "pdf"},
	"jpeg": {"png", "webp", "jpeg", "svg", "pdf"},
	"webp": {"png", "webp", "jpeg", "svg", "pdf"},
	"svg":  {"png", "webp", "jpeg", "svg", "pdf"},
}

func NewTask(file File, targetFormat string) (*Task, error) {
	allowed, ok := AllowedConversions[file.OriginalFormat]
	if !ok {
		return nil, errors.New("unsupported source format: " + file.OriginalFormat)
	}

	valid := slices.Contains(allowed, targetFormat)
	if !valid {
		return nil, errors.New("conversion from " + file.OriginalFormat + " to " + targetFormat + " is not allowed")
	}

	now := time.Now()
	return &Task{
		File:         file,
		TargetFormat: targetFormat,
		Status:       StatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}
