package model

import "time"

// Entitas Utama
type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// ListQuery menyimpan parameter untuk pagination, pencarian, dan pengurutan
type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
}

// Meta menyimpan informasi halaman (pagination)
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// WebResponse adalah format standar balasan (response) dari API
type WebResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    any               `json:"data,omitempty"`
	Meta    *Meta             `json:"meta,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
}

// CreateStudentRequest adalah bentuk data yang diterima saat POST
type CreateStudentRequest struct {
	NIM   string  `json:"nim"`
	Name  string  `json:"name"`
	Grade float64 `json:"grade"`
}

// ReplaceStudentRequest adalah bentuk data yang diterima saat PUT
type ReplaceStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// PatchStudentRequest menggunakan pointer (*) agar bisa mendeteksi field mana yang kosong saat dikirim
type PatchStudentRequest struct {
	NIM      *string  `json:"nim"`
	Name     *string  `json:"name"`
	Grade    *float64 `json:"grade"`
	IsActive *bool    `json:"is_active"`
}
