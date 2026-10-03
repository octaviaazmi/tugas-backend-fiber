package model

import "time"

type Student struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	NIM         string     `json:"nim"`
	Nama        string     `json:"nama"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir float64    `json:"ipk_terakhir"`
	DeletedAt   *time.Time `json:"-"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateStudentRequest struct {
	NIM         string  `json:"nim"`
	Nama        string  `json:"nama"`
	Email       string  `json:"email"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

type UpdateStudentRequest struct {
	Nama        string   `json:"nama"`
	Prodi       string   `json:"prodi"`
	Angkatan    *int     `json:"angkatan"`
	IPKTerakhir *float64 `json:"ipk_terakhir"`
}
