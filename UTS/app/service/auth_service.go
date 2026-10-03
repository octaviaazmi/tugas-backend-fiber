package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"siakad-mini-fiber/app/model"
	"siakad-mini-fiber/app/repository"
	"siakad-mini-fiber/helper"
)

type AuthService struct {
	users    repository.UserRepository
	students repository.StudentRepository
	jwt      *helper.JWTManager
}

func NewAuthService(
	users repository.UserRepository,
	students repository.StudentRepository,
	jwtManager *helper.JWTManager,
) *AuthService {
	return &AuthService{users: users, students: students, jwt: jwtManager}
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	if strings.TrimSpace(req.Email) == "" {
		errs["email"] = "Wajib diisi"
	}
	if len(req.Password) < 8 {
		errs["password"] = "Minimal 8 karakter"
	}
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	user, err := s.users.FindByEmail(ctx, strings.TrimSpace(req.Email))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			helper.VerifyDummyPassword(req.Password)
			return helper.Fail(c, fiber.StatusUnauthorized, "Email atau password salah")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal memproses login")
	}

	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.Fail(c, fiber.StatusUnauthorized, "Email atau password salah")
	}

	accessToken, err := s.jwt.GenerateAccess(user)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal membuat token")
	}

	return helper.Success(c, fiber.StatusOK, "Login berhasil", fiber.Map{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   int(s.jwt.AccessTTL().Seconds()),
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func (s *AuthService) Me(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "Belum terautentikasi")
	}

	user, err := s.users.FindByID(ctx, authUser.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "User tidak ditemukan")
	}

	resp := fiber.Map{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	}

	if user.Role == "mahasiswa" {
		student, err := s.students.FindByUserID(ctx, user.ID)
		if err == nil {
			resp["student"] = fiber.Map{
				"nim":      student.NIM,
				"nama":     student.Nama,
				"prodi":    student.Prodi,
				"angkatan": student.Angkatan,
			}
		}
	}

	return helper.Success(c, fiber.StatusOK, "Profil berhasil diambil", resp)
}
