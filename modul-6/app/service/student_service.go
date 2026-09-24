package service

import (
	"github.com/gofiber/fiber/v2"

	"modul6/app/model"
	"modul6/app/repository"
	"modul6/helper"
)

type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

func NewStudentService(repo repository.StudentRepository, perms *helper.PermissionSet) *StudentService {
	return &StudentService{repo: repo, perms: perms}
}

func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	students, err := s.repo.List(ctx)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar student")
	}
	return helper.Success(c, fiber.StatusOK, "daftar student", students)
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	if req.NIM == "" {
		errs["nim"] = "wajib diisi"
	}
	if req.Nama == "" {
		errs["nama"] = "wajib diisi"
	}
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	// owner_id diambil dari identitas token (current.UserID), BUKAN dari body.
	// Model CreateStudentRequest memang tidak punya field owner_id, jadi
	// mengirim {"owner_id": 1} di body tidak berpengaruh apa pun.
	student, err := s.repo.Create(ctx, req, current.UserID)
	if err != nil {
		if err == repository.ErrDuplicate {
			return helper.Fail(c, fiber.StatusConflict, "NIM sudah terdaftar")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat data student")
	}
	return helper.Success(c, fiber.StatusCreated, "student berhasil dibuat", student)
}

func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return helper.Fail(c, fiber.StatusNotFound, "student tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data student")
	}

	// Beda dengan users: di sini data HARUS diambil dulu untuk tahu owner_id-nya,
	// baru bisa diputuskan boleh atau tidak. Ini dibahas di jawaban C.4 nanti.
	if !CanAccessStudent(current, student.OwnerID, s.perms, "student:read:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengakses data student ini")
	}

	return helper.Success(c, fiber.StatusOK, "student ditemukan", student)
}

func (s *StudentService) Update(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return helper.Fail(c, fiber.StatusNotFound, "student tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data student")
	}

	if !CanAccessStudent(current, existing.OwnerID, s.perms, "student:update:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengubah data student ini")
	}

	var req model.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if req.Nama == "" {
		return helper.FailValidation(c, map[string]string{"nama": "wajib diisi"})
	}

	student, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengubah data student")
	}
	return helper.Success(c, fiber.StatusOK, "student berhasil diubah", student)
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if err == repository.ErrNotFound {
			return helper.Fail(c, fiber.StatusNotFound, "student tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menghapus data student")
	}
	return helper.NoContent(c)
}
