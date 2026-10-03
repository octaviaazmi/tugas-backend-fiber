package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"

	"siakad-mini-fiber/app/model"
	"siakad-mini-fiber/app/repository"
	"siakad-mini-fiber/helper"
)

var reTahunAkademik = regexp.MustCompile(`^\d{4}/\d{4}-(Ganjil|Genap)$`)

type EnrollmentService struct {
	enrolls  repository.EnrollmentRepository
	students repository.StudentRepository
	courses  repository.CourseRepository
}

func NewEnrollmentService(
	enrolls repository.EnrollmentRepository,
	students repository.StudentRepository,
	courses repository.CourseRepository,
) *EnrollmentService {
	return &EnrollmentService{enrolls: enrolls, students: students, courses: courses}
}

// POST /enrollments
func (s *EnrollmentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, _ := helper.CurrentUser(c)

	student, err := s.students.FindByUserID(ctx, current.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusForbidden, "Data mahasiswa tidak ditemukan")
	}

	var req model.CreateEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Body harus berupa JSON yang valid")
	}

	if req.CourseID < 1 {
		return helper.FailValidation(c, map[string]string{"course_id": "Wajib diisi"})
	}
	if !reTahunAkademik.MatchString(req.TahunAkademik) {
		return helper.FailValidation(c, map[string]string{
			"tahun_akademik": "Format harus YYYY/YYYY-Ganjil atau YYYY/YYYY-Genap",
		})
	}

	// Cek course
	course, err := s.courses.FindByID(ctx, req.CourseID)
	if err != nil {
		return helper.FailValidation(c, map[string]string{"course_id": "Mata kuliah tidak ditemukan"})
	}

	// Cek duplikasi
	exists, err := s.enrolls.Exists(ctx, student.ID, course.ID, req.TahunAkademik)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal memeriksa data")
	}
	if exists {
		return helper.Fail(c, fiber.StatusConflict,
			"Mata kuliah sudah pernah diambil pada tahun akademik ini")
	}

	// Cek kuota
	terisi, err := s.courses.CountTerisi(ctx, course.ID)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal memeriksa kuota")
	}
	if terisi >= course.Kuota {
		return helper.Fail(c, fiber.StatusUnprocessableEntity, "Kuota mata kuliah sudah penuh")
	}

	// Cek batas SKS
	batas := hitungBatasSKS(student.IPKTerakhir)
	totalSKS, err := s.enrolls.TotalSKS(ctx, student.ID, req.TahunAkademik)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal memeriksa SKS")
	}
	if totalSKS+course.SKS > batas {
		sisa := batas - totalSKS
		if sisa < 0 {
			sisa = 0
		}
		return helper.Fail(c, fiber.StatusUnprocessableEntity,
			fmt.Sprintf("Total SKS melebihi batas. Sisa SKS Anda: %d", sisa))
	}

	enrollment, err := s.enrolls.Create(ctx, model.Enrollment{
		StudentID:     student.ID,
		CourseID:      course.ID,
		TahunAkademik: req.TahunAkademik,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Fail(c, fiber.StatusConflict,
				"Mata kuliah sudah pernah diambil pada tahun akademik ini")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal menyimpan enrollment")
	}

	return helper.Created(c, "Berhasil mengambil mata kuliah", fiber.Map{
		"id":             enrollment.ID,
		"student_id":     enrollment.StudentID,
		"course_id":      enrollment.CourseID,
		"kode_mk":        course.KodeMK,
		"nama_mk":        course.NamaMK,
		"sks":            course.SKS,
		"tahun_akademik": enrollment.TahunAkademik,
	}, "/api/v1/enrollments/"+fmt.Sprint(enrollment.ID))
}

// DELETE /enrollments/:id
func (s *EnrollmentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, _ := helper.CurrentUser(c)

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusNotFound, "Enrollment tidak ditemukan")
	}

	enrollment, err := s.enrolls.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Enrollment tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal mengambil data")
	}

	student, err := s.students.FindByUserID(ctx, current.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusForbidden, "Data mahasiswa tidak ditemukan")
	}

	if enrollment.StudentID != student.ID {
		return helper.Fail(c, fiber.StatusForbidden, "Akses ditolak")
	}

	if err := s.enrolls.Delete(ctx, id); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal menghapus enrollment")
	}

	return helper.NoContent(c)
}

var _ = strings.TrimSpace
