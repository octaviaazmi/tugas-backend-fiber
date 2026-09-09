package repository

import (
	"context"
	"errors"
	"fmt"

	"modul-4/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data duplikat")
)

type StudentRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	Create(ctx context.Context, student model.Student) (model.Student, error)
	Update(ctx context.Context, student model.Student) (model.Student, error)
	Delete(ctx context.Context, id int) error
}

type studentRepository struct {
	db *pgxpool.Pool
}

func NewStudentRepository(db *pgxpool.Pool) StudentRepository {
	return &studentRepository{db: db}
}

func (r *studentRepository) FindAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error) {
	var total int
	err := r.db.QueryRow(ctx, "SELECT COUNT(id) FROM students").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (q.Page - 1) * q.Limit
	query := fmt.Sprintf("SELECT id, nim, name, grade, is_active, created_at FROM students ORDER BY %s %s LIMIT $1 OFFSET $2", q.Sort, q.Order)

	rows, err := r.db.Query(ctx, query, q.Limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []model.Student
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.CreatedAt); err != nil {
			return nil, 0, err
		}
		students = append(students, s)
	}

	return students, total, nil
}

func (r *studentRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
	var s model.Student
	err := r.db.QueryRow(ctx, "SELECT id, nim, name, grade, is_active, created_at FROM students WHERE id = $1", id).
		Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return s, ErrNotFound
		}
		return s, err
	}
	return s, nil
}

func (r *studentRepository) Create(ctx context.Context, student model.Student) (model.Student, error) {
	query := `INSERT INTO students (nim, name, grade, is_active) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	err := r.db.QueryRow(ctx, query, student.NIM, student.Name, student.Grade, student.IsActive).Scan(&student.ID, &student.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return student, ErrDuplicate
		}
		return student, err
	}
	return student, nil
}

func (r *studentRepository) Update(ctx context.Context, student model.Student) (model.Student, error) {
	query := `UPDATE students SET nim = $1, name = $2, grade = $3, is_active = $4 WHERE id = $5 RETURNING created_at`
	err := r.db.QueryRow(ctx, query, student.NIM, student.Name, student.Grade, student.IsActive, student.ID).Scan(&student.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return student, ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return student, ErrDuplicate
		}
		return student, err
	}
	return student, nil
}

func (r *studentRepository) Delete(ctx context.Context, id int) error {
	cmd, err := r.db.Exec(ctx, "DELETE FROM students WHERE id = $1", id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
