package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini-fiber/app/model"
)

type EnrollmentRepository interface {
	FindByID(ctx context.Context, id int64) (model.Enrollment, error)
	Exists(ctx context.Context, studentID, courseID int64, tahunAkademik string) (bool, error)
	ExistsWithTx(ctx context.Context, tx pgx.Tx, studentID, courseID int64, tahunAkademik string) (bool, error)
	TotalSKS(ctx context.Context, studentID int64, tahunAkademik string) (int, error)
	TotalSKSWithTx(ctx context.Context, tx pgx.Tx, studentID int64, tahunAkademik string) (int, error)
	Create(ctx context.Context, e model.Enrollment) (model.Enrollment, error)
	CreateWithTx(ctx context.Context, tx pgx.Tx, e model.Enrollment) (model.Enrollment, error)
	Delete(ctx context.Context, id int64) error
	FindByStudent(ctx context.Context, studentID int64) ([]model.EnrollmentDetail, error)
}

type EnrollmentDetail struct {
	EnrollmentID  int64
	CourseID      int64
	KodeMK        string
	NamaMK        string
	SKS           int
	Semester      int
	TahunAkademik string
}

type enrollmentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) EnrollmentRepository {
	return &enrollmentPostgresRepository{pool: pool}
}

func (r *enrollmentPostgresRepository) FindByID(ctx context.Context, id int64) (model.Enrollment, error) {
	var e model.Enrollment
	err := r.pool.QueryRow(ctx,
		`SELECT id, student_id, course_id, tahun_akademik, created_at
		 FROM enrollments WHERE id = $1`, id,
	).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, fmt.Errorf("mengambil enrollment: %w", err)
	}
	return e, nil
}

func (r *enrollmentPostgresRepository) Exists(ctx context.Context, studentID, courseID int64, tahunAkademik string) (bool, error) {
	return existsEnrollment(ctx, r.pool, studentID, courseID, tahunAkademik)
}

func (r *enrollmentPostgresRepository) ExistsWithTx(ctx context.Context, tx pgx.Tx, studentID, courseID int64, tahunAkademik string) (bool, error) {
	return existsEnrollment(ctx, tx, studentID, courseID, tahunAkademik)
}

func existsEnrollment(ctx context.Context, q queryRower, studentID, courseID int64, tahunAkademik string) (bool, error) {
	var count int
	err := q.QueryRow(ctx,
		`SELECT COUNT(*) FROM enrollments
		 WHERE student_id = $1 AND course_id = $2 AND tahun_akademik = $3`,
		studentID, courseID, tahunAkademik,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("cek duplikasi: %w", err)
	}
	return count > 0, nil
}

func (r *enrollmentPostgresRepository) TotalSKS(ctx context.Context, studentID int64, tahunAkademik string) (int, error) {
	return totalSKS(ctx, r.pool, studentID, tahunAkademik)
}

func (r *enrollmentPostgresRepository) TotalSKSWithTx(ctx context.Context, tx pgx.Tx, studentID int64, tahunAkademik string) (int, error) {
	return totalSKS(ctx, tx, studentID, tahunAkademik)
}

func totalSKS(ctx context.Context, q queryRower, studentID int64, tahunAkademik string) (int, error) {
	var total int
	err := q.QueryRow(ctx,
		`SELECT COALESCE(SUM(c.sks), 0)
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1 AND e.tahun_akademik = $2`,
		studentID, tahunAkademik,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("menghitung total SKS: %w", err)
	}
	return total, nil
}

func (r *enrollmentPostgresRepository) Create(ctx context.Context, e model.Enrollment) (model.Enrollment, error) {
	return createEnrollment(ctx, r.pool, e)
}

func (r *enrollmentPostgresRepository) CreateWithTx(ctx context.Context, tx pgx.Tx, e model.Enrollment) (model.Enrollment, error) {
	return createEnrollment(ctx, tx, e)
}

func createEnrollment(ctx context.Context, q queryRower, e model.Enrollment) (model.Enrollment, error) {
	err := q.QueryRow(ctx,
		`INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at`,
		e.StudentID, e.CourseID, e.TahunAkademik,
	).Scan(&e.ID, &e.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return model.Enrollment{}, ErrDuplicate
		}
		return model.Enrollment{}, fmt.Errorf("menyimpan enrollment: %w", err)
	}
	return e, nil
}

func (r *enrollmentPostgresRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM enrollments WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("menghapus enrollment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *enrollmentPostgresRepository) FindByStudent(ctx context.Context, studentID int64) ([]model.EnrollmentDetail, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT e.id, c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, e.tahun_akademik
		 FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1
		 ORDER BY e.id ASC`, studentID,
	)
	if err != nil {
		return nil, fmt.Errorf("mengambil enrollment student: %w", err)
	}
	defer rows.Close()

	result := []model.EnrollmentDetail{}
	for rows.Next() {
		var d model.EnrollmentDetail
		if err := rows.Scan(
			&d.EnrollmentID, &d.CourseID, &d.KodeMK, &d.NamaMK,
			&d.SKS, &d.Semester, &d.TahunAkademik,
		); err != nil {
			return nil, fmt.Errorf("membaca row enrollment: %w", err)
		}
		result = append(result, d)
	}
	return result, rows.Err()
}
