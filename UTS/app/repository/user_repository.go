package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini-fiber/app/model"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (model.User, error)
	FindByID(ctx context.Context, id int64) (model.User, error)
	Create(ctx context.Context, u model.User) (model.User, error)
	CreateWithTx(ctx context.Context, tx pgx.Tx, u model.User) (model.User, error)
}

type userPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

const userColumns = "id, email, password, role, created_at, updated_at"

func (r *userPostgresRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		"SELECT "+userColumns+" FROM users WHERE email = $1", email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) FindByID(ctx context.Context, id int64) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		"SELECT "+userColumns+" FROM users WHERE id = $1", id,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}

func (r *userPostgresRepository) Create(ctx context.Context, u model.User) (model.User, error) {
	return createUser(ctx, r.pool, u)
}

func (r *userPostgresRepository) CreateWithTx(ctx context.Context, tx pgx.Tx, u model.User) (model.User, error) {
	return createUser(ctx, tx, u)
}

type queryRower interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func createUser(ctx context.Context, q queryRower, u model.User) (model.User, error) {
	err := q.QueryRow(ctx,
		`INSERT INTO users (email, password, role)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at, updated_at`,
		u.Email, u.Password, u.Role,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("menyimpan user: %w", err)
	}
	return u, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
