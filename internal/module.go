package internal

import (
	"github.com/meraf00/swytch/core"
	"github.com/meraf00/swytch/core/lib/logger"
	"github.com/meraf00/swytch/internal/app"
	"github.com/meraf00/swytch/internal/infra"
	"github.com/meraf00/swytch/internal/interface/handler"
	"github.com/redis/go-redis/v9"
)

func InitModules(config *core.AppConfig, logger *logger.Log, server *core.Server, db core.Database, rdb *redis.Client) {
	// Core
	// hd, err := hashids.NewHashIDService(config.Encryption)
	// if err != nil {
	// 	logger.Fatalf("Failed to register auth module:", err)
	// }

	fileService, err := infra.NewMinioFileService(&config.Storage)
	if err != nil {
		logger.Fatalf("Failed to initiate minio service", err)
	}

	conversionService := app.NewConversionService(nil, nil, fileService)

	// API Surface
	apiRouter := server.ApiRouter

	// Files
	apiRouter.HandleFunc("/file", handler.HandleFileUpload(fileService)).Methods("POST")

	// Tasks
	apiRouter.HandleFunc("/task/{task_id}/download", handler.HandleTaskDownload(conversionService)).Methods("POST")

}
