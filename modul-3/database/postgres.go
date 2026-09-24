package database

import (
	"context"
	"fmt"
	"log"

	"modul-3/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectPostgres membuat koneksi pool ke database PostgreSQL
func ConnectPostgres() (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.GetEnv("DB_USER", "postgres"),
		config.GetEnv("DB_PASSWORD", ""),
		config.GetEnv("DB_HOST", "localhost"),
		config.GetEnv("DB_PORT", "5432"),
		config.GetEnv("DB_NAME", "student_db"),
	)

	configPool, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("gagal parsing config database: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), configPool)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat connection pool: %w", err)
	}

	// Cek koneksi
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("gagal ping database: %w", err)
	}

	log.Println("Berhasil terhubung ke database PostgreSQL!")
	return pool, nil
}
