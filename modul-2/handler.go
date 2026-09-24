package main

import (
	"sort"
	"strconv"
	"strings"

	"api-students/app/model"

	"github.com/gofiber/fiber/v2"
)

// Data sementara disimpan di memori (hilang kalau server mati)
var students []model.Student
var nextID = 1

// Fungsi pencari indeks mahasiswa berdasarkan ID
func findStudentIndex(id int) int {
	for i := range students {
		if students[i].ID == id {
			return i
		}
	}
	return -1
}

// Fungsi pencarian kata kunci pada Nama atau NIM
func cocokPencarian(s model.Student, kata string) bool {
	kata = strings.ToLower(kata)
	return strings.Contains(strings.ToLower(s.Name), kata) || strings.Contains(strings.ToLower(s.NIM), kata)
}

// Fungsi pembaca parameter ID dari URL
func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// ==========================================
// 1. GET (Daftar Mahasiswa + Paginasi & Filter)
// ==========================================
func listStudents(c *fiber.Ctx) error {
	q := parseListQuery(c)
	hasil := []model.Student{}

	// A. Saring (Filter is_active & Search name/nim)
	for _, s := range students {
		if q.IsActive != nil && s.IsActive != *q.IsActive {
			continue
		}
		if q.Search != "" && !cocokPencarian(s, q.Search) {
			continue
		}
		hasil = append(hasil, s)
	}

	// B. Urutkan (Sort)
	sort.SliceStable(hasil, func(i, j int) bool {
		var lebihKecil bool
		switch q.Sort {
		case "name":
			lebihKecil = strings.ToLower(hasil[i].Name) < strings.ToLower(hasil[j].Name)
		case "grade":
			lebihKecil = hasil[i].Grade < hasil[j].Grade
		default: // default urut berdasarkan id
			lebihKecil = hasil[i].ID < hasil[j].ID
		}
		if q.Order == "desc" {
			return !lebihKecil
		}
		return lebihKecil
	})

	// C. Potong sesuai halaman (Paginasi)
	total := len(hasil)
	totalPages := (total + q.Limit - 1) / q.Limit
	mulai := (q.Page - 1) * q.Limit
	if mulai > total {
		mulai = total
	}
	akhir := mulai + q.Limit
	if akhir > total {
		akhir = total
	}

	return okList(c, "daftar mahasiswa berhasil diambil", hasil[mulai:akhir], &model.Meta{
		Page: q.Page, Limit: q.Limit, Total: total, TotalPages: totalPages,
	})
}

// ==========================================
// 2. GET (Ambil 1 Mahasiswa by ID)
// ==========================================
func getStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "data mahasiswa tidak ditemukan")
	}
	return ok(c, "data mahasiswa ditemukan", students[i])
}

// ==========================================
// 3. POST (Tambah Data Baru)
// ==========================================
func createStudent(c *fiber.Ctx) error {
	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	req.Name = strings.TrimSpace(req.Name)
	req.NIM = strings.TrimSpace(req.NIM)

	if req.Name == "" {
		errs["name"] = "nama wajib diisi"
	}
	if req.NIM == "" {
		errs["nim"] = "NIM wajib diisi"
	}

	// Status 409 Conflict: Cek kalau NIM udah dipakai
	for _, s := range students {
		if strings.EqualFold(s.NIM, req.NIM) {
			return fail(c, fiber.StatusConflict, "NIM sudah terdaftar dalam sistem")
		}
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	baru := model.Student{
		ID:       nextID,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true, // Otomatis aktif saat dibuat
	}
	students = append(students, baru)
	nextID++

	// Status 201 Created & Header Location
	return created(c, "data mahasiswa berhasil ditambahkan", baru, "/api/v1/students/"+strconv.Itoa(baru.ID))
}

// ==========================================
// 4. PUT (Ganti Seluruh Data)
// ==========================================
func replaceStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "data mahasiswa tidak ditemukan")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi pada PUT"
	}
	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}
	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	// Status 409 Conflict: Cek kalau ngubah ke NIM yang dipakai mahasiswa lain
	for index, s := range students {
		if index != i && strings.EqualFold(s.NIM, req.NIM) {
			return fail(c, fiber.StatusConflict, "NIM sudah dipakai mahasiswa lain")
		}
	}

	students[i].NIM = req.NIM
	students[i].Name = req.Name
	students[i].Grade = req.Grade
	students[i].IsActive = req.IsActive

	return ok(c, "data mahasiswa berhasil diganti seluruhnya", students[i])
}

// ==========================================
// 5. PATCH (Ubah Data Sebagian)
// ==========================================
func patchStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "data mahasiswa tidak ditemukan")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if req.Name == nil && req.NIM == nil && req.Grade == nil && req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return failValidation(c, map[string]string{"name": "tidak boleh kosong"})
		}
		students[i].Name = *req.Name
	}

	if req.NIM != nil {
		if strings.TrimSpace(*req.NIM) == "" {
			return failValidation(c, map[string]string{"nim": "tidak boleh kosong"})
		}
		// Cek konflik NIM
		for index, s := range students {
			if index != i && strings.EqualFold(s.NIM, *req.NIM) {
				return fail(c, fiber.StatusConflict, "NIM sudah dipakai mahasiswa lain")
			}
		}
		students[i].NIM = *req.NIM
	}

	if req.Grade != nil {
		students[i].Grade = *req.Grade
	}

	if req.IsActive != nil {
		students[i].IsActive = *req.IsActive
	}

	return ok(c, "data mahasiswa berhasil diperbarui sebagian", students[i])
}

// ==========================================
// 6. DELETE (Hapus Data)
// ==========================================
func deleteStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "data mahasiswa tidak ditemukan")
	}

	// Hapus elemen dari slice (array)
	students = append(students[:i], students[i+1:]...)

	// Status 204: Berhasil, tapi tidak ada body respons yang dikirim
	return noContent(c)
}
