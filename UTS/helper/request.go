package helper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"siakad-mini-fiber/app/model"
)

func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

func ParamID(c *fiber.Ctx) (int64, bool) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

var allowedStudentSort = map[string]bool{
	"nama":          true,
	"-ipk_terakhir": true,
	"id":            true,
}

func ParseStudentListQuery(c *fiber.Ctx) model.StudentListQuery {
	q := model.StudentListQuery{
		Page:    c.QueryInt("page", 1),
		PerPage: c.QueryInt("per_page", 10),
		Prodi:   strings.TrimSpace(c.Query("prodi")),
		Search:  strings.TrimSpace(c.Query("search")),
		Sort:    c.Query("sort", "id"),
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 {
		q.PerPage = 10
	}
	if q.PerPage > 50 {
		q.PerPage = 50
	}

	if angkatan := c.Query("angkatan"); angkatan != "" {
		if v, err := strconv.Atoi(angkatan); err == nil {
			q.Angkatan = v
		}
	}

	if !allowedStudentSort[q.Sort] {
		q.Sort = "id"
	}

	return q
}

func ParseCourseListQuery(c *fiber.Ctx) model.CourseListQuery {
	q := model.CourseListQuery{
		Search: strings.TrimSpace(c.Query("search")),
	}

	if semester := c.Query("semester"); semester != "" {
		if v, err := strconv.Atoi(semester); err == nil {
			q.Semester = v
		}
	}

	if av := c.Query("available"); av == "true" {
		q.Available = true
	}

	return q
}
