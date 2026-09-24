package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool membuat koneksi pool ke database PostgreSQL
func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	// Coba ambil dari variabel DATABASE_URL di file .env
	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"),
		)
	}

	return pgxpool.New(ctx, dsn)
}
