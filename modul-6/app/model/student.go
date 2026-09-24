package model

import "time"

type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Nama      string    `json:"nama"`
	Jurusan   string    `json:"jurusan"`
	Angkatan  int       `json:"angkatan"`
	OwnerID   int       `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateStudentRequest SENGAJA tidak punya field OwnerID.
// Kalau owner_id ada di sini, pemanggil bisa mengirim {"owner_id": 1}
// dan mengaku-aku sebagai pemilik lain — persis contoh mass assignment di A.8.
type CreateStudentRequest struct {
	NIM      string `json:"nim"`
	Nama     string `json:"nama"`
	Jurusan  string `json:"jurusan"`
	Angkatan int    `json:"angkatan"`
}

type UpdateStudentRequest struct {
	Nama     string `json:"nama"`
	Jurusan  string `json:"jurusan"`
	Angkatan int    `json:"angkatan"`
}
