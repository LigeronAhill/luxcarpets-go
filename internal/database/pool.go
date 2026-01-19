package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, dbURL string) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		panic(err)
	}
	if err = pool.Ping(ctx); err != nil {
		panic(err)
	}
	if err = migrateDB(ctx, dbURL); err != nil {
		panic(err)
	}
	return pool
}
