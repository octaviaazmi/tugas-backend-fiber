package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini-fiber/app/model"
)

type StudentRepository interface {
	FindAll(ctx context.Context, q model.StudentListQuery) ([]model.Student, int64, error)
	FindByID(ctx context.Context, id int64) (model.Student, error)
	FindByUserID(ctx context.Context, userID int64) (model.Student, error)
	Create(ctx context.Context, s model.Student) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	SoftDelete(ctx context.Context, id int64) error
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

const studentColumns = "id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at, updated_at"

func (r *studentPostgresRepository) FindAll(ctx context.Context, q model.StudentListQuery) ([]model.Student, int64, error) {
	where := " WHERE deleted_at IS NULL"
	args := []any{}

	if q.Prodi != "" {
		args = append(args, q.Prodi)
		where += fmt.Sprintf(" AND prodi = $%d", len(args))
	}
	if q.Angkatan > 0 {
		args = append(args, q.Angkatan)
		where += fmt.Sprintf(" AND angkatan = $%d", len(args))
	}
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		n := len(args)
		where += fmt.Sprintf(" AND (nim ILIKE $%d OR nama ILIKE $%d)", n, n)
	}

	// Count total
	var total int64
	countQuery := "SELECT COUNT(*) FROM students" + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("menghitung student: %w", err)
	}

	// Order
	orderBy := "id ASC"
	switch q.Sort {
	case "nama":
		orderBy = "nama ASC"
	case "-ipk_terakhir":
		orderBy = "ipk_terakhir DESC"
	}

	// Pagination
	args = append(args, q.PerPage, (q.Page-1)*q.PerPage)
	limitIdx := len(args) - 1
	offsetIdx := len(args)

	query := fmt.Sprintf(
		"SELECT %s FROM students%s ORDER BY %s LIMIT $%d OFFSET $%d",
		studentColumns, where, orderBy, limitIdx, offsetIdx,
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar student: %w", err)
	}
	defer rows.Close()

	result := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan,
			&s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("membaca row student: %w", err)
		}
		result = append(result, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("membaca hasil query student: %w", err)
	}

	return result, total, nil
}

func (r *studentPostgresRepository) FindByID(ctx context.Context, id int64) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		"SELECT "+studentColumns+" FROM students WHERE id = $1 AND deleted_at IS NULL", id,
	).Scan(
		&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan,
		&s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil student: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) FindByUserID(ctx context.Context, userID int64) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		"SELECT "+studentColumns+" FROM students WHERE user_id = $1 AND deleted_at IS NULL", userID,
	).Scan(
		&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan,
		&s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil student by user_id: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) Create(ctx context.Context, s model.Student) (model.Student, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at, updated_at`,
		s.UserID, s.NIM, s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("menyimpan student: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) Update(ctx context.Context, s model.Student) (model.Student, error) {
	err := r.pool.QueryRow(ctx,
		`UPDATE students
		 SET nama = $1, prodi = $2, angkatan = $3, ipk_terakhir = $4, updated_at = NOW()
		 WHERE id = $5 AND deleted_at IS NULL
		 RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at, created_at, updated_at`,
		s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir, s.ID,
	).Scan(
		&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan,
		&s.IPKTerakhir, &s.DeletedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("memperbarui student: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) SoftDelete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx,
		"UPDATE students SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id,
	)
	if err != nil {
		return fmt.Errorf("menghapus student: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

var _ = strings.TrimSpace
