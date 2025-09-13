package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	sql "github.com/meraf00/swytch/core/db/sqlc"
	"github.com/meraf00/swytch/core/lib/logger"
)

type postgres struct {
	pool    *pgxpool.Pool
	queries *sql.Queries
}

func NewPG(ctx context.Context, connString string, log logger.Log) (*postgres, error) {
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Errorf("Invalid database config: %v", err)
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Errorf("Failed to connect to database: %v", err)
		return nil, err
	}
	return &postgres{
		pool:    pool,
		queries: sql.New(pool),
	}, nil
}

func (p *postgres) Queries() *sql.Queries {
	return p.queries
}

func (pg *postgres) WithTransaction(ctx context.Context, fn func(q *sql.Queries) error) error {
	tx, err := pg.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	queries := sql.New(tx)
	if err := fn(queries); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (pg *postgres) Ping(ctx context.Context) error {
	return pg.pool.Ping(ctx)
}

func (pg *postgres) Close() {
	pg.pool.Close()
}
