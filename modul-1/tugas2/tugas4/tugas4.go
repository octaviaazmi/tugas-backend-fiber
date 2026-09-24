package main

import "fmt"

// 1. Deklarasi Struct
type Student struct {
	ID       int
	Name     string
	Grade    float64
	IsActive bool
}

// 2. Method GetInfo (Memakai Value Receiver karena hanya membaca data)
func (s Student) GetInfo() string {
	status := "Non-Aktif"
	if s.IsActive {
		status = "Aktif"
	}
	return fmt.Sprintf("ID: %d | Nama: %-7s | Nilai: %.2f | Status: %s", s.ID, s.Name, s.Grade, status)
}

// 3. Method UpdateGrade (Memakai Pointer Receiver karena mengubah data)
func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

// 4. Method Activate (Memakai Pointer Receiver karena mengubah data)
func (s *Student) Activate() {
	s.IsActive = true
}

// 5. Method Deactivate (Memakai Pointer Receiver karena mengubah data)
func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	fmt.Println("=== 4. Demonstrasi Struct Student dan Method ===")

	// Membuat data mahasiswa baru
	mhs1 := Student{
		ID:       2024001,
		Name:     "Pia",
		Grade:    85.5,
		IsActive: false, // Awalnya non-aktif
	}

	fmt.Println("-> Data Awal:")
	fmt.Println(mhs1.GetInfo())

	// Memperbarui nilai dan mengaktifkan status
	fmt.Println("\n-> Memperbarui Nilai dan Mengaktifkan Status...")
	mhs1.UpdateGrade(95.0)
	mhs1.Activate()

	fmt.Println("\n-> Data Setelah Diperbarui:")
	fmt.Println(mhs1.GetInfo())

	// Menonaktifkan kembali
	fmt.Println("\n-> Menonaktifkan Mahasiswa...")
	mhs1.Deactivate()
	fmt.Println(mhs1.GetInfo())
}
