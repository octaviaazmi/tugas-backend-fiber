package seeder

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini-fiber/app/model"
	"siakad-mini-fiber/app/repository"
	"siakad-mini-fiber/helper"
)

func Run(
	ctx context.Context,
	pool *pgxpool.Pool,
	studentRepo repository.StudentRepository,
	courseRepo repository.CourseRepository,
) error {
	if err := seedAdmin(ctx, pool); err != nil {
		return err
	}
	if err := seedMahasiswa(ctx, pool, studentRepo, 20); err != nil {
		return err
	}
	if err := seedCourses(ctx, pool, 10); err != nil {
		return err
	}
	log.Println("Seeding selesai")
	return nil
}

func seedAdmin(ctx context.Context, pool *pgxpool.Pool) error {
	var count int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := helper.HashPassword("admin12345")
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx,
		"INSERT INTO users (email, password, role) VALUES ($1, $2, $3)",
		"admin@siakad.test", hash, "admin",
	)
	if err != nil {
		return err
	}
	log.Println("Admin dibuat: admin@siakad.test / admin12345")
	return nil
}

func seedMahasiswa(ctx context.Context, pool *pgxpool.Pool, studentRepo repository.StudentRepository, n int) error {
	prodiList := []string{"Sistem Informasi", "Teknik Informatika", "Manajemen Informatika"}

	for i := 1; i <= n; i++ {
		nim := fmt.Sprintf("187221%06d", i)
		email := fmt.Sprintf("mhs%d@siakad.test", i)

		var exists int
		if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", email).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}

		hash, err := helper.HashPassword(nim)
		if err != nil {
			return err
		}

		var userID int64
		err = pool.QueryRow(ctx,
			"INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING id",
			email, hash, "mahasiswa",
		).Scan(&userID)
		if err != nil {
			return err
		}

		ipk := 2.00 + rand.Float64()*2.00
		ipk = float64(int(ipk*100)) / 100

		_, err = studentRepo.Create(ctx, model.Student{
			UserID:      userID,
			NIM:         nim,
			Nama:        fmt.Sprintf("Mahasiswa %d", i),
			Prodi:       prodiList[i%len(prodiList)],
			Angkatan:    2022 + (i % 3),
			IPKTerakhir: ipk,
		})
		if err != nil {
			return err
		}
	}
	log.Printf("%d mahasiswa di-seed", n)
	return nil
}

func seedCourses(ctx context.Context, pool *pgxpool.Pool, n int) error {
	for i := 1; i <= n; i++ {
		kode := fmt.Sprintf("MK%03d", i)

		var exists int
		if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM courses WHERE kode_mk = $1", kode).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}

		sks := 2 + (i % 3)
		semester := 1 + (i % 8)

		_, err := pool.Exec(ctx,
			`INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota)
			 VALUES ($1, $2, $3, $4, $5)`,
			kode, fmt.Sprintf("Mata Kuliah %d", i), sks, semester, 30,
		)
		if err != nil {
			return err
		}
	}
	log.Printf("%d mata kuliah di-seed", n)
	return nil
}
