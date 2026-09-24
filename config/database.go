package config

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		GetEnv("DB_USER", "postgres"),
		GetEnv("DB_PASSWORD", "postgres"),
		GetEnv("DB_HOST", "localhost"),
		GetEnv("DB_PORT", "5432"),
		GetEnv("DB_NAME", "modul6_rbac"),
	)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return pool, nil
}
