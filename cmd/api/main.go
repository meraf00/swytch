package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/meraf00/swytch/core"
	"github.com/meraf00/swytch/core/lib/logger"
	"github.com/meraf00/swytch/internal"
)

func main() {

	log := logger.NewLogger()
	config := core.LoadConfig(log)

	db, shutdownDB, err := core.NewDatabase(config.Database, *log)
	if err != nil {
		log.Fatal("Failed to initialize database: ", err)
	}

	redisClient, shutdownRedis, err := core.NewRedis(config.Redis, log)
	if err != nil {
		log.Fatal("Failed to initialize Redis: ", err)
	}

	httpServer, shutdownServer := core.NewServer(config, log)

	// Start HTTP Server
	core.StartServer(httpServer, log)

	internal.InitModules(config, log, httpServer, db, redisClient)

	// wait till server starts
	<-httpServer.Started

	// Graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	<-shutdownChan
	shutdownServer()
	shutdownDB()
	shutdownRedis()

	log.Info("Application exited successfully.")
}
