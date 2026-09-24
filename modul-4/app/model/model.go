package model

import "time" //utk created_at

// Student merepresentasikan entitas mahasiswa di database.
type Student struct {
	ID        int       `json:"id" db:"id"`
	NIM       string    `json:"nim" db:"nim"`
	Name      string    `json:"name" db:"name"`
	Grade     float64   `json:"grade" db:"grade"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CreateStudentRequest merepresentasikan struktur data yang diterima saat membuat mahasiswa baru (POST).
type CreateStudentRequest struct {
	NIM   string  `json:"nim"`
	Name  string  `json:"name"`
	Grade float64 `json:"grade"`
}

// ReplaceStudentRequest merepresentasikan struktur data untuk mengganti seluruh data mahasiswa (PUT).
type ReplaceStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// PatchStudentRequest merepresentasikan struktur data untuk memperbarui sebagian data mahasiswa (PATCH).
// Menggunakan pointer agar bisa mendeteksi field mana saja yang dikirim oleh client.
type PatchStudentRequest struct {
	NIM      *string  `json:"nim"`
	Name     *string  `json:"name"`
	Grade    *float64 `json:"grade"`
	IsActive *bool    `json:"is_active"`
}

// ListQuery menampung parameter query string untuk proses filtering, pagination, dan sorting.
type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
}

// Meta menampung informasi metadata untuk pagination daftar data.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// WebResponse adalah standar format *envelope* JSON untuk respons API.
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

// Achievement adalah entitas prestasi mahasiswa.
type Achievement struct {
	IDPrestasi   int       `json:"id_prestasi"`
	IDMhs        int       `json:"id_mhs"`
	NamaPrestasi string    `json:"nama_prestasi"`
	Juara        string    `json:"juara"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateAchievementRequest adalah bentuk body saat POST prestasi baru.
type CreateAchievementRequest struct {
	IDMhs        int    `json:"id_mhs"`
	NamaPrestasi string `json:"nama_prestasi"`
	Juara        string `json:"juara"`
}
