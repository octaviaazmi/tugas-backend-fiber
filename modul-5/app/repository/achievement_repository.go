package repository

import (
	"context"
	"errors"

	"modul-5/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AchievementRepository interface {
	FindByID(ctx context.Context, id int) (model.Achievement, error)
	Create(ctx context.Context, a model.Achievement) (model.Achievement, error)
}

type achievementRepository struct {
	db *pgxpool.Pool
}

func NewAchievementRepository(db *pgxpool.Pool) AchievementRepository {
	return &achievementRepository{db: db}
}

func (r *achievementRepository) FindByID(ctx context.Context, id int) (model.Achievement, error) {
	var a model.Achievement
	query := `SELECT id_prestasi, id_mhs, nama_prestasi, juara, created_at
              FROM achievements WHERE id_prestasi = $1`

	err := r.db.QueryRow(ctx, query, id).
		Scan(&a.IDPrestasi, &a.IDMhs, &a.NamaPrestasi, &a.Juara, &a.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return a, ErrNotFound
		}
		return a, err
	}
	return a, nil
}

func (r *achievementRepository) Create(ctx context.Context, a model.Achievement) (model.Achievement, error) {
	query := `INSERT INTO achievements (id_mhs, nama_prestasi, juara)
              VALUES ($1, $2, $3)
              RETURNING id_prestasi, created_at`

	err := r.db.QueryRow(ctx, query, a.IDMhs, a.NamaPrestasi, a.Juara).
		Scan(&a.IDPrestasi, &a.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return a, errors.New("id_mhs tidak ditemukan di tabel students")
		}
		return a, err
	}
	return a, nil
}
