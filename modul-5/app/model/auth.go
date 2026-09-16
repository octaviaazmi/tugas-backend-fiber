package model

import "time"

// User adalah akun untuk login. Berbeda dari Student.
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // tanda "-" = TIDAK PERNAH keluar sebagai JSON
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	// Perhatikan: TIDAK ADA field Role di sini. Kalau ada, siapa pun
	// bisa mendaftar sebagai admin (kerentanan mass assignment).
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // dalam detik
}

// RefreshToken adalah satu baris pada tabel refresh_tokens.
// Perhatikan: yang disimpan TokenHash, bukan tokennya sendiri.
type RefreshToken struct {
	ID        int64
	UserID    int
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

// AuthUser adalah identitas yang dibawa access token.
// Isinya sengaja minimal: hanya yang benar-benar diperlukan middleware.
type AuthUser struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}
