package core

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/meraf00/swytch/core/lib/logger"
	"github.com/meraf00/swytch/core/lib/middleware"
)

type Server struct {
	HttpServer *http.Server
	RootRouter *mux.Router
	ApiRouter  *mux.Router
	AuthRouter *mux.Router
	Started    chan struct{}
}

func NewServer(config *AppConfig, log logger.Log) (*Server, func()) {
	// Router
	router := mux.NewRouter()

	apiRouter := router.PathPrefix("/api").Subrouter()
	authRouter := router.PathPrefix("/auth").Subrouter()

	// Routes
	router.HandleFunc("/status", middleware.StatusHandler(config.StartedAt, log)).Methods("GET")

	// Middlewares
	logOpts := middleware.DefaultLoggerOptions()
	router.Use(middleware.HTTPLoggerMiddleware(log, logOpts))

	// Server
	server := http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.HTTP.Host, config.HTTP.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	shutdown := func() {
		log.Info("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v\n", err)
		}
		log.Info("Server gracefully stopped.")
	}

	return &Server{
		HttpServer: &server,
		RootRouter: router,
		ApiRouter:  apiRouter,
		AuthRouter: authRouter,
		Started:    make(chan struct{}),
	}, shutdown
}

func StartServer(server *Server, logger logger.Log) {
	go func() {
		logger.Infof("Starting server on %s", server.HttpServer.Addr)

		// notify main thread
		close(server.Started)

		if err := server.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Error starting server on %v\n", err)
		}
	}()
}
