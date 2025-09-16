package internal

import (
	"github.com/meraf00/swytch/core"
	"github.com/meraf00/swytch/core/lib/hashids"
	"github.com/meraf00/swytch/core/lib/logger"
	"github.com/meraf00/swytch/internal/pipeline/app"
	"github.com/meraf00/swytch/internal/pipeline/infra"
	handler "github.com/meraf00/swytch/internal/pipeline/interfaces/http"
	"github.com/redis/go-redis/v9"
)

func InitModules(config *core.AppConfig, logger *logger.Log, server *core.Server, db core.Database, rdb *redis.Client) {
	// Core
	hd, err := hashids.NewHashIDService(config.Encryption)
	if err != nil {
		logger.Fatalf("Failed to register auth module:", err)
	}

	// Repositories
	jobRepo := infra.NewJobRepositoryPG(db, hd)
	taskRepo := infra.NewTaskRepositoryPG(db, hd)

	// Services
	fileService, err := infra.NewMinioFileService(&config.Storage)
	if err != nil {
		logger.Fatalf("Failed to initiate minio service", err)
	}
	conversionService := app.NewConversionService(taskRepo, jobRepo, fileService)

	// API Surface
	apiRouter := server.ApiRouter

	// Files
	apiRouter.HandleFunc("/files", handler.HandleGetUploadPresignedURL(fileService)).Methods("POST")

	// Jobs and Tasks
	apiRouter.HandleFunc("/jobs", handler.HandleCreateJob(conversionService)).Methods("POST")
	apiRouter.HandleFunc("/jobs/{job_id}", handler.HandleGetJob(conversionService)).Methods("GET")
	apiRouter.HandleFunc("/jobs/{job_id}/tasks", handler.HandleGetJobTasks(conversionService)).Methods("GET")
	apiRouter.HandleFunc("/tasks/{task_id}/download", handler.HandleGetCompletedTaskDownloadURL(conversionService)).Methods("POST")
}
