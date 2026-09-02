package model

// 1. Entitas Utama
type Student struct {
	ID       int     `json:"id"`
	NIM      string  `json:"nim"` // Tambahan field unik sesuai tugas
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// 2. Struct untuk POST (Membuat Data Baru)
// Semua field wajib. IsActive sengaja dihilangkan karena default-nya akan di-set true di handler.
type CreateStudentRequest struct {
	NIM   string  `json:"nim"`
	Name  string  `json:"name"`
	Grade float64 `json:"grade"`
}

// 3. Struct untuk PUT (Mengganti Seluruh Data)
// Semua field wajib diisi karena prinsip PUT adalah membuang data lama dan mengganti dengan yang baru
type ReplaceStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// 4. Struct untuk PATCH (Mengubah Sebagian Data)
// Menggunakan pointer (*) agar Go bisa membedakan antara field yang "tidak dikirim" (nil)
// dengan field yang dikirim tapi nilainya kosong/false.
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty"`
	Name     *string  `json:"name,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// 5. Amplop Baku untuk Semua Respons API
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// 6. Struct untuk Paginasi (Informasi Halaman)
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// 7. Struct untuk Menampung Query String (Filter, Sort, Pagination)
type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
}

// Offset menghitung berapa baris yang dilewati untuk halaman ini.
// Perhitungan ini pindah ke sini karena kini dipakai langsung oleh SQL.
func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}
