package service

import (
	"testing"

	"modul-4/app/model"
)

// Uji fungsi POST (Create)
func TestValidateCreate(t *testing.T) {
	// Sengaja kita kasih data kosong dan nilai minus biar error
	req := model.CreateStudentRequest{NIM: "", Name: "", Grade: -10}
	errs := ValidateCreate(req)

	if len(errs) != 3 {
		t.Errorf("Harusnya ada 3 error (NIM, Name, Grade), tapi malah dapat %d", len(errs))
	}
}

// Uji fungsi PUT (Replace)
func TestValidateReplace(t *testing.T) {
	// Sengaja kita kasih data yang benar semua
	req := model.ReplaceStudentRequest{NIM: "111", Name: "Pia", Grade: 90, IsActive: true}
	errs := ValidateReplace(req)

	if len(errs) != 0 {
		t.Errorf("Data valid harusnya tidak error, tapi malah dapat %v", errs)
	}
}

// Uji fungsi PATCH (ApplyPatch)
func TestApplyPatch(t *testing.T) {
	// Data awal mahasiswa
	initial := model.Student{ID: 1, NIM: "123", Name: "Pia", Grade: 95, IsActive: true}

	// Kita cuma mau ngubah status IsActive jadi false
	inactive := false
	result, errs := ApplyPatch(initial, model.PatchStudentRequest{IsActive: &inactive})

	if len(errs) != 0 {
		t.Fatalf("Tidak seharusnya ada error, tapi malah dapat: %v", errs)
	}
	if result.IsActive != false {
		t.Error("is_active seharusnya berubah menjadi false")
	}
	if result.Name != "Pia" {
		t.Error("Field name yang tidak diubah seharusnya tetap 'Pia'")
	}
}
