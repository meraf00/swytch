package core

import (
	"context"
	"fmt"
	"time"

	"github.com/meraf00/swytch/core/lib/logger"
	"github.com/redis/go-redis/v9"
)

func NewRedis(cfg RedisConfig, log *logger.Log) (*redis.Client, func(), error) {
	options := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		// Connection pool settings
		MinIdleConns: 5,
		PoolSize:     20,
		PoolTimeout:  30 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	client := redis.NewClient(options)

	// Ping Redis to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info("Connected to Redis at ", cfg.Addr)

	shutdown := func() {
		if err := client.Close(); err != nil {
			log.Error("Failed to close Redis connection: ", err)
		} else {
			log.Info("Redis connection closed.")
		}
	}

	return client, shutdown, nil
}
