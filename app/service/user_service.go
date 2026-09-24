package service

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"modul6/app/model"
	"modul6/app/repository"
	"modul6/helper"
)

type UserService struct {
	repo  repository.UserRepository
	perms *helper.PermissionSet
}

func NewUserService(repo repository.UserRepository, perms *helper.PermissionSet) *UserService {
	return &UserService{repo: repo, perms: perms}
}

func (s *UserService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	users, err := s.repo.List(ctx)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil daftar user")
	}
	return helper.Success(c, fiber.StatusOK, "daftar user", users)
}

func (s *UserService) Get(c *fiber.Ctx) error {
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

	// Untuk users, ownership cukup dibandingkan dari token vs param (tanpa query
	// database), jadi hak akses bisa diperiksa SEBELUM data diambil.
	if !CanAccessUser(current, id, s.perms, "user:read:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengakses data user lain")
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return helper.Fail(c, fiber.StatusNotFound, "user tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data user")
	}
	return helper.Success(c, fiber.StatusOK, "user ditemukan", user)
}

func (s *UserService) Update(c *fiber.Ctx) error {
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

	if !CanAccessUser(current, id, s.perms, "user:update:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengubah data user lain")
	}

	var req model.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	if req.Username == "" || req.Email == "" {
		return helper.FailValidation(c, map[string]string{"username_email": "wajib diisi"})
	}

	user, err := s.repo.Update(ctx, id, req.Username, req.Email)
	if err != nil {
		if err == repository.ErrNotFound {
			return helper.Fail(c, fiber.StatusNotFound, "user tidak ditemukan")
		}
		if err == repository.ErrDuplicate {
			return helper.Fail(c, fiber.StatusConflict, "username atau email sudah dipakai")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengubah user")
	}
	return helper.Success(c, fiber.StatusOK, "user berhasil diubah", user)
}

func (s *UserService) Delete(c *fiber.Ctx) error {
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

	// Punya permission user:delete tidak berarti boleh menghapus diri sendiri.
	if current.UserID == id {
		return helper.Fail(c, fiber.StatusForbidden, "tidak boleh menghapus akun sendiri")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if err == repository.ErrNotFound {
			return helper.Fail(c, fiber.StatusNotFound, "user tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menghapus user")
	}
	return helper.NoContent(c)
}

func (s *UserService) AssignRole(c *fiber.Ctx) error {
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

	var req model.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if errs := ValidateAssignRole(current, id, req.Role, s.perms); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	user, err := s.repo.UpdateRole(ctx, id, strings.TrimSpace(req.Role))
	if err != nil {
		if err == repository.ErrNotFound {
			return helper.Fail(c, fiber.StatusNotFound, "user tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengubah role user")
	}
	return helper.Success(c, fiber.StatusOK, "role user berhasil diubah", user)
}
