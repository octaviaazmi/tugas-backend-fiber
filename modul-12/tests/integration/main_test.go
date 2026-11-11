//go:build integration

package integration_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"modul-12/app/model"
)

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	// Load file .env.test dari root folder modul-12
	_ = godotenv.Load(filepath.Join("..", "..", ".env.test"))

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		fmt.Println("TEST_DATABASE_URL tidak diisi: integration test dilewati")
		os.Exit(0)
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		fmt.Println("TEST_DATABASE_URL tidak valid:", err)
		os.Exit(1)
	}

	// PENGAMAN: Pastikan nama database berakhiran _test
	if !strings.HasSuffix(cfg.ConnConfig.Database, "_test") {
		fmt.Printf("menolak berjalan: database %q tidak berakhiran _test\n", cfg.ConnConfig.Database)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err = pgxpool.NewWithConfig(ctx, cfg)
	if err == nil {
		err = pool.Ping(ctx)
	}
	if err != nil {
		fmt.Println("gagal terhubung ke database test:", err)
		os.Exit(1)
	}

	// Jalankan migration
	if err := applyMigrations(ctx); err != nil {
		fmt.Println("gagal menjalankan migration:", err)
		os.Exit(1)
	}

	code := m.Run() // Jalankan semua test
	pool.Close()
	os.Exit(code)
}

func applyMigrations(ctx context.Context) error {
	files, _ := filepath.Glob(filepath.Join("..", "..", "migrations", "*.sql"))
	slices.Sort(files)
	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(f), err)
		}
	}
	return nil
}

// Fungsi bantuan untuk bikin nama acak
func uniqueSuffix(t *testing.T) string {
	t.Helper()
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		t.Fatal("membuat akhiran acak:", err)
	}
	return hex.EncodeToString(b)
}

// Fungsi bantuan untuk menyimpan mahasiswa khusus test
func seedStudent(t *testing.T) model.Student {
	t.Helper()
	ctx := context.Background()
	s := uniqueSuffix(t)
	student := model.Student{Nama: "Siswa_" + s}

	err := pool.QueryRow(ctx, "INSERT INTO students (nama) VALUES ($1) RETURNING id", student.Nama).Scan(&student.ID)
	if err != nil {
		t.Fatal("seed mahasiswa:", err)
	}

	// t.Cleanup akan menghapus data ini setelah test selesai
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM students WHERE id = $1", student.ID)
	})
	return student
}
