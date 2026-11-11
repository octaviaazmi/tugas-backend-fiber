package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"modul-12/app/model"
)

var ErrNotFound = errors.New("data tidak ditemukan")

type StudentRepository interface {
	FindByID(ctx context.Context, id int) (model.Student, error)
	Insert(ctx context.Context, s model.Student) (model.Student, error)
}

type studentPostgresRepository struct {
	db *pgxpool.Pool
}

func NewStudentRepository(db *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{db: db}
}

func (r *studentPostgresRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
	var s model.Student
	err := r.db.QueryRow(ctx, "SELECT id, nama FROM students WHERE id = $1", id).Scan(&s.ID, &s.Nama)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, err
	}
	return s, nil
}

func (r *studentPostgresRepository) Insert(ctx context.Context, s model.Student) (model.Student, error) {
	err := r.db.QueryRow(ctx, "INSERT INTO students (nama) VALUES ($1) RETURNING id", s.Nama).Scan(&s.ID)
	if err != nil {
		return model.Student{}, err
	}
	return s, nil
}
