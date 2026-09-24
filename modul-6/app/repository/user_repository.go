package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"modul6/app/model"
)

type UserRepository interface {
	Create(ctx context.Context, username, email, passwordHash string) (model.User, error)
	FindByUsername(ctx context.Context, username string) (model.User, string, error)
	FindByID(ctx context.Context, id int) (model.User, error)
	List(ctx context.Context) ([]model.User, error)
	Update(ctx context.Context, id int, username, email string) (model.User, error)
	UpdateRole(ctx context.Context, id int, role string) (model.User, error)
	Delete(ctx context.Context, id int) error
}

type userPostgresRepository struct{ pool *pgxpool.Pool }

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userPostgresRepository{pool: pool}
}

const userColumns = "id, username, email, role, created_at"

func scanUser(row pgx.Row) (model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt)
	return u, err
}

func (r *userPostgresRepository) Create(ctx context.Context, username, email, passwordHash string) (model.User, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password_hash, role)
		 VALUES ($1, $2, $3, 'user')
		 RETURNING `+userColumns,
		username, email, passwordHash)
	user, err := scanUser(row)
	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("membuat user: %w", err)
	}
	return user, nil
}

func (r *userPostgresRepository) FindByUsername(ctx context.Context, username string) (model.User, string, error) {
	var u model.User
	var passwordHash string
	err := r.pool.QueryRow(ctx,
		`SELECT id, username, email, role, created_at, password_hash
		 FROM users WHERE username = $1`, username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.CreatedAt, &passwordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, "", ErrNotFound
		}
		return model.User{}, "", fmt.Errorf("mencari user: %w", err)
	}
	return u, passwordHash, nil
}

func (r *userPostgresRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mencari user: %w", err)
	}
	return user, nil
}

func (r *userPostgresRepository) List(ctx context.Context) ([]model.User, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+userColumns+` FROM users ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar user: %w", err)
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("membaca baris user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *userPostgresRepository) Update(ctx context.Context, id int, username, email string) (model.User, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE users SET username = $1, email = $2 WHERE id = $3 RETURNING `+userColumns,
		username, email, id)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, fmt.Errorf("mengubah user: %w", err)
	}
	return user, nil
}

func (r *userPostgresRepository) UpdateRole(ctx context.Context, id int, role string) (model.User, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE users SET role = $1 WHERE id = $2 RETURNING `+userColumns,
		role, id)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengubah role user: %w", err)
	}
	return user, nil
}

func (r *userPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("menghapus user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
