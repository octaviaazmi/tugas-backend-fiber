package repository

import (
	"context"
	"errors"
	"fmt"

	"modul-5/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, u model.User) (model.User, error)
	GetByUsername(ctx context.Context, username string) (model.User, error)
	FindByUsername(ctx context.Context, username string) (model.User, error)
	GetByID(ctx context.Context, id int) (model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

func (r *userPostgresRepository) Create(ctx context.Context, u model.User) (model.User, error) {
	query := `INSERT INTO users (username, email, password, role)
	          VALUES ($1, $2, $3, $4)
	          RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query, u.Username, u.Email, u.Password, u.Role).
		Scan(&u.ID, &u.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("menyimpan user: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) GetByUsername(ctx context.Context, username string) (model.User, error) {
	query := `SELECT id, username, email, password, role, created_at 
	          FROM users WHERE username = $1`

	var u model.User
	err := r.pool.QueryRow(ctx, query, username).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("gagal mengambil user berdasarkan username: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) FindByUsername(ctx context.Context, username string) (model.User, error) {
	return r.GetByUsername(ctx, username)
}

func (r *userPostgresRepository) GetByID(ctx context.Context, id int) (model.User, error) {
	query := `SELECT id, username, email, password, role, created_at 
	          FROM users WHERE id = $1`

	var u model.User
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("gagal mengambil user berdasarkan id: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	return r.GetByID(ctx, id)
}
