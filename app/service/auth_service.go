package service

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"modul6/app/model"
	"modul6/app/repository"
	"modul6/helper"
)

type AuthService struct {
	repo  repository.UserRepository
	jwt   *helper.JWTManager
	perms *helper.PermissionSet
}

func NewAuthService(repo repository.UserRepository, jwt *helper.JWTManager, perms *helper.PermissionSet) *AuthService {
	return &AuthService{repo: repo, jwt: jwt, perms: perms}
}

func (s *AuthService) Register(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	if req.Username == "" {
		errs["username"] = "wajib diisi"
	}
	if req.Email == "" {
		errs["email"] = "wajib diisi"
	}
	if len(req.Password) < 8 {
		errs["password"] = "minimal 8 karakter"
	}
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	hash, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memproses password")
	}

	user, err := s.repo.Create(ctx, req.Username, req.Email, hash)
	if err != nil {
		if err == repository.ErrDuplicate {
			return helper.Fail(c, fiber.StatusConflict, "username atau email sudah dipakai")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mendaftarkan user")
	}

	// role selalu 'user' saat register; menaikkan role dilakukan admin lewat
	// PATCH /users/:id/role, bukan lewat body register (mencegah mass assignment).
	return helper.Success(c, fiber.StatusCreated, "registrasi berhasil", user)
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	user, hash, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "username atau password salah")
	}
	if !helper.CheckPassword(hash, req.Password) {
		return helper.Fail(c, fiber.StatusUnauthorized, "username atau password salah")
	}

	token, err := s.jwt.Generate(user.ID, user.Role)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat token")
	}

	return helper.Success(c, fiber.StatusOK, "login berhasil", fiber.Map{
		"access_token": token,
		"user":         user,
	})
}

func (s *AuthService) Me(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	user, err := s.repo.FindByID(ctx, current.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusNotFound, "user tidak ditemukan")
	}

	return helper.Success(c, fiber.StatusOK, "profil berhasil diambil", fiber.Map{
		"user":        user,
		"permissions": s.perms.PermissionsOf(user.Role),
	})
}
