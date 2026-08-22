package main

import "fmt"

func main() {
	// ==========================================
	// 1. DEKLARASI 5 VARIABEL DENGAN TIPE BERBEDA
	// ==========================================
	var namaMahasiswa string = "Octavia"
	var semester int = 5
	var ipk float64 = 3.85
	var isAktif bool = true
	var mataKuliah []string = []string{"Backend Lanjut", "Manajemen Proyek", "Software Testing"}

	fmt.Println("=== 1. Demonstrasi 5 Tipe Data Variabel ===")
	fmt.Printf("Nama       (string)  : %s\n", namaMahasiswa)
	fmt.Printf("Semester   (int)     : %d\n", semester)
	fmt.Printf("IPK        (float64) : %.2f\n", ipk)
	fmt.Printf("Status     (bool)    : %t\n", isAktif)
	fmt.Printf("Matkul     (slice)   : %v\n\n", mataKuliah)

	// ==========================================
	// 2. OPERASI MAP DATA MAHASISWA
	// ==========================================
	fmt.Println("=== 2. Operasi Map Data Nilai Mahasiswa ===")

	// A. Inisialisasi Map (Nama sebagai Kunci, Nilai sebagai Isinya)
	nilaiMahasiswa := make(map[string]float64)

	// B. Menambah Data ke Map
	nilaiMahasiswa["Pia"] = 92.5
	nilaiMahasiswa["Budi"] = 80.0
	nilaiMahasiswa["Sari"] = 88.5
	fmt.Println("Data awal map:", nilaiMahasiswa)

	// C. Membaca Data dengan Pengecekan Keberadaan (Pattern dua nilai di Go)
	namaCari := "Pia"
	if nilai, exists := nilaiMahasiswa[namaCari]; exists {
		fmt.Printf("-> [CEK] Nilai %s ditemukan: %.2f\n", namaCari, nilai)
	} else {
		fmt.Printf("-> [CEK] Data %s tidak ditemukan.\n", namaCari)
	}

	namaCari2 := "Zibril"
	if nilai, exists := nilaiMahasiswa[namaCari2]; exists {
		fmt.Printf("-> [CEK] Nilai %s ditemukan: %.2f\n", namaCari2, nilai)
	} else {
		fmt.Printf("-> [CEK] Data %s tidak ditemukan dalam sistem.\n", namaCari2)
	}

	// D. Menghapus Data dari Map
	delete(nilaiMahasiswa, "Budi")
	fmt.Println("-> [HAPUS] Menghapus data 'Budi'.")

	// E. Menelusuri (Looping) Seluruh Isi Map
	fmt.Println("\n-> [ITERASI] Daftar Mahasiswa yang Tersisa:")
	for nama, nilai := range nilaiMahasiswa {
		fmt.Printf("   • Nama: %-6s | Nilai: %.2f\n", nama, nilai)
	}
}
