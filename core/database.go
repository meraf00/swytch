package core

import (
	"context"
	"fmt"
	"time"

	"github.com/meraf00/swytch/core/db"
	sql "github.com/meraf00/swytch/core/db/sqlc"
	"github.com/meraf00/swytch/core/lib/logger"
)

type Database interface {
	Ping(context.Context) error
	Queries() *sql.Queries
	WithTransaction(ctx context.Context, fn func(q *sql.Queries) error) error
}

func NewDatabase(config DatabaseConfig, log logger.Log) (Database, func(), error) {
	ctx := context.Background()

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		config.Host, config.Username, config.Password, config.Database, config.Port, config.SSLMode,
	)
	db, err := db.NewPG(ctx, dsn, log)

	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = db.Ping(ctx)
	log.Info("Connected to database.")

	cleanup := func() {
		db.Close()
		log.Info("Database connection closed.")
	}

	return db, cleanup, err
}
