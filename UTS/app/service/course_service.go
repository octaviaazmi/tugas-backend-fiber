package service

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini-fiber/app/repository"
	"siakad-mini-fiber/helper"
)

type CourseService struct {
	courses repository.CourseRepository
}

func NewCourseService(courses repository.CourseRepository) *CourseService {
	return &CourseService{courses: courses}
}

// GET /courses
func (s *CourseService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := helper.ParseCourseListQuery(c)

	rows, err := s.courses.FindAll(ctx, q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal mengambil data mata kuliah")
	}

	return helper.Success(c, fiber.StatusOK, "Data mata kuliah berhasil diambil", rows)
}
