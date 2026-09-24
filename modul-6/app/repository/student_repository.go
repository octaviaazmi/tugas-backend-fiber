package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"modul6/app/model"
)

type StudentRepository interface {
	Create(ctx context.Context, req model.CreateStudentRequest, ownerID int) (model.Student, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	List(ctx context.Context) ([]model.Student, error)
	Update(ctx context.Context, id int, req model.UpdateStudentRequest) (model.Student, error)
	Delete(ctx context.Context, id int) error
}

type studentPostgresRepository struct{ pool *pgxpool.Pool }

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

const studentColumns = "id, nim, nama, jurusan, angkatan, owner_id, created_at"

func scanStudent(row pgx.Row) (model.Student, error) {
	var s model.Student
	err := row.Scan(&s.ID, &s.NIM, &s.Nama, &s.Jurusan, &s.Angkatan, &s.OwnerID, &s.CreatedAt)
	return s, err
}

func (r *studentPostgresRepository) Create(ctx context.Context, req model.CreateStudentRequest, ownerID int) (model.Student, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO students (nim, nama, jurusan, angkatan, owner_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING `+studentColumns,
		req.NIM, req.Nama, req.Jurusan, req.Angkatan, ownerID)
	student, err := scanStudent(row)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("membuat student: %w", err)
	}
	return student, nil
}

func (r *studentPostgresRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+studentColumns+` FROM students WHERE id = $1`, id)
	student, err := scanStudent(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mencari student: %w", err)
	}
	return student, nil
}

func (r *studentPostgresRepository) List(ctx context.Context) ([]model.Student, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+studentColumns+` FROM students ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar student: %w", err)
	}
	defer rows.Close()

	students := []model.Student{}
	for rows.Next() {
		s, err := scanStudent(rows)
		if err != nil {
			return nil, fmt.Errorf("membaca baris student: %w", err)
		}
		students = append(students, s)
	}
	return students, rows.Err()
}

func (r *studentPostgresRepository) Update(ctx context.Context, id int, req model.UpdateStudentRequest) (model.Student, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE students SET nama = $1, jurusan = $2, angkatan = $3 WHERE id = $4 RETURNING `+studentColumns,
		req.Nama, req.Jurusan, req.Angkatan, id)
	student, err := scanStudent(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengubah student: %w", err)
	}
	return student, nil
}

func (r *studentPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM students WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("menghapus student: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
