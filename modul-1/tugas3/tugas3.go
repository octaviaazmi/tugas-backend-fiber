package main

import "fmt"

// 1. Fungsi Swap (Menukar nilai melalui pointer)
func swap(a, b *int) {
	*a, *b = *b, *a
}

// 2. Fungsi updateSlice (Menambah item ke slice melalui pointer)
func updateSlice(s *[]string, newitem string) {
	*s = append(*s, newitem)
}

// ==========================================
// Fungsi Pembuktian: Value vs Pointer
// ==========================================
func ubahPassByValue(angka int) {
	angka = 100 // Hanya mengubah salinan lokal
}

func ubahPassByPointer(angka *int) {
	*angka = 100 // Mengubah nilai asli di alamat memori
}

func main() {
	// --- A. Demo Swap ---
	fmt.Println("=== 1. Demo Swap (Pointer) ===")
	x, y := 10, 20
	fmt.Printf("Sebelum swap : x = %d, y = %d\n", x, y)
	swap(&x, &y)
	fmt.Printf("Sesudah swap : x = %d, y = %d\n\n", x, y)

	// --- B. Demo Update Slice ---
	fmt.Println("=== 2. Demo Update Slice (Pointer) ===")
	keranjang := []string{"Apel", "Jeruk"}
	fmt.Println("Sebelum update :", keranjang)
	updateSlice(&keranjang, "Mangga")
	fmt.Println("Sesudah update :", keranjang, "\n")

	// --- C. Demo Pass by Value vs Pass by Pointer ---
	fmt.Println("=== 3. Perbandingan Value vs Pointer ===")
	nilaiAsli := 42

	fmt.Println("Nilai awal         :", nilaiAsli)

	ubahPassByValue(nilaiAsli)
	fmt.Println("Setelah by Value   :", nilaiAsli, "(Tidak berubah)")

	ubahPassByPointer(&nilaiAsli)
	fmt.Println("Setelah by Pointer :", nilaiAsli, "(Berubah jadi 100)")
}
