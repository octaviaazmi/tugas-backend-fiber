//go:build integration

package integration_test

import (
	"context"
	"testing"

	"modul-12/app/repository"
)

func TestStudentRepository_InsertAndFind(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewStudentRepository(pool)

	// 1. Uji Fungsi Insert
	// Kita pakai helper seedStudent yang sudah kita buat di main_test.go
	student := seedStudent(t)

	if student.ID == 0 {
		t.Fatal("Gagal insert: ID masih 0")
	}

	// 2. Uji Fungsi FindByID
	found, err := repo.FindByID(ctx, int(student.ID))
	if err != nil {
		t.Fatalf("Gagal FindByID: %v", err)
	}

	// 3. Bandingkan hasilnya
	if found.Nama != student.Nama {
		t.Errorf("Nama tidak cocok. Dapat: %s, Seharusnya: %s", found.Nama, student.Nama)
	}

	// 4. Uji Kasus Data Tidak Ditemukan
	_, err = repo.FindByID(ctx, 9999999)
	if err == nil {
		t.Error("Seharusnya error ErrNotFound, tapi malah sukses")
	}
}
