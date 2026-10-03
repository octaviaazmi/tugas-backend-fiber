package service

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"siakad-mini-fiber/app/model"
	"siakad-mini-fiber/app/repository"
	"siakad-mini-fiber/helper"
)

var (
	reNIM   = regexp.MustCompile(`^\d{12}$`)
	reEmail = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
)

func hitungBatasSKS(ipk float64) int {
	switch {
	case ipk >= 3.00:
		return 24
	case ipk >= 2.50:
		return 21
	default:
		return 18
	}
}

type StudentService struct {
	students repository.StudentRepository
	users    repository.UserRepository
	courses  repository.CourseRepository
	enrolls  repository.EnrollmentRepository
}

func NewStudentService(
	students repository.StudentRepository,
	users repository.UserRepository,
	courses repository.CourseRepository,
	enrolls repository.EnrollmentRepository,
) *StudentService {
	return &StudentService{students: students, users: users, courses: courses, enrolls: enrolls}
}

// GET /students
func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := helper.ParseStudentListQuery(c)

	rows, total, err := s.students.FindAll(ctx, q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal mengambil data mahasiswa")
	}

	lastPage := 0
	if q.PerPage > 0 {
		lastPage = int((total + int64(q.PerPage) - 1) / int64(q.PerPage))
	}

	return helper.SuccessList(c, "Data mahasiswa berhasil diambil", rows, &model.Meta{
		CurrentPage: q.Page,
		PerPage:     q.PerPage,
		Total:       total,
		LastPage:    lastPage,
	})
}

// POST /students
func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Body harus berupa JSON yang valid")
	}

	errs := map[string]string{}

	if !reNIM.MatchString(req.NIM) {
		errs["nim"] = "NIM harus 12 digit angka"
	}
	if strings.TrimSpace(req.Nama) == "" {
		errs["nama"] = "Wajib diisi"
	}
	if !reEmail.MatchString(req.Email) {
		errs["email"] = "Format email tidak valid"
	}
	if strings.TrimSpace(req.Prodi) == "" {
		errs["prodi"] = "Wajib diisi"
	}
	currentYear := time.Now().Year()
	if req.Angkatan < 1900 || req.Angkatan > currentYear {
		errs["angkatan"] = "Angkatan harus 4 digit dan tidak melebihi tahun berjalan"
	}
	if req.IPKTerakhir < 0 || req.IPKTerakhir > 4 {
		errs["ipk_terakhir"] = "IPK harus antara 0.00 dan 4.00"
	}
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	hash, err := helper.HashPassword(req.NIM)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal memproses password")
	}

	// Create user
	user, err := s.users.Create(ctx, model.User{
		Email:    req.Email,
		Password: hash,
		Role:     "mahasiswa",
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.FailValidation(c, map[string]string{"email": "Email sudah terdaftar"})
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal menyimpan user")
	}

	// Create student
	student, err := s.students.Create(ctx, model.Student{
		UserID:      user.ID,
		NIM:         req.NIM,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IPKTerakhir: req.IPKTerakhir,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.FailValidation(c, map[string]string{"nim": "NIM sudah terdaftar"})
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal menyimpan mahasiswa")
	}

	return helper.Created(c, "Mahasiswa berhasil ditambahkan", fiber.Map{
		"id":           student.ID,
		"nim":          student.NIM,
		"nama":         student.Nama,
		"prodi":        student.Prodi,
		"angkatan":     student.Angkatan,
		"ipk_terakhir": student.IPKTerakhir,
		"email":        user.Email,
	}, "/api/v1/students/"+strconv.FormatInt(student.ID, 10))
}

// GET /students/:id
func (s *StudentService) Detail(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, _ := helper.CurrentUser(c)

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}

	student, err := s.students.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal mengambil data")
	}

	// Cek ownership
	if current.Role == "mahasiswa" && student.UserID != current.UserID {
		return helper.Fail(c, fiber.StatusForbidden, "Akses ditolak")
	}

	// Ambil MK yang diambil
	enrollments, err := s.enrolls.FindByStudent(ctx, student.ID)
	if err != nil {
		enrollments = nil
	}

	totalSKS := 0
	for _, e := range enrollments {
		totalSKS += e.SKS
	}

	return helper.Success(c, fiber.StatusOK, "Detail mahasiswa berhasil diambil", fiber.Map{
		"id":           student.ID,
		"nim":          student.NIM,
		"nama":         student.Nama,
		"prodi":        student.Prodi,
		"angkatan":     student.Angkatan,
		"ipk_terakhir": student.IPKTerakhir,
		"mata_kuliah":  enrollments,
		"total_sks":    totalSKS,
		"batas_sks":    hitungBatasSKS(student.IPKTerakhir),
	})
}

// PUT /students/:id
func (s *StudentService) Update(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}

	student, err := s.students.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal mengambil data")
	}

	var req model.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	if req.Angkatan != nil {
		currentYear := time.Now().Year()
		if *req.Angkatan < 1900 || *req.Angkatan > currentYear {
			errs["angkatan"] = "Angkatan harus 4 digit dan tidak melebihi tahun berjalan"
		}
	}
	if req.IPKTerakhir != nil {
		if *req.IPKTerakhir < 0 || *req.IPKTerakhir > 4 {
			errs["ipk_terakhir"] = "IPK harus antara 0.00 dan 4.00"
		}
	}
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	if req.Nama != "" {
		student.Nama = req.Nama
	}
	if req.Prodi != "" {
		student.Prodi = req.Prodi
	}
	if req.Angkatan != nil {
		student.Angkatan = *req.Angkatan
	}
	if req.IPKTerakhir != nil {
		student.IPKTerakhir = *req.IPKTerakhir
	}

	updated, err := s.students.Update(ctx, student)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal memperbarui data")
	}

	return helper.Success(c, fiber.StatusOK, "Data mahasiswa berhasil diperbarui", updated)
}

// DELETE /students/:id
func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
	}

	if err := s.students.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Mahasiswa tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal menghapus data")
	}

	return helper.NoContent(c)
}
