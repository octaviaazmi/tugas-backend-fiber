package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini-fiber/app/service"
	"siakad-mini-fiber/helper"
	"siakad-mini-fiber/middleware"
)

type Dependencies struct {
	Pool              *pgxpool.Pool
	JWT               *helper.JWTManager
	AuthService       *service.AuthService
	StudentService    *service.StudentService
	CourseService     *service.CourseService
	EnrollmentService *service.EnrollmentService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// ---- publik ----
	api.Get("/health", healthCheck(deps.Pool))

	// ---- authentication ----
	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/login", deps.AuthService.Login)

	authed := api.Group("/auth", middleware.RequireAuth(deps.JWT))
	authed.Get("/me", deps.AuthService.Me)

	// ---- wajib login ----
	students := api.Group("/students", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	students.Get("/:id", deps.StudentService.Detail)

	courses := api.Group("/courses", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))
	courses.Get("/", deps.CourseService.List)

	// ---- admin only ----
	admin := api.Group("/students",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT),
		middleware.RequireRole("admin"),
	)
	admin.Get("/", deps.StudentService.List)
	admin.Post("/", deps.StudentService.Create)
	admin.Put("/:id", deps.StudentService.Update)
	admin.Delete("/:id", deps.StudentService.Delete)

	// ---- mahasiswa only ----
	mhs := api.Group("/enrollments",
		middleware.RequireJSON,
		middleware.RequireAuth(deps.JWT),
		middleware.RequireRole("mahasiswa"),
	)
	mhs.Post("/", deps.EnrollmentService.Create)
	mhs.Delete("/:id", deps.EnrollmentService.Delete)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable, "Database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "Server dan database berjalan", nil)
	}
}
