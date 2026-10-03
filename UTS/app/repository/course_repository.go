package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini-fiber/app/model"
)

type CourseRepository interface {
	FindAll(ctx context.Context, q model.CourseListQuery) ([]model.CourseWithKuota, error)
	FindByID(ctx context.Context, id int64) (model.Course, error)
	CountTerisi(ctx context.Context, courseID int64) (int, error)
}

type coursePostgresRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) CourseRepository {
	return &coursePostgresRepository{pool: pool}
}

func (r *coursePostgresRepository) FindAll(ctx context.Context, q model.CourseListQuery) ([]model.CourseWithKuota, error) {
	where := " WHERE 1 = 1"
	args := []any{}

	if q.Semester > 0 {
		args = append(args, q.Semester)
		where += fmt.Sprintf(" AND c.semester = $%d", len(args))
	}
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		n := len(args)
		where += fmt.Sprintf(" AND (c.kode_mk ILIKE $%d OR c.nama_mk ILIKE $%d)", n, n)
	}

	query := `
		SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
		       COALESCE(COUNT(e.id), 0) AS terisi,
		       c.kuota - COALESCE(COUNT(e.id), 0) AS sisa_kuota
		FROM courses c
		LEFT JOIN enrollments e ON e.course_id = c.id
		` + where + `
		GROUP BY c.id
		ORDER BY c.id ASC
	`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar course: %w", err)
	}
	defer rows.Close()

	result := []model.CourseWithKuota{}
	for rows.Next() {
		var c model.CourseWithKuota
		if err := rows.Scan(
			&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota,
			&c.Terisi, &c.SisaKuota,
		); err != nil {
			return nil, fmt.Errorf("membaca row course: %w", err)
		}

		if q.Available && c.SisaKuota <= 0 {
			continue
		}
		result = append(result, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query course: %w", err)
	}

	return result, nil
}

func (r *coursePostgresRepository) FindByID(ctx context.Context, id int64) (model.Course, error) {
	var c model.Course
	err := r.pool.QueryRow(ctx,
		`SELECT id, kode_mk, nama_mk, sks, semester, kuota, created_at, updated_at
		 FROM courses WHERE id = $1`, id,
	).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Course{}, ErrNotFound
		}
		return model.Course{}, fmt.Errorf("mengambil course: %w", err)
	}
	return c, nil
}

func (r *coursePostgresRepository) CountTerisi(ctx context.Context, courseID int64) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM enrollments WHERE course_id = $1", courseID,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("menghitung terisi: %w", err)
	}
	return total, nil
}
