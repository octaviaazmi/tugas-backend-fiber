package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"modul-5/app/model"
	"modul-5/app/repository"
	"modul-5/helper"
)

type AchievementService struct {
	repo repository.AchievementRepository
}

func NewAchievementService(repo repository.AchievementRepository) *AchievementService {
	return &AchievementService{repo: repo}
}

func (s *AchievementService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	achievement, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "prestasi tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data prestasi")
	}

	return helper.Success(c, fiber.StatusOK, "prestasi ditemukan", achievement)
}

func (s *AchievementService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.NamaPrestasi = strings.TrimSpace(req.NamaPrestasi)
	req.Juara = strings.TrimSpace(req.Juara)

	errs := map[string]string{}
	if req.IDMhs <= 0 {
		errs["id_mhs"] = "wajib diisi dan harus angka positif"
	}
	if req.NamaPrestasi == "" {
		errs["nama_prestasi"] = "wajib diisi"
	}
	if req.Juara == "" {
		errs["juara"] = "wajib diisi"
	}
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	newAchievement, err := s.repo.Create(ctx, model.Achievement{
		IDMhs:        req.IDMhs,
		NamaPrestasi: req.NamaPrestasi,
		Juara:        req.Juara,
	})
	if err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, err.Error())
	}

	return helper.Created(c, "prestasi berhasil dibuat", newAchievement, "")
}
