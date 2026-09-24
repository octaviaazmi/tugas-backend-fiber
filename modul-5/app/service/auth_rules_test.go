package service

import (
	"testing"

	"modul-5/app/model"
)

// Perhatikan: pengujian ini tidak menyalakan server, tidak menyentuh
// database, dan tidak membuat fiber.Ctx.

func TestCheckPasswordStrength(t *testing.T) {
	cases := []struct {
		password string
		wantErr  bool
		alasan   string
	}{
		{"rahasia123", false, "valid: 10 karakter, ada huruf dan angka"},
		{"pendek1", true, "kurang dari 8 karakter"},
		{"semuahuruf", true, "tidak ada angka"},
		{"12345678", true, "termasuk daftar password umum"},
		{"password1", true, "termasuk daftar password umum"},
	}

	for _, tc := range cases {
		got := CheckPasswordStrength(tc.password)
		if tc.wantErr && got == "" {
			t.Errorf("password %q: harusnya ditolak (%s), tapi lolos", tc.password, tc.alasan)
		}
		if !tc.wantErr && got != "" {
			t.Errorf("password %q: harusnya lolos, tapi ditolak dengan pesan %q", tc.password, got)
		}
	}
}

func TestValidateRegister(t *testing.T) {
	// Semua field kosong -> harus ada 3 error
	errs := ValidateRegister(model.RegisterRequest{})
	if len(errs) != 3 {
		t.Errorf("request kosong: harap 3 error, dapat %d (%v)", len(errs), errs)
	}

	// Request yang benar -> tidak boleh ada error
	valid := model.RegisterRequest{
		Username: "sari",
		Email:    "sari@example.com",
		Password: "rahasia123",
	}
	if errs := ValidateRegister(valid); len(errs) != 0 {
		t.Errorf("request valid seharusnya tanpa error, dapat %v", errs)
	}

	// Username dengan karakter terlarang
	aneh := model.RegisterRequest{
		Username: "sari!!!",
		Email:    "sari@example.com",
		Password: "rahasia123",
	}
	if errs := ValidateRegister(aneh); errs["username"] == "" {
		t.Error("username dengan tanda seru seharusnya ditolak")
	}
}

func TestValidateLogin(t *testing.T) {
	if errs := ValidateLogin(model.LoginRequest{}); len(errs) != 2 {
		t.Errorf("login kosong: harap 2 error, dapat %d", len(errs))
	}

	ok := model.LoginRequest{Username: "sari", Password: "apapun"}
	if errs := ValidateLogin(ok); len(errs) != 0 {
		t.Errorf("login lengkap seharusnya tanpa error, dapat %v", errs)
	}
}
