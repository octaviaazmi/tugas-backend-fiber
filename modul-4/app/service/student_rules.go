package service

import (
	"strings"

	"modul-4/app/model"
)

// ValidateCreate memeriksa kelengkapan data saat menambah mahasiswa baru (POST).
func ValidateCreate(req model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi"
	}
	if req.Grade < 0 {
		errs["grade"] = "nilai tidak boleh negatif"
	}
	return errs
}

// ValidateReplace memeriksa kelengkapan data saat memperbarui seluruh data mahasiswa (PUT).
func ValidateReplace(req model.ReplaceStudentRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi pada PUT"
	}
	if req.Grade < 0 {
		errs["grade"] = "nilai tidak boleh negatif"
	}
	return errs
}

// ApplyPatch menyalin field yang dikirim ke data mahasiswa yang sudah ada (PATCH).
func ApplyPatch(current model.Student, req model.PatchStudentRequest) (model.Student, map[string]string) {
	errs := map[string]string{}

	if req.NIM != nil {
		if strings.TrimSpace(*req.NIM) == "" {
			errs["nim"] = "tidak boleh kosong"
		} else {
			current.NIM = *req.NIM
		}
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			errs["name"] = "tidak boleh kosong"
		} else {
			current.Name = *req.Name
		}
	}
	if req.Grade != nil {
		if *req.Grade < 0 {
			errs["grade"] = "nilai tidak boleh negatif"
		} else {
			current.Grade = *req.Grade
		}
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current, errs
}

// IsEmptyPatch mendeteksi jika permintaan PATCH tidak mengirim perubahan apa-apa.
func IsEmptyPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}

// CountTotalPages menghitung jumlah halaman untuk pagination.
func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
}
