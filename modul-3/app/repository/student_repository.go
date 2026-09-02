package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"modul-3/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StudentRepository struct {
	db *pgxpool.Pool
}

// NewStudentRepository membuat instance repository student baru.
func NewStudentRepository(db *pgxpool.Pool) *StudentRepository {
	return &StudentRepository{db: db}
}

// 1. FindAll (Mengambil daftar mahasiswa dengan filter, sorting, dan paginasi dari database)
func (r *StudentRepository) FindAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error) {
	whereClauses := []string{"1=1"}
	var args []any
	argIdx := 1

	if q.IsActive != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *q.IsActive)
		argIdx++
	}

	if q.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(name) LIKE $%d OR LOWER(nim) LIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+strings.ToLower(q.Search)+"%")
		argIdx++
	}

	whereStr := strings.Join(whereClauses, " AND ")

	// Hitung total data untuk keperluan paginasi
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM students WHERE %s", whereStr)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total data: %w", err)
	}

	// Mapping kolom sort yang aman
	sortColumn := "id"
	switch q.Sort {
	case "name":
		sortColumn = "name"
	case "grade":
		sortColumn = "grade"
	}

	orderDir := "ASC"
	if q.Order == "desc" {
		orderDir = "DESC"
	}

	// Query utama dengan Limit dan Offset
	dataQuery := fmt.Sprintf(
		"SELECT id, nim, name, grade, is_active FROM students WHERE %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		whereStr, sortColumn, orderDir, argIdx, argIdx+1,
	)

	queryArgs := append(args, q.Limit, q.Offset())

	rows, err := r.db.Query(ctx, dataQuery, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil data mahasiswa: %w", err)
	}
	defer rows.Close()

	var students []model.Student
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive); err != nil {
			return nil, 0, fmt.Errorf("gagal membaca baris data: %w", err)
		}
		students = append(students, s)
	}

	if students == nil {
		students = []model.Student{}
	}

	return students, total, nil
}

// 2. FindByID (Mengambil 1 mahasiswa berdasarkan ID)
func (r *StudentRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
	var s model.Student
	query := "SELECT id, nim, name, grade, is_active FROM students WHERE id = $1"
	err := r.db.QueryRow(ctx, query, id).Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return s, fmt.Errorf("mahasiswa tidak ditemukan")
		}
		return s, fmt.Errorf("gagal query mahasiswa by id: %w", err)
	}
	return s, nil
}

// 3. Create (Menambah data mahasiswa baru ke PostgreSQL)
func (r *StudentRepository) Create(ctx context.Context, req model.CreateStudentRequest) (model.Student, error) {
	var s model.Student
	query := `
		INSERT INTO students (nim, name, grade, is_active)
		VALUES ($1, $2, $3, true)
		RETURNING id, nim, name, grade, is_active
	`
	err := r.db.QueryRow(ctx, query, req.NIM, req.Name, req.Grade).Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive)
	if err != nil {
		return s, fmt.Errorf("gagal menambah mahasiswa: %w", err)
	}
	return s, nil
}

// 4. Update (Mengganti seluruh data mahasiswa - PUT)
func (r *StudentRepository) Update(ctx context.Context, id int, req model.ReplaceStudentRequest) (model.Student, error) {
	var s model.Student
	query := `
		UPDATE students
		SET nim = $1, name = $2, grade = $3, is_active = $4
		WHERE id = $5
		RETURNING id, nim, name, grade, is_active
	`
	err := r.db.QueryRow(ctx, query, req.NIM, req.Name, req.Grade, req.IsActive, id).Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return s, fmt.Errorf("mahasiswa tidak ditemukan")
		}
		return s, fmt.Errorf("gagal update mahasiswa: %w", err)
	}
	return s, nil
}

// 5. Delete (Menghapus data mahasiswa)
func (r *StudentRepository) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM students WHERE id = $1"
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus mahasiswa: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("mahasiswa tidak ditemukan")
	}
	return nil
}

// 6. CheckNIMExists (Mengecek apakah NIM sudah dipakai mahasiswa lain)
func (r *StudentRepository) CheckNIMExists(ctx context.Context, nim string, excludeID int) (bool, error) {
	var count int
	var query string
	var err error

	if excludeID > 0 {
		query = "SELECT COUNT(*) FROM students WHERE LOWER(nim) = LOWER($1) AND id != $2"
		err = r.db.QueryRow(ctx, query, nim, excludeID).Scan(&count)
	} else {
		query = "SELECT COUNT(*) FROM students WHERE LOWER(nim) = LOWER($1)"
		err = r.db.QueryRow(ctx, query, nim).Scan(&count)
	}

	if err != nil {
		return false, fmt.Errorf("gagal cek duplikasi NIM: %w", err)
	}
	return count > 0, nil
}
