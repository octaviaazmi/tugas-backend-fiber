package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/bson"

	"modul-10/app/model"
	"modul-10/app/repository"
	"modul-10/helper"
)

type PortfolioService struct {
	pool       *pgxpool.Pool
	portfolios repository.PortfolioRepository
}

func NewPortfolioService(pool *pgxpool.Pool, portfolios repository.PortfolioRepository) *PortfolioService {
	return &PortfolioService{pool: pool, portfolios: portfolios}
}

// Fungsi bantuan untuk cek apakah student_id ada di PostgreSQL
func (s *PortfolioService) checkStudentExists(ctx context.Context, studentID int) error {
	var exists bool
	err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM students WHERE id = $1 AND deleted_at IS NULL)", studentID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("gagal memeriksa mahasiswa: %w", err)
	}
	if !exists {
		return errors.New("mahasiswa tidak ditemukan")
	}
	return nil
}

// GET /students/:id/portfolio
func (s *PortfolioService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id mahasiswa tidak valid")
	}

	if err := s.checkStudentExists(ctx, int(studentID)); err != nil {
		return helper.Fail(c, fiber.StatusNotFound, err.Error())
	}

	p, err := s.portfolios.FindByStudentID(ctx, int(studentID))
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "portofolio belum dibuat")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil portofolio")
	}

	return helper.Success(c, fiber.StatusOK, "portofolio berhasil diambil", p)
}

// PUT /students/:id/portfolio
func (s *PortfolioService) Upsert(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id mahasiswa tidak valid")
	}

	var req model.UpsertPortfolioRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.FailValidation(c, errs)
	}

	if err := s.checkStudentExists(ctx, int(studentID)); err != nil {
		return helper.Fail(c, fiber.StatusNotFound, err.Error())
	}

	p, err := s.portfolios.Upsert(ctx, model.Portfolio{
		StudentID:    int(studentID),
		Headline:     req.Headline,
		Bio:          req.Bio,
		Certificates: req.Certificates,
	})
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan portofolio")
	}

	return helper.Success(c, fiber.StatusOK, "portofolio berhasil disimpan", p)
}

// POST /students/:id/portfolio/projects
func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id mahasiswa tidak valid")
	}

	var req model.AddProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.FailValidation(c, errs)
	}

	p, err := s.portfolios.AddProject(ctx, int(studentID), model.Project{
		ProjectID:    bson.NewObjectID().Hex(), // Generate ID unik untuk proyek
		Title:        req.Title,
		Technologies: req.Technologies,
		Link:         req.Link,
		Year:         req.Year,
	})
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "buat portofolio terlebih dahulu")
		}
		if errors.Is(err, repository.ErrProjectLimit) {
			return helper.Fail(c, fiber.StatusConflict, err.Error())
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menambah proyek")
	}

	return helper.Created(c, "proyek berhasil ditambahkan", p, "")
}

// DELETE /students/:id/portfolio/projects/:projectId
func (s *PortfolioService) DeleteProject(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	studentID, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id mahasiswa tidak valid")
	}

	projectID := c.Params("projectId")
	if projectID == "" {
		return helper.Fail(c, fiber.StatusBadRequest, "id proyek tidak valid")
	}

	err := s.portfolios.DeleteProject(ctx, int(studentID), projectID)
	if err != nil {
		if errors.Is(err, repository.ErrProjectNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "proyek tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menghapus proyek")
	}

	return helper.NoContent(c)
}

// GET /portfolios/stats/technologies
func (s *PortfolioService) Stats(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	stats, err := s.portfolios.StatsTechnologies(ctx)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil statistik")
	}

	return helper.Success(c, fiber.StatusOK, "statistik teknologi berhasil diambil", stats)
}
